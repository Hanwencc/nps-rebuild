package main

import (
	"path/filepath"

	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/crypt"
	"ehang.io/nps/lib/file/sqlitedb"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

// initBridgeTlsCert replaces the legacy ephemeral crypt.InitTls() with
// a persistent on-disk keypair so the bridge's SHA-256 fingerprint
// stays stable across restarts. NPC operators pin this fingerprint
// via tls_server_fingerprint to defeat MITM.
//
// Locations: <runpath>/conf/server.pem and <runpath>/conf/server.key
// (overrideable in nps.conf via tls_cert_file / tls_key_file).
//
// `storeAny` is the untyped sqlitedb store kept on JsonDb (so the npc
// build doesn't link in modernc.org/sqlite). nil / wrong-type values
// are tolerated — we just skip the persistence step.
func initBridgeTlsCert(storeAny any) {
	store, _ := storeAny.(*sqlitedb.Store)
	confDir := filepath.Join(common.GetRunPath(), "conf")
	certPath := beego.AppConfig.DefaultString("tls_cert_file",
		filepath.Join(confDir, "server.pem"))
	keyPath := beego.AppConfig.DefaultString("tls_key_file",
		filepath.Join(confDir, "server.key"))

	if err := crypt.LoadOrInitServerCert(certPath, keyPath); err != nil {
		// Fall back to ephemeral cert so the bridge still comes up,
		// but loudly: pinning will be impossible until disk perms are
		// fixed.
		logs.Error("tls cert: load/init failed (%v) — falling back to in-memory ephemeral cert; pinning disabled", err)
		crypt.InitTls()
	}

	fp := crypt.GetCertFingerprintHex()
	if fp == "" {
		logs.Error("tls cert: fingerprint unavailable (cert load returned empty chain)")
		return
	}
	logs.Warn("bridge TLS cert SHA-256 指纹: %s", fp)
	logs.Warn("请将此值复制到 NPC 配置 tls_server_fingerprint=<指纹>，否则 NPC 无法防御中间人攻击")

	// Persist for the Web UI / API to display. The on-disk PEM is
	// authoritative — this row is overwritten on every boot to track
	// any cert rotation, so manual edits via the UI are not honoured.
	if store != nil {
		if err := store.UpsertSetting("tls_cert_fingerprint", fp); err != nil {
			logs.Warn("tls cert: failed to persist fingerprint to app_settings: %v", err)
		}
	}
	_ = beego.AppConfig.Set("tls_cert_fingerprint", fp)
}
