package testutil

import (
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func getOpenPort() string {
	rand.Seed(time.Now().UnixNano())
	min, max := 12000, 15000
	for {
		port := strconv.Itoa(rand.Intn(max-min) + min)
		server, err := net.Listen("tcp", ":"+port)
		if err == nil {
			server.Close()
			return port
		}
	}
}

type TestServerConfig struct {
	DataDir  string
	HTTPPort string
}

type ServerConfigCallback func(c *TestServerConfig)

// TestServer is a Geth test server
type TestServer struct {
	cmd    *exec.Cmd
	config *TestServerConfig
	t      *testing.T
}

// NewTestServer creates a new Geth test server
func NewTestServer(t *testing.T, cb ServerConfigCallback) *TestServer {
	path := "geth"

	vcmd := exec.Command(path, "version")
	vcmd.Stdout = nil
	vcmd.Stderr = nil
	if err := vcmd.Run(); err != nil {
		t.Skipf("geth version failed: %v", err)
	}

	dir, err := ioutil.TempDir("/tmp", "geth-")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	config := &TestServerConfig{
		DataDir:  dir,
		HTTPPort: getOpenPort(),
	}
	if cb != nil {
		cb(config)
	}

	// Build arguments
	args := []string{"--dev"}

	// add data dir
	args = append(args, "--datadir", filepath.Join(dir, "data"))

	// add ipcpath
	args = append(args, "--ipcpath", filepath.Join(dir, "geth.ipc"))

	// enable rpc
	args = append(args, "--rpc", "--rpcport", config.HTTPPort)

	// Start the server
	cmd := exec.Command(path, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		t.Fatalf("err: %s", err)
	}

	server := &TestServer{
		t:      t,
		cmd:    cmd,
		config: config,
	}

	// wait till the jsonrpc endpoint is running
	for {
		if server.testHTTPEndpoint() {
			break
		}
	}
	return server
}

func (t *TestServer) HttpAddr() string {
	return "http://localhost:" + t.config.HTTPPort
}

func (t *TestServer) testHTTPEndpoint() bool {
	resp, err := http.Post(t.HttpAddr(), "application/json", nil)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return true
}

func (t *TestServer) exit(err error) {
	t.Stop()
	t.t.Fatal(err)
}

// Stop stops the server
func (t *TestServer) Stop() {
	defer os.RemoveAll(t.config.DataDir)

	if err := t.cmd.Process.Kill(); err != nil {
		t.t.Errorf("err: %s", err)
	}
	t.cmd.Wait()
}
