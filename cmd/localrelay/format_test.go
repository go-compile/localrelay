package main

import (
	"fmt"
	"testing"
)

func TestFormatBytes(t *testing.T) {
	fmt.Println(formatBytes(283))
	fmt.Println(formatBytes(324235))
	fmt.Println(formatBytes(3242335))
	fmt.Println(formatBytes(2124235))
	fmt.Println(formatBytes(3242321352355))
}
