package api

import (
	"encoding/json"
	"strings"
	"time"

	"ehang.io/nps/lib/file"
	"ehang.io/nps/lib/file/sqlitedb"
)

// TokenController exposes CRUD over file.ApiToken. Admin-only.
//
// Routes (registered in router.go):
//
//	GET    /api/v1/tokens           List
//	POST   /api/v1/tokens           Create   (returns plaintext secret ONCE)
//	GET    /api/v1/tokens/:id       Get
//	PUT    /api/v1/tokens/:id       Update   (everything except the secret)
//	DELETE /api/v1/tokens/:id       Delete
//	POST   /api/v1/tokens/:id/rotate Rotate  (issues new secret, returns it ONCE)
type TokenController struct {
	baseController
}

// payload mirrors file.ApiToken minus the secret hash; the secret is
// only ever surfaced via `secret` in create/rotate responses.
type tokenPayload struct {
	Id                int      `json:"id"`
	KeyId             string   `json:"keyId"`
	Remark            string   `json:"remark"`
	AllowedPathPrefix string   `json:"allowedPathPrefix"`
	AllowedMethods    []string `json:"allowedMethods"`
	AllowIps          []string `json:"allowIps"`
	ExpiresAt         int64    `json:"expiresAt"`
	CreatedAt         int64    `json:"createdAt"`
	LastUsedAt        int64    `json:"lastUsedAt"`
	LastUsedIp        string   `json:"lastUsedIp"`
	Disabled          bool     `json:"disabled"`
}

func toTokenPayload(t *file.ApiToken) tokenPayload {
	return tokenPayload{
		Id:                t.Id,
		KeyId:             t.KeyId,
		Remark:            t.Remark,
		AllowedPathPrefix: t.AllowedPathPrefix,
		AllowedMethods:    t.AllowedMethods,
		AllowIps:          t.AllowIps,
		ExpiresAt:         t.ExpiresAt,
		CreatedAt:         t.CreatedAt,
		LastUsedAt:        t.LastUsedAt,
		LastUsedIp:        t.LastUsedIp,
		Disabled:          t.Disabled,
	}
}

// tokenWriteRequest is the body accepted by Create/Update. Fields not
// supplied keep their previous values on Update.
type tokenWriteRequest struct {
	Remark            *string   `json:"remark"`
	AllowedPathPrefix *string   `json:"allowedPathPrefix"`
	AllowedMethods    *[]string `json:"allowedMethods"`
	AllowIps          *[]string `json:"allowIps"`
	ExpiresAt         *int64    `json:"expiresAt"`
	Disabled          *bool     `json:"disabled"`
}

func (c *TokenController) requireAdmin() bool {
	if !c.currentIsAdmin() {
		c.forbidden("permission denied")
		return false
	}
	return true
}

// store returns the SQLite-backed api_tokens store, or nil after
// emitting a 5xx response.
func (c *TokenController) store() *sqlitedb.Store {
	s := sqlitedb.From(file.GetDb())
	if s == nil {
		c.serverErr("sqlite store not initialised")
		return nil
	}
	return s
}

// List GET /api/v1/tokens
func (c *TokenController) List() {
	if !c.requireAdmin() {
		return
	}
	s := c.store()
	if s == nil {
		return
	}
	rows, err := s.ListApiTokens()
	if err != nil {
		c.serverErr(err.Error())
		return
	}
	out := make([]tokenPayload, 0, len(rows))
	for _, t := range rows {
		out = append(out, toTokenPayload(t))
	}
	c.ok(out)
}

// Get GET /api/v1/tokens/:id
func (c *TokenController) Get() {
	if !c.requireAdmin() {
		return
	}
	id, err := c.GetInt(":id")
	if err != nil {
		c.badRequest("invalid id")
		return
	}
	s := c.store()
	if s == nil {
		return
	}
	t, err := s.GetApiToken(id)
	if err != nil {
		c.notFound(err.Error())
		return
	}
	c.ok(toTokenPayload(t))
}

// Create POST /api/v1/tokens
//
// Response data:
//
//	{ "token": <tokenPayload>, "secret": "<plaintext, shown ONCE>" }
func (c *TokenController) Create() {
	if !c.requireAdmin() {
		return
	}
	req := tokenWriteRequest{}
	if body := c.Ctx.Input.RequestBody; len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			c.badRequest("invalid JSON body: " + err.Error())
			return
		}
	}
	t := &file.ApiToken{
		AllowedMethods: []string{},
		AllowIps:       []string{},
	}
	applyTokenWrite(t, &req)

	keyId, err := file.GenerateApiTokenKeyId()
	if err != nil {
		c.serverErr("failed to generate keyId: " + err.Error())
		return
	}
	secret, err := file.GenerateApiTokenSecret()
	if err != nil {
		c.serverErr("failed to generate secret: " + err.Error())
		return
	}
	hash, err := file.HashApiTokenSecret(secret)
	if err != nil {
		c.serverErr("failed to hash secret: " + err.Error())
		return
	}
	t.KeyId = keyId
	t.SecretHash = hash
	t.CreatedAt = time.Now().Unix()

	s := c.store()
	if s == nil {
		return
	}
	if err := s.NewApiToken(t); err != nil {
		c.serverErr("insert api_token failed: " + err.Error())
		return
	}

	c.ok(map[string]interface{}{
		"token":  toTokenPayload(t),
		"secret": secret,
	})
}

// Update PUT /api/v1/tokens/:id
func (c *TokenController) Update() {
	if !c.requireAdmin() {
		return
	}
	id, err := c.GetInt(":id")
	if err != nil {
		c.badRequest("invalid id")
		return
	}
	s := c.store()
	if s == nil {
		return
	}
	t, err := s.GetApiToken(id)
	if err != nil {
		c.notFound(err.Error())
		return
	}
	req := tokenWriteRequest{}
	if body := c.Ctx.Input.RequestBody; len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			c.badRequest("invalid JSON body: " + err.Error())
			return
		}
	}
	applyTokenWrite(t, &req)
	if err := s.UpdateApiToken(t); err != nil {
		c.serverErr("update api_token failed: " + err.Error())
		return
	}
	c.ok(toTokenPayload(t))
}

// Delete DELETE /api/v1/tokens/:id
func (c *TokenController) Delete() {
	if !c.requireAdmin() {
		return
	}
	id, err := c.GetInt(":id")
	if err != nil {
		c.badRequest("invalid id")
		return
	}
	s := c.store()
	if s == nil {
		return
	}
	if _, err := s.GetApiToken(id); err != nil {
		c.notFound(err.Error())
		return
	}
	if err := s.DelApiToken(id); err != nil {
		c.serverErr("delete api_token failed: " + err.Error())
		return
	}
	c.okMsg("deleted")
}

// Rotate POST /api/v1/tokens/:id/rotate
//
// Issues a fresh secret and invalidates the previous one. Response is
// the same shape as Create.
func (c *TokenController) Rotate() {
	if !c.requireAdmin() {
		return
	}
	id, err := c.GetInt(":id")
	if err != nil {
		c.badRequest("invalid id")
		return
	}
	s := c.store()
	if s == nil {
		return
	}
	t, err := s.GetApiToken(id)
	if err != nil {
		c.notFound(err.Error())
		return
	}
	secret, err := file.GenerateApiTokenSecret()
	if err != nil {
		c.serverErr("failed to generate secret: " + err.Error())
		return
	}
	hash, err := file.HashApiTokenSecret(secret)
	if err != nil {
		c.serverErr("failed to hash secret: " + err.Error())
		return
	}
	t.SecretHash = hash
	if err := s.UpdateApiToken(t); err != nil {
		c.serverErr("rotate api_token failed: " + err.Error())
		return
	}
	c.ok(map[string]interface{}{
		"token":  toTokenPayload(t),
		"secret": secret,
	})
}

// applyTokenWrite copies non-nil fields from req into t and normalizes
// list fields (trimmed, deduplicated, empty entries dropped).
func applyTokenWrite(t *file.ApiToken, req *tokenWriteRequest) {
	if req.Remark != nil {
		t.Remark = strings.TrimSpace(*req.Remark)
	}
	if req.AllowedPathPrefix != nil {
		p := strings.TrimSpace(*req.AllowedPathPrefix)
		if p != "" && !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		t.AllowedPathPrefix = p
	}
	if req.AllowedMethods != nil {
		t.AllowedMethods = cleanStringSlice(*req.AllowedMethods, true)
	}
	if req.AllowIps != nil {
		t.AllowIps = cleanStringSlice(*req.AllowIps, false)
	}
	if req.ExpiresAt != nil {
		t.ExpiresAt = *req.ExpiresAt
	}
	if req.Disabled != nil {
		t.Disabled = *req.Disabled
	}
}

func cleanStringSlice(in []string, upper bool) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, raw := range in {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		if upper {
			s = strings.ToUpper(s)
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
