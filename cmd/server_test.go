package cmd

import (
	"testing"
)

func TestServerCommand(t *testing.T) {
	if serverCmd == nil {
		t.Fatal("serverCmd should be defined")
	}
	if serverCmd.Use != "server" {
		t.Errorf("expected command use 'server', got %s", serverCmd.Use)
	}
	portFlag := serverCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Error("expected 'port' flag to be defined")
	}
}

func TestGetServerKubeClient_InvalidPath(t *testing.T) {
	_, err := getServerKubeClient("/invalid/path", false)
	if err == nil {
		t.Error("expected error for invalid kubeconfig path")
	}
}

// import (
// 	"bytes"
// 	"errors"
// 	"strings"
// 	"testing"

// 	"github.com/rs/zerolog"
// 	"github.com/rs/zerolog/log"
// 	"github.com/valyala/fasthttp"
// )
// func TestRunServer(t *testing.T) {

// 	serverPort = 12345

// 	var buf bytes.Buffer
// 	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: &buf}).With().Timestamp().Logger()

// 	listenAndServe = func(addr string, handler fasthttp.RequestHandler) error {
// 		if addr != ":12345" {
// 			t.Errorf("Expected addr ':12345', got %s", addr)
// 		}
// 		return errors.New("mock server")
// 	}

// 	cmd := GetServerCmd()
// 	cmd.SetArgs([]string{"--port", "12345"})

// 	err := cmd.Execute()
// 	if err != nil {
// 		t.Errorf("Command execution failed: %v", err)
// 	}

// 	output := buf.String()
// 	expected := `"message":"Starting FastHTTP server on :12345"`
// 	if !strings.Contains(output, expected) {
// 		t.Errorf("expected output %q, got %q", expected, output)
// 	}
// }
