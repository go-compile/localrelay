package localrelay_test

import (
	"testing"

	"github.com/go-compile/localrelay/v2"
)

func TestTargetParseTCP(t *testing.T) {
	target := localrelay.TargetLink("tcp://127.0.0.1:443")

	if x := target.Addr(); x != "127.0.0.1:443" {
		t.Errorf("unexpected target address: %s", x)
	}

	if x := target.Host(); x != "127.0.0.1" {
		t.Errorf("unexpected target host: %s", x)
	}

	if x := target.Port(); x != "443" {
		t.Errorf("unexpected target port: %s", x)
	}

	if x := target.Protocol(); x != "tcp" {
		t.Errorf("unexpected target protocol: %s", x)
	}
}

func TestTargetParseHTTPS(t *testing.T) {
	target := localrelay.TargetLink("https://example.com")

	if target.Addr() != "example.com:443" {
		t.Error("unexpected target address")
	}

	if target.Host() != "example.com" {
		t.Error("unexpected target host")
	}

	if target.Port() != "443" {
		t.Error("unexpected target port")
	}

	if target.Protocol() != "https" {
		t.Error("unexpected target protocol")
	}
}
