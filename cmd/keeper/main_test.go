package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain_VersionCommand(t *testing.T) {
	originalArgs := os.Args
	originalStdout := os.Stdout
	defer func() {
		os.Args = originalArgs
		os.Stdout = originalStdout
	}()

	buildVersion = "test-version"
	buildDate = "2026-03-23"
	os.Args = []string{"keeper", "version"}

	reader, writer, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdout = writer

	main()

	assert.NoError(t, writer.Close())
	var output bytes.Buffer
	_, err = io.Copy(&output, reader)
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Build version: test-version")
	assert.Contains(t, output.String(), "Build date: 2026-03-23")
}
