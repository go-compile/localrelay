package httperror

import _ "embed"

//go:embed 503.html
var error503 []byte

func Get503() []byte {
	return error503
}
