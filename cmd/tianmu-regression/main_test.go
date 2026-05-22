package main

import (
	"errors"
	"testing"
)

func TestIsReleaseGateBlocked(t *testing.T) {
	if !isReleaseGateBlocked(errors.New("release_gate_blocked: detected 1 critical control degradation cases")) {
		t.Fatal("release gate error must map to blocked exit code")
	}
	if isReleaseGateBlocked(errors.New("load dataset: missing file")) {
		t.Fatal("non-gate error must not map to blocked exit code")
	}
}
