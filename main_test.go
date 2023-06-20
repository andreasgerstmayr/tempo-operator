package main

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestSetupLogging(t *testing.T) {
	prevStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	setupLogging()
	log := ctrl.LoggerFrom(context.Background())
	log = log.WithValues("tempo", "simplest")
	log.V(1).Info("a test debug message")
	log.Info("a test info message")
	log.Error(errors.New("test error"), "a test error occurred")

	err := w.Close()
	require.NoError(t, err)
	output, _ := io.ReadAll(r)
	os.Stderr = prevStderr

	require.Regexp(t, `{"level":"info","ts":".+","msg":"a test info message","tempo":"simplest"}`, string(output))
	require.NotRegexp(t, `{"level":"debug",.+}`, string(output))
	require.Regexp(t, `{"level":"error","ts":".+","msg":"a test error occurred","tempo":"simplest","error":"test error","stacktrace":`, string(output))
}
