package main

import (
	"bytes"
	"testing"
)

func TestRunHealthcheck(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := run([]string{"healthcheck"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if got := stdout.String(); got != "ok\n" {
		t.Fatalf("expected stdout %q, got %q", "ok\\n", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := run([]string{"deploy"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if got := stderr.String(); got != "unknown command: deploy\n" {
		t.Fatalf("expected stderr %q, got %q", "unknown command: deploy\\n", got)
	}
}

func TestRunWithoutCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := run(nil, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if got := stderr.String(); got != "usage: releaserail <healthcheck>\n" {
		t.Fatalf("expected stderr %q, got %q", "usage: releaserail <healthcheck>\\n", got)
	}
}
