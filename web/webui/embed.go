// Package webui embeds the compiled Vue 3 single-page application
// (built into web-ui/dist) so that the NPS server can serve the SPA
// from the same binary as the API.
//
// To populate the embedded files, run:
//
//	cd web-ui && yarn install && yarn build
//
// The build output is copied into web/webui/dist/ via the build
// script (web-ui/scripts/copy-dist.mjs). When dist/ is empty the
// embedded FS still compiles — the runtime fallback below serves a
// helpful placeholder page.
package webui

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// SubFS returns the embedded dist directory as an fs.FS rooted at the
// dist folder so that "/" maps to dist/index.html.
func SubFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}

// HasIndex reports whether dist/index.html exists in the embedded FS.
// Used to decide whether to fall back to the legacy Beego templates.
func HasIndex() bool {
	sub, err := SubFS()
	if err != nil {
		return false
	}
	_, err = fs.Stat(sub, "index.html")
	return err == nil
}

// Handler returns an http.Handler that serves the SPA. Unknown paths
// fall back to index.html so client-side routing works (HTML5 history
// mode). Static asset paths (containing a "." in the last segment) get
// 404 instead of falling back, so missing JS/CSS shows up clearly.
func Handler(prefix string) http.Handler {
	sub, err := SubFS()
	if err != nil {
		return placeholder(err)
	}
	if !HasIndex() {
		return placeholder(errors.New(
			"web-ui dist is empty — run `cd web-ui && yarn install && yarn build`"))
	}
	indexBytes, ierr := fs.ReadFile(sub, "index.html")
	if ierr != nil {
		return placeholder(ierr)
	}
	fileServer := http.FileServer(http.FS(sub))
	serveIndex := func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(indexBytes)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, prefix)
		if p == "" || p == "/" {
			serveIndex(w)
			return
		}
		// Try to serve as static file; if missing AND the path looks
		// like a SPA route (no file extension on last segment), fall
		// back to index.html. Reading file directly avoids
		// http.FileServer's automatic redirect of /index.html → /.
		clean := strings.TrimPrefix(p, "/")
		if _, err := fs.Stat(sub, clean); err == nil {
			r2 := *r
			r2.URL.Path = p
			fileServer.ServeHTTP(w, &r2)
			return
		}
		last := clean
		if i := strings.LastIndex(clean, "/"); i >= 0 {
			last = clean[i+1:]
		}
		if strings.Contains(last, ".") {
			http.NotFound(w, r)
			return
		}
		serveIndex(w)
	})
}

func placeholder(err error) http.Handler {
	msg := "<!doctype html><meta charset=utf-8><title>NPS Web UI</title>" +
		"<body style=\"font-family:system-ui;padding:32px;color:#334155\">" +
		"<h2>NPS Web UI is not built yet</h2><p>" +
		htmlEscape(err.Error()) + "</p>" +
		"<pre style=\"background:#f1f5f9;padding:12px;border-radius:8px\">" +
		"cd web-ui\nyarn install\nyarn build</pre></body>"
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(msg))
	})
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}
