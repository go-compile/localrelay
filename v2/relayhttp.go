package localrelay

import (
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func relayHTTP(r *Relay, l net.Listener) error {

	r.logger.Info.Println("STARTING HTTP RELAY")

	return r.httpServer.Serve(l)
}

// HandleHTTP is to be used as the HTTP relay's handler set in the
// http.Server object
func HandleHTTP(relay *Relay) http.HandlerFunc {
	// Forwards relay object to request handler
	return func(w http.ResponseWriter, r *http.Request) {
		handleHTTP(w, r, relay)
	}
}

func handleHTTP(w http.ResponseWriter, r *http.Request, re *Relay) {
	re.Metrics.requests(1)

	remoteURL := re.ForwardAddr + r.URL.Path + "?" + r.URL.Query().Encode()

	// BUG: sometimes requests redirect and cause a loop (Loop is auto stopped)
	req, err := http.NewRequest(r.Method, remoteURL, r.Body)
	if err != nil {
		re.logger.Error.Println("BUILD REQUEST ERROR: ", err)
		return
	}

	re.Metrics.bandwidth(int(req.ContentLength)+len(remoteURL), 0)

	// Append request headers
	for k, v := range r.Header {
		req.Header.Set(k, strings.Join(v, ","))
	}

	start := time.Now()
	response, err := re.httpClient.Do(req)
	if err != nil {
		re.logger.Error.Println("FORWARD REQUEST ERROR: ", err)
		re.Metrics.dial(0, 1, start)
		return
	}

	re.Metrics.dial(1, 0, start)

	defer response.Body.Close()

	// Append response headers
	for k, v := range response.Header {
		w.Header().Set(k, strings.Join(v, ","))
	}

	w.WriteHeader(response.StatusCode)

	in, _ := io.Copy(w, response.Body)
	re.Metrics.bandwidth(0, int(in))

}
