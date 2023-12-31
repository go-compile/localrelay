package localrelay_test

import (
	"testing"

	"github.com/go-compile/localrelay/v2"
)

func TestTargetParseTCP(t *testing.T) {
	target := localrelay.TargetLink("tcp://127.0.0.1:443")

	if target.Addr() != "127.0.0.1:443" {
		t.Error("unexpected target address")
	}

	if target.Host() != "127.0.0.1" {
		t.Error("unexpected target host")
	}

	if target.Port() != "443" {
		t.Error("unexpected target port")
	}

	if target.Protocol() != "tcp" {
		t.Error("unexpected target protocol")
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
