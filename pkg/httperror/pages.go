package httperror

import (
	"bytes"
	_ "embed"
)

//go:embed 503.html
var error503 []byte

var (
	templateVerson = []byte("{(VERSION)}")

	version = "v2"
)

// SetVersion allows you to set the current version of the program which will
// be displayed in the http error pages.
func SetVersion(v string) {
	version = v
}

// Get503 returns a rendered Error 503 service unavaliable error page
func Get503() []byte {
	return bytes.Replace(error503, templateVerson, []byte(version), -1)
}
