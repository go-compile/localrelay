package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fasthttp/router"
	"github.com/go-compile/localrelay"
	"github.com/naoina/toml"

	"github.com/valyala/fasthttp"
)

type msgResponse struct {
	Message string `json:"message"`
}

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
	r.GET("/status", ipcRouteStatus)
	r.GET("/connections", ipcRouteConns)
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

	if err := json.Unmarshal(ctx.Request.Body(), &files); err != nil {
		ctx.SetStatusCode(400)
		ctx.Write([]byte(`{"message":"Invalid json body."}`))
		return
	}

	// TODO: support run multiple files
	if len(files) != 1 {
		ctx.SetStatusCode(400)
		ctx.Write([]byte(`{"message":"Endpoint currently requires at maximum and minimum one relay."}`))
		return
	}

	relayFile := files[0]

	exists, err := pathExists(relayFile)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.Write([]byte(`{"message":"Relay path could not be verified."}`))
		return
	}

	if !exists {
		ctx.SetStatusCode(404)
		ctx.Write([]byte(`{"message":"Relay file does not exist."}`))
		return
	}

	f, err := os.Open(relayFile)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.Write([]byte(`{"message":"Failed to open relay config."}`))
		return
	}

	var relay Relay
	if err := toml.NewDecoder(f).Decode(&relay); err != nil {
		f.Close()
		ctx.SetStatusCode(500)
		ctx.Write([]byte(`{"message":"Failed to decode relay config."}`))
		return
	}

	f.Close()

	if isRunning(relay.Name) {
		ctx.SetStatusCode(500)
		ctx.Write([]byte(`{"message":"Relay is already running."}`))
		return
	}

	if err := launchRelays([]Relay{relay}, false); err != nil {
		ctx.SetStatusCode(500)
		ctx.Write([]byte(`{"message":` + strconv.Quote("Error launching relay. "+err.Error()) + `}`))
		return
	}

	ctx.SetStatusCode(200)
	ctx.Write([]byte(`{"message":"Relay successfully launched."}`))
	return
}

func ipcRouteStatus(ctx *fasthttp.RequestCtx) {
	relayMetrics := make(map[string]metrics)

	relays := runningRelaysCopy()
	for _, r := range relays {
		active, total := r.Metrics.Connections()
		relayMetrics[r.Name] = metrics{
			In:            r.Metrics.Download(),
			Out:           r.Metrics.Upload(),
			Active:        active,
			DialAvg:       r.DialerAvg(),
			TotalConns:    total,
			TotalRequests: r.Metrics.Requests(),
		}
	}

	ctx.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(&status{
		Relays:  relays,
		Pid:     os.Getpid(),
		Version: VERSION,
		Started: daemonStarted.Unix(),

		Metrics: relayMetrics,
	})
}

func ipcRouteConns(ctx *fasthttp.RequestCtx) {
	relayConns := make([]connection, 0, 200)

	relays := runningRelaysCopy()
	for _, r := range relays {
		for _, conn := range r.GetConns() {

			relayConns = append(relayConns, connection{
				LocalAddr:  conn.Conn.LocalAddr().String(),
				RemoteAddr: conn.Conn.RemoteAddr().String(),
				Network:    conn.Conn.LocalAddr().Network(),

				RelayName:     r.Name,
				RelayHost:     r.Host,
				ForwardedAddr: conn.RemoteAddr,

				Opened: conn.Opened.Unix(),
			})
		}
	}

	ctx.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(relayConns)
}
