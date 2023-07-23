package main

import (
	"testing"
)

func TestFormatBytes(t *testing.T) {
	Println(formatBytes(283))
	Println(formatBytes(324235))
	Println(formatBytes(3242335))
	Println(formatBytes(2124235))
	Println(formatBytes(3242321352355))
}
