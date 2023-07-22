package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/fasthttp/router"
	"github.com/go-compile/localrelay"

	"github.com/valyala/fasthttp"
)

func newIPCServer() *fasthttp.Server {
	r := router.New()
	assignIPCRoutes(r)

	return &fasthttp.Server{
		Handler:      ipcHeadersMiddleware(r.Handler),
		Name:         "localrelay-ipc",
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
	}
}

func assignIPCRoutes(r *router.Router) {
	r.GET("/", ipcRouteRoot)
	r.GET("/stop/{relay}", ipcRouteStop)
	r.POST("/run", ipcRouteRun)
}

func ipcHeadersMiddleware(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("application/json")
		handler(ctx)
	}
}

func ipcRouteRoot(ctx *fasthttp.RequestCtx) {
	ctx.Write([]byte(`{"version":"` + VERSION + `","commit":"` + COMMIT + `"}`))
}

func ipcRouteStop(ctx *fasthttp.RequestCtx) {
	relayName := ctx.UserValue("relay").(string)

	var relay *localrelay.Relay
	for _, r := range runningRelays() {
		if r.Name == strings.ToLower(relayName) {
			relay = r
			break
		}
	}

	// relay not found
	if relay == nil {
		ctx.SetStatusCode(404)
		ctx.Write([]byte(`{"message":"Relay not found."}`))
		return
	}

	if err := relay.Close(); err != nil {
		ctx.SetStatusCode(500)
		ctx.Write([]byte(`{"message":"Error encountered when trying to close the relay."}`))
		return
	}

	// send success
	ctx.SetStatusCode(200)
	ctx.Write([]byte(`{"message":"Relay has been closed."}`))
	return
}

func ipcRouteRun(ctx *fasthttp.RequestCtx) {
	var files []string
	if err := json.NewDecoder(ctx.Response.BodyStream()).Decode(&files); err != nil {
		ctx.SetStatusCode(400)
		ctx.Write([]byte(`{"message":"Invalid json body."}`))
		return
	}

}
