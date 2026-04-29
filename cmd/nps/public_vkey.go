package main

import (
	"crypto/rand"
	"strings"

	"ehang.io/nps/lib/file/sqlitedb"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

// legacySettingKeys are app_settings rows that nothing in the codebase
// reads any more. They were either backfilled from an old nps.conf or
// added before a feature was retired; leaving them in the DB causes
// the Web UI's "其他" group to expose stale, no-op fields. Scrubbed on
// every boot so a single restart cleans up upgraded installs.
var legacySettingKeys = []string{
	"auth_key",        // removed: legacy md5(key+ts) HTTP API auth
	"auth_crypt_key",  // removed: AES wrapper around auth_key
}

// pruneLegacySettings deletes the rows listed in legacySettingKeys
// from app_settings and clears the corresponding beego.AppConfig
// entries so the running process does not keep them in memory either.
// Failures are logged but never fatal — at worst the operator sees the
// fields one more boot until the issue is fixed.
func pruneLegacySettings(store *sqlitedb.Store) {
	if store == nil {
		return
	}
	n, err := store.DeleteSettings(legacySettingKeys...)
	if err != nil {
		logs.Warn("legacy settings prune: %v", err)
		return
	}
	for _, k := range legacySettingKeys {
		_ = beego.AppConfig.Set(k, "")
	}
	if n > 0 {
		logs.Info("legacy settings prune: removed %d obsolete row(s) (%s)",
			n, strings.Join(legacySettingKeys, ", "))
	}
}

// publicVkeyAlphabet is restricted to lowercase letters + digits to
// keep the value safe in URLs, command lines and shell-quoted configs.
const publicVkeyAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// publicVkeyMinLen is also enforced by server.InitFromCsv as the
// gating threshold for actually publishing the public client.
const publicVkeyMinLen = 16

// ensurePublicVkey hardens the optional "public client" feature.
//
// On every boot we look at app_settings.public_vkey (already loaded
// into beego.AppConfig by BackfillOrLoadAppSettings). If the value is
// missing, equal to a known weak default, or shorter than 16 chars we
// generate a fresh 20-character random vkey (lowercase letters +
// digits, 103 bits of entropy from crypto/rand) and persist it via
// the regular UpsertSetting path. The cleartext value is logged
// exactly once at startup so the operator can copy it into the NPC.
//
// If SQLite is unavailable we still refresh the in-memory AppConfig
// so the server.InitFromCsv length gate behaves consistently, but
// the value is not persisted across restarts in that degraded mode.
func ensurePublicVkey(store *sqlitedb.Store) {
	current := strings.TrimSpace(beego.AppConfig.String("public_vkey"))
	if !isWeakPublicVkey(current) {
		return
	}
	newVkey, err := generatePublicVkey(20)
	if err != nil {
		logs.Error("public_vkey: failed to generate random value: %v (公共客户端将保持禁用)", err)
		// Force the in-memory value to empty so InitFromCsv's len gate
		// definitely refuses to publish the weak default.
		_ = beego.AppConfig.Set("public_vkey", "")
		return
	}
	if store != nil {
		if err := store.UpsertSetting("public_vkey", newVkey); err != nil {
			logs.Error("public_vkey: persist failed: %v (本次启动仍生效，但下次启动会再生成新值)", err)
			_ = beego.AppConfig.Set("public_vkey", newVkey)
		}
	} else {
		_ = beego.AppConfig.Set("public_vkey", newVkey)
	}
	logs.Warn("public_vkey 缺失或为弱值，已自动生成新的 20 位密钥 (仅本次启动打印明文): %s", newVkey)
	logs.Warn("public_vkey 可在 Web 设置页查看/修改；如不需要公共客户端，请清空该项")
}

// isWeakPublicVkey returns true when the given value should be
// auto-rotated. Empty / "0" disable the feature; "123" is the
// historical default shipped in conf/nps.conf; len < 16 is a sanity
// floor — anything shorter is brute-forceable.
func isWeakPublicVkey(v string) bool {
	switch v {
	case "", "0", "123":
		return true
	}
	return len(v) < publicVkeyMinLen
}

// generatePublicVkey draws n bytes from crypto/rand and maps each
// byte to publicVkeyAlphabet using rejection-free modulo (alphabet
// size 36 divides 252 exactly when we treat the byte as 0..251 — we
// reroll the rare 252..255 values to keep the distribution uniform).
func generatePublicVkey(n int) (string, error) {
	if n <= 0 {
		n = 20
	}
	out := make([]byte, 0, n)
	buf := make([]byte, n*2) // overdraw to amortise rerolls
	for len(out) < n {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		for _, b := range buf {
			if b >= 252 { // 252,253,254,255 would bias modulo 36
				continue
			}
			out = append(out, publicVkeyAlphabet[int(b)%len(publicVkeyAlphabet)])
			if len(out) == n {
				break
			}
		}
	}
	return string(out), nil
}
