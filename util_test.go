package main

import (
	"testing"
)

func TestOutputDefaultDir(t *testing.T) {
	d := mustCreateOutDir("haha", false)
	if d != "haha" {
		t.Fail()
	}
}
