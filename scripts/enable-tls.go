//go:build ignore

// One-shot helper: set tls_enable=true and tls_bridge_port=8025 in
// the SQLite app_settings table, so NPS will start the TLS bridge
// listener on next restart. Run with:
//
//	go run scripts/enable-tls.go
//
// Build tag `ignore` keeps it out of normal `go build ./...`.
package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	dbPath := filepath.Join("conf", "nps.db")
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}
	db, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=busy_timeout(5000)")
	if err != nil {
		fmt.Println("open:", err)
		os.Exit(1)
	}
	defer db.Close()

	now := time.Now().Unix()
	upsert := func(k, v string) {
		_, err := db.Exec(
			`INSERT INTO app_settings(key,value,updated_at) VALUES(?,?,?)
			 ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at`,
			k, v, now)
		if err != nil {
			fmt.Printf("upsert %s: %v\n", k, err)
			os.Exit(1)
		}
		fmt.Printf("ok  %s = %s\n", k, v)
	}
	upsert("tls_enable", "true")
	upsert("tls_bridge_port", "8025")
	fmt.Println("done. restart nps to apply.")
}
