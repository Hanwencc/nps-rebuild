package crypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
)

var (
	cert        tls.Certificate
	certFpRaw   []byte // SHA-256 of leaf DER, empty until cert loaded
)

// InitTls keeps the historical "generate ephemeral self-signed cert in
// memory only" behaviour. Use LoadOrInitServerCert instead when you
// want the cert persisted so NPC fingerprint pinning is stable across
// restarts.
func InitTls() {
	c, k, err := generateKeyPair("NPS Org")
	if err == nil {
		cert, err = tls.X509KeyPair(c, k)
	}
	if err != nil {
		log.Fatalln("Error initializing crypto certs", err)
	}
	updateCertFingerprint()
}

// LoadOrInitServerCert wires the bridge TLS certificate so its
// fingerprint stays stable across restarts. Behaviour:
//
//   - if both certPath and keyPath exist and parse, load them;
//   - otherwise generate a fresh 10y self-signed RSA-2048 keypair,
//     write both PEM files (key with 0600 perm), and use them.
//
// Either way the public certFpRaw is updated so GetCertFingerprintHex
// returns the canonical pin value.
func LoadOrInitServerCert(certPath, keyPath string) error {
	if certPath != "" && keyPath != "" && fileExists(certPath) && fileExists(keyPath) {
		c, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return fmt.Errorf("load tls cert from %s/%s: %w", certPath, keyPath, err)
		}
		cert = c
		updateCertFingerprint()
		return nil
	}
	// No cert on disk yet — generate, persist, install.
	rawCert, rawKey, err := generateKeyPair("NPS Org")
	if err != nil {
		return fmt.Errorf("generate tls cert: %w", err)
	}
	if certPath != "" {
		if err := writeFileSecure(certPath, rawCert, 0644); err != nil {
			return fmt.Errorf("write %s: %w", certPath, err)
		}
	}
	if keyPath != "" {
		// Private key gets 0600 — only the nps process should read it.
		if err := writeFileSecure(keyPath, rawKey, 0600); err != nil {
			return fmt.Errorf("write %s: %w", keyPath, err)
		}
	}
	c, err := tls.X509KeyPair(rawCert, rawKey)
	if err != nil {
		return fmt.Errorf("install generated tls cert: %w", err)
	}
	cert = c
	updateCertFingerprint()
	return nil
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func writeFileSecure(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(dirOf(path), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, perm)
}

func dirOf(path string) string {
	if i := strings.LastIndexAny(path, `/\`); i >= 0 {
		return path[:i]
	}
	return "."
}

// updateCertFingerprint must be called whenever the in-memory `cert`
// is replaced. Caches sha256(leaf DER) so callers don't pay the hash
// per request.
func updateCertFingerprint() {
	if len(cert.Certificate) == 0 {
		certFpRaw = nil
		return
	}
	sum := sha256.Sum256(cert.Certificate[0])
	certFpRaw = sum[:]
}

// GetCertFingerprint returns the SHA-256 digest of the loaded server
// certificate's DER encoding (32 bytes). Returns nil when no cert is
// loaded.
func GetCertFingerprint() []byte {
	out := make([]byte, len(certFpRaw))
	copy(out, certFpRaw)
	return out
}

// GetCertFingerprintHex returns the fingerprint as lowercase hex with
// colon separators every byte (eg. "7c:8e:93:...:af"). Returns "" when
// no cert is loaded.
func GetCertFingerprintHex() string {
	if len(certFpRaw) == 0 {
		return ""
	}
	parts := make([]string, len(certFpRaw))
	for i, b := range certFpRaw {
		parts[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(parts, ":")
}

// ParseFingerprint normalises a user-supplied pin string to the raw
// 32-byte SHA-256 digest. Accepts:
//
//   - bare hex                       7c8e93...af
//   - colon-separated hex            7c:8e:93:...:af
//   - space-separated hex            7c 8e 93 ... af
//   - any combination of upper/lower case
//   - optional "sha256:" / "SHA-256:" prefix
//
// Returns an error if the result is not exactly 32 bytes.
func ParseFingerprint(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty fingerprint")
	}
	for _, p := range []string{"sha256:", "sha-256:", "SHA256:", "SHA-256:"} {
		if strings.HasPrefix(s, p) {
			s = s[len(p):]
			break
		}
	}
	s = strings.Map(func(r rune) rune {
		switch r {
		case ':', '-', ' ', '\t', '\n', '\r':
			return -1
		}
		return r
	}, s)
	raw, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("decode fingerprint hex: %w", err)
	}
	if len(raw) != sha256.Size {
		return nil, fmt.Errorf("fingerprint must be %d bytes (sha256), got %d", sha256.Size, len(raw))
	}
	return raw, nil
}

func GetCert() tls.Certificate {
	return cert
}

func NewTlsServerConn(conn net.Conn) net.Conn {
	var err error
	if err != nil {
		logs.Error(err)
		os.Exit(0)
		return nil
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	return tls.Server(conn, config)
}

// NewTlsClientConn keeps the legacy "skip all verification" behaviour
// for backward compatibility with callers that aren't pinning yet. New
// code should prefer NewTlsClientConnPin.
func NewTlsClientConn(conn net.Conn) net.Conn {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	return tls.Client(conn, conf)
}

// NewTlsClientConnPin wraps `conn` in a TLS client that **rejects** the
// handshake unless the leaf certificate's SHA-256 fingerprint matches
// `expectedFP` (32 bytes). This is the MITM-resistant path: the system
// CA bundle is irrelevant, only the pinned fingerprint matters.
//
// Pass nil/empty `expectedFP` to fall back to the legacy unsafe
// behaviour — handy when the operator hasn't configured a pin yet but
// still wants TLS-encrypted (unauthenticated) transport. Caller is
// expected to log a clear warning in that case.
func NewTlsClientConnPin(conn net.Conn, expectedFP []byte) net.Conn {
	if len(expectedFP) == 0 {
		return NewTlsClientConn(conn)
	}
	pinned := make([]byte, len(expectedFP))
	copy(pinned, expectedFP)
	conf := &tls.Config{
		// We *intentionally* disable hostname / chain verification —
		// the bridge cert is a self-signed never-rotated keypair and
		// is identified solely by its fingerprint. A standard CA chain
		// would add no security here.
		InsecureSkipVerify: true,
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return errors.New("tls pin: server presented no certificate")
			}
			got := sha256.Sum256(rawCerts[0])
			if subtle.ConstantTimeCompare(got[:], pinned) != 1 {
				return fmt.Errorf("tls pin: server cert sha256 mismatch (got %x)", got[:])
			}
			return nil
		},
	}
	return tls.Client(conn, conf)
}

func generateKeyPair(CommonName string) (rawCert, rawKey []byte, err error) {
	// Create private key and self-signed certificate
	// Adapted from https://golang.org/src/crypto/tls/generate_cert.go

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}
	validFor := time.Hour * 24 * 365 * 10 // ten years
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"My Company Name LTD."},
			CommonName:   CommonName,
			Country:      []string{"US"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return
	}

	rawCert = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	rawKey = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return
}
