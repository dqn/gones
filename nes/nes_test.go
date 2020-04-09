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
