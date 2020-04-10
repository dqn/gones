package nes

import (
	"testing"
)

func TestNew(t *testing.T) {
	_, err := New("./../sample1/sample1.nes")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun(t *testing.T) {
	n, _ := New("./../sample1/sample1.nes")
	err := n.Run()
	if err != nil {
		t.Fatal(err)
	}
}
