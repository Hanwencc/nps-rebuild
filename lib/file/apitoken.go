package file

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ApiToken is a scoped, revocable credential used for machine-to-machine
// access to /api/v1/*. It supersedes the single shared API key from
// nps.conf; the legacy key still works but its use is discouraged.
//
// On the wire callers send:
//
//	X-Api-Key:    <KeyId>
//	X-Api-Secret: <plaintext secret>
//
// or equivalently `Authorization: Bearer <KeyId>.<secret>`.
//
// SecretHash is a bcrypt hash; the plaintext secret is shown to the
// admin exactly once (at creation) and never persisted.
type ApiToken struct {
	Id                 int      `json:"id"`
	KeyId              string   `json:"keyId"`
	SecretHash         string   `json:"secretHash"`
	Remark             string   `json:"remark"`
	AllowedPathPrefix  string   `json:"allowedPathPrefix"` // "" = any path under /api/v1
	AllowedMethods     []string `json:"allowedMethods"`    // empty = any
	AllowIps           []string `json:"allowIps"`          // exact IP or CIDR; empty = any
	ExpiresAt          int64    `json:"expiresAt"`         // unix seconds; 0 = never
	CreatedAt          int64    `json:"createdAt"`
	LastUsedAt         int64    `json:"lastUsedAt"`
	LastUsedIp         string   `json:"lastUsedIp"`
	Disabled           bool     `json:"disabled"`

	mu sync.Mutex `json:"-"`
}

// MatchesRequest returns nil if the token is valid for the given
// request context, otherwise a descriptive error.
func (t *ApiToken) MatchesRequest(method, path, ip string) error {
	if t.Disabled {
		return errors.New("token disabled")
	}
	if t.ExpiresAt > 0 && time.Now().Unix() > t.ExpiresAt {
		return errors.New("token expired")
	}
	if t.AllowedPathPrefix != "" && !strings.HasPrefix(path, t.AllowedPathPrefix) {
		return errors.New("path not allowed")
	}
	if len(t.AllowedMethods) > 0 {
		ok := false
		mu := strings.ToUpper(method)
		for _, m := range t.AllowedMethods {
			if strings.ToUpper(strings.TrimSpace(m)) == mu {
				ok = true
				break
			}
		}
		if !ok {
			return errors.New("method not allowed")
		}
	}
	if len(t.AllowIps) > 0 && !ipMatchesAny(ip, t.AllowIps) {
		return errors.New("ip not allowed")
	}
	return nil
}

// VerifySecret returns true when the supplied plaintext secret matches
// the token's stored bcrypt hash.
func (t *ApiToken) VerifySecret(plaintext string) bool {
	if plaintext == "" || t.SecretHash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(t.SecretHash), []byte(plaintext)) == nil
}

// Touch updates LastUsedAt/LastUsedIp at most once per second to avoid
// thrashing the JSON file on hot endpoints.
func (t *ApiToken) Touch(ip string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now().Unix()
	if now-t.LastUsedAt < 1 && t.LastUsedIp == ip {
		return false
	}
	t.LastUsedAt = now
	t.LastUsedIp = ip
	return true
}

func ipMatchesAny(ip string, allow []string) bool {
	if ip == "" {
		return false
	}
	parsed := net.ParseIP(ip)
	for _, raw := range allow {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		if strings.Contains(s, "/") {
			_, n, err := net.ParseCIDR(s)
			if err == nil && parsed != nil && n.Contains(parsed) {
				return true
			}
			continue
		}
		if s == ip {
			return true
		}
	}
	return false
}

// GenerateApiTokenSecret returns a 32-byte hex-encoded random secret
// suitable for storage as the plaintext component of an API token.
func GenerateApiTokenSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateApiTokenKeyId returns an 8-byte hex KeyId.
func GenerateApiTokenKeyId() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "k_" + hex.EncodeToString(b), nil
}

// HashApiTokenSecret applies bcrypt with the default cost.
func HashApiTokenSecret(plaintext string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// ----- DbUtils helpers ----------------------------------------------------

func (s *DbUtils) NewApiToken(t *ApiToken) {
	if t.Id == 0 {
		t.Id = int(atomic.AddInt32(&s.JsonDb.ApiTokenIncreaseId, 1))
	}
	if t.CreatedAt == 0 {
		t.CreatedAt = time.Now().Unix()
	}
	s.JsonDb.ApiTokens.Store(t.Id, t)
}

func (s *DbUtils) UpdateApiToken(t *ApiToken) {
	s.JsonDb.ApiTokens.Store(t.Id, t)
}

func (s *DbUtils) DelApiToken(id int) {
	s.JsonDb.ApiTokens.Delete(id)
}

func (s *DbUtils) GetApiToken(id int) (*ApiToken, error) {
	if v, ok := s.JsonDb.ApiTokens.Load(id); ok {
		return v.(*ApiToken), nil
	}
	return nil, errors.New("api token not found")
}

// FindApiTokenByKeyId scans the in-memory map; the count is small
// (admin-managed credentials, expected dozens at most) so a linear
// scan is acceptable.
func (s *DbUtils) FindApiTokenByKeyId(keyId string) (*ApiToken, error) {
	if keyId == "" {
		return nil, errors.New("empty keyId")
	}
	var found *ApiToken
	s.JsonDb.ApiTokens.Range(func(_, value interface{}) bool {
		t := value.(*ApiToken)
		if t.KeyId == keyId {
			found = t
			return false
		}
		return true
	})
	if found == nil {
		return nil, errors.New("api token not found")
	}
	return found, nil
}

// ListApiTokens returns a stable-ordered slice (by Id ascending) for UI.
func (s *DbUtils) ListApiTokens() []*ApiToken {
	out := make([]*ApiToken, 0)
	s.JsonDb.ApiTokens.Range(func(_, value interface{}) bool {
		out = append(out, value.(*ApiToken))
		return true
	})
	// simple insertion sort by Id
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1].Id > out[j].Id; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}
