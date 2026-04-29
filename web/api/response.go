// Package api implements the RESTful HTTP API consumed by the Vue 3
// single-page application that lives under web-ui/. Every endpoint
// returns the same JSON envelope so the frontend can have a single
// response interceptor.
package api

import "github.com/astaxie/beego"

// Envelope is the common JSON response shape.
//
//	{ "code": 0, "message": "ok", "data": <any> }
//
// code == 0 means success. Non-zero codes are reserved for
// application-level errors (the HTTP status is also set accordingly so
// generic HTTP handling still works).
type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Page wraps a paginated list payload.
type Page struct {
	Total int64       `json:"total"`
	Items interface{} `json:"items"`
}

// baseController is embedded by every API controller. It provides the
// helpers below + the shared Prepare() that performs auth.
type baseController struct {
	beego.Controller
}

func (c *baseController) ok(data interface{}) {
	c.Data["json"] = Envelope{Code: 0, Message: "ok", Data: data}
	c.ServeJSON()
	c.StopRun()
}

func (c *baseController) okMsg(msg string) {
	c.Data["json"] = Envelope{Code: 0, Message: msg}
	c.ServeJSON()
	c.StopRun()
}

func (c *baseController) fail(httpStatus, code int, msg string) {
	c.Ctx.Output.SetStatus(httpStatus)
	c.Data["json"] = Envelope{Code: code, Message: msg}
	c.ServeJSON()
	c.StopRun()
}

func (c *baseController) badRequest(msg string)  { c.fail(400, 4000, msg) }
func (c *baseController) unauthorized(msg string) { c.fail(401, 4010, msg) }
func (c *baseController) forbidden(msg string)    { c.fail(403, 4030, msg) }
func (c *baseController) notFound(msg string)     { c.fail(404, 4040, msg) }
func (c *baseController) serverErr(msg string)    { c.fail(500, 5000, msg) }
