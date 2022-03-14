package localrelay

import (
	"io"
	"net"
	"net/http"
	"strings"
)

func relayHTTP(r *Relay, l net.Listener) error {

	r.logger.Info.Println("STARTING HTTP RELAY")

	return r.server.Serve(l)
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

	// BUG: sometimes requests redirect and cause a loop (Loop is auto stopped)
	req, err := http.NewRequest(r.Method, re.ForwardAddr+r.URL.Path+"?"+r.URL.Query().Encode(), r.Body)
	if err != nil {
		re.logger.Error.Println("BUILD REQUEST ERROR: ", err)
		return
	}

	// Append request headers
	for k, v := range r.Header {
		req.Header.Set(k, strings.Join(v, ","))
	}

	response, err := re.httpClient.Do(req)
	if err != nil {
		re.logger.Error.Println("FORWARD REQUEST ERROR: ", err)
		return
	}

	defer response.Body.Close()

	// Append response headers
	for k, v := range response.Header {
		w.Header().Set(k, strings.Join(v, ","))
	}

	w.WriteHeader(response.StatusCode)

	io.Copy(w, response.Body)
}
