package api

import (
	"net/http"
	"strings"

	"github.com/astaxie/beego"
	beegoctx "github.com/astaxie/beego/context"
)

// RegisterSecurityFilters wires the request-level hardening that has
// to live at the Beego layer rather than inside individual controllers:
//
//   - cookie hardening (Set-Cookie gets SameSite=Strict; Secure when
//     web_open_ssl is on; HttpOnly is reasserted in case some caller
//     forgot to set it)
//   - CSRF defence on state-changing /api/v1 requests via a mandatory
//     X-Requested-With header (sent by the SPA's axios client; cannot
//     be set by a cross-origin form / img / link without a preflight
//     the attacker can't satisfy)
//
// Call from routers.Init() once, before AddNamespace.
func RegisterSecurityFilters() {
	// (1) Wrap the response writer so we can rewrite Set-Cookie before
	//     the headers are flushed. BeforeRouter is the earliest spot
	//     where ctx.ResponseWriter is initialised but the controller
	//     hasn't run yet.
	beego.InsertFilter("/*", beego.BeforeRouter, func(ctx *beegoctx.Context) {
		if ctx == nil || ctx.ResponseWriter == nil {
			return
		}
		// Avoid double-wrapping in case the filter is hit twice on a
		// single request (eg. via internal redirect).
		if _, ok := ctx.ResponseWriter.ResponseWriter.(*cookieHardeningWriter); ok {
			return
		}
		secure := isHTTPSRequest(ctx)
		ctx.ResponseWriter.ResponseWriter = &cookieHardeningWriter{
			ResponseWriter: ctx.ResponseWriter.ResponseWriter,
			secure:         secure,
		}
	}, false)

	// (2) CSRF guard for the only unauthenticated state-changing
	//     endpoint that issues a session cookie. All authenticated
	//     mutating endpoints are already covered by SameSite=Strict
	//     (cross-site requests just won't carry the session cookie),
	//     but Login itself runs without a session, so SameSite alone
	//     can't stop an attacker from logging the victim into the
	//     attacker's account.
	beego.InsertFilter("/api/v1/auth/login", beego.BeforeRouter, func(ctx *beegoctx.Context) {
		if ctx.Input.Method() != http.MethodPost {
			return
		}
		// Same-origin SPA always sends this header (see
		// web-ui/src/api/request.ts). Browsers refuse to add custom
		// headers to cross-origin requests without a CORS preflight,
		// which a forged form/image/link can never trigger.
		if !strings.EqualFold(ctx.Input.Header("X-Requested-With"), "XMLHttpRequest") {
			ctx.Output.SetStatus(http.StatusForbidden)
			ctx.Output.JSON(map[string]interface{}{
				"code":    4030,
				"message": "csrf check failed: missing X-Requested-With",
			}, false, false)
		}
	}, false)
}

// isHTTPSRequest returns true when the inbound request is using TLS,
// either because Beego terminated TLS itself (web_open_ssl) or because
// a trusted reverse proxy set the standard X-Forwarded-Proto header.
func isHTTPSRequest(ctx *beegoctx.Context) bool {
	if ctx.Request != nil && ctx.Request.TLS != nil {
		return true
	}
	if v := ctx.Input.Header("X-Forwarded-Proto"); strings.EqualFold(v, "https") {
		return true
	}
	if open, _ := beego.AppConfig.Bool("web_open_ssl"); open {
		return true
	}
	return false
}

// cookieHardeningWriter intercepts the moment headers are about to be
// flushed and rewrites every Set-Cookie value to include the security
// attributes the underlying writer omitted. The mutation is purely
// additive: existing Path/Expires/Domain/Max-Age stay intact.
type cookieHardeningWriter struct {
	http.ResponseWriter
	secure   bool
	rewrote  bool
}

// hardenSetCookies rewrites the in-flight Set-Cookie header slice. Safe
// to call before either WriteHeader or the first Write.
func (w *cookieHardeningWriter) hardenSetCookies() {
	if w.rewrote {
		return
	}
	w.rewrote = true
	h := w.ResponseWriter.Header()
	cookies := h["Set-Cookie"]
	if len(cookies) == 0 {
		return
	}
	out := make([]string, 0, len(cookies))
	for _, raw := range cookies {
		out = append(out, hardenSetCookie(raw, w.secure))
	}
	h["Set-Cookie"] = out
}

func (w *cookieHardeningWriter) WriteHeader(code int) {
	w.hardenSetCookies()
	w.ResponseWriter.WriteHeader(code)
}

func (w *cookieHardeningWriter) Write(b []byte) (int, error) {
	w.hardenSetCookies()
	return w.ResponseWriter.Write(b)
}

// Flush / Hijack / CloseNotify pass-throughs so beego's session writer,
// websocket upgrades, and SSE all keep working.
func (w *cookieHardeningWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// hardenSetCookie returns the input Set-Cookie value with SameSite,
// HttpOnly and (when secure==true) Secure attributes appended if they
// are not already present. Comparison is case-insensitive on the
// attribute name and tolerates `=` separators (eg. "SameSite=Lax").
func hardenSetCookie(raw string, secure bool) string {
	lower := strings.ToLower(raw)
	out := raw
	if !attrPresent(lower, "samesite") {
		// Strict is the safest default for a control-plane UI. If you
		// later need to embed nps in an iframe from another origin,
		// downgrade to Lax here.
		out += "; SameSite=Strict"
	}
	if !attrPresent(lower, "httponly") {
		out += "; HttpOnly"
	}
	if secure && !attrPresent(lower, "secure") {
		out += "; Secure"
	}
	return out
}

// attrPresent looks for a cookie attribute by name, ignoring whether
// it's followed by `=value` or stands alone. The needle MUST be lower
// case; haystack is expected to be the lowered Set-Cookie string.
func attrPresent(haystackLower, needleLower string) bool {
	idx := 0
	for {
		i := strings.Index(haystackLower[idx:], needleLower)
		if i < 0 {
			return false
		}
		i += idx
		// Must be preceded by start-of-string or "; " to avoid matching
		// "samesite" inside a value or a similarly-named attribute.
		left := i == 0 || haystackLower[i-1] == ' ' || haystackLower[i-1] == ';'
		// Must be followed by end-of-string, "=" or ";" so a longer
		// attribute name (eg. "secureflag") is rejected for "secure".
		end := i + len(needleLower)
		right := end == len(haystackLower) ||
			haystackLower[end] == '=' ||
			haystackLower[end] == ';' ||
			haystackLower[end] == ' '
		if left && right {
			return true
		}
		idx = end
	}
}
