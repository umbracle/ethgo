package compiler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Output struct {
	Contracts map[string]*Artifact
	Sources   map[string]*Source
	Version   string
}

type Source struct {
	AST map[string]interface{}
}

type Artifact struct {
	Abi           string
	Bin           string
	BinRuntime    string `json:"bin-runtime"`
	SrcMap        string `json:"srcmap"`
	SrcMapRuntime string `json:"srcmap-runtime"`
}

// Solidity is the solidity compiler
type Solidity struct {
	path string
}

// NewSolidityCompiler instantiates a new solidity compiler
func NewSolidityCompiler(path string) *Solidity {
	return &Solidity{path}
}

// CompileCode compiles a solidity code
func (s *Solidity) CompileCode(code string) (*Output, error) {
	if code == "" {
		return nil, fmt.Errorf("code is empty")
	}
	output, err := s.compileImpl(code)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// Compile implements the compiler interface
func (s *Solidity) Compile(files ...string) (*Output, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no input files")
	}
	return s.compileImpl("", files...)
}

func (s *Solidity) compileImpl(code string, files ...string) (*Output, error) {
	args := []string{
		"--combined-json",
		"bin,bin-runtime,srcmap-runtime,abi,srcmap,ast",
	}
	if code != "" {
		args = append(args, "-")
	}
	if len(files) != 0 {
		args = append(args, files...)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(s.path, args...)
	if code != "" {
		cmd.Stdin = strings.NewReader(code)
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to compile: %s", string(stderr.Bytes()))
	}

	var output *Output
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return nil, err
	}
	return output, nil
}

// DownloadSolidity downloads the solidity compiler
func DownloadSolidity(version string, dst string, renameDst bool) error {
	url := "https://github.com/ethereum/solidity/releases/download/v" + version + "/solc-static-linux"

	// check if the dst is correct
	exists := false
	fi, err := os.Stat(dst)
	if err == nil {
		switch mode := fi.Mode(); {
		case mode.IsDir():
			exists = true
		case mode.IsRegular():
			return fmt.Errorf("dst is a file")
		}
	} else {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat dst '%s': %v", dst, err)
		}
	}

	// create the destiny path if does not exists
	if !exists {
		if err := os.MkdirAll(dst, 0755); err != nil {
			return fmt.Errorf("cannot create dst path: %v", err)
		}
	}

	// rename binary
	name := "solidity"
	if renameDst {
		name += "-" + version
	}

	// tmp folder to download the binary
	tmpDir, err := ioutil.TempDir("/tmp", "solc-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, name)

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// make binary executable
	if err := os.Chmod(path, 0755); err != nil {
		return err
	}

	// move file to dst
	if err := os.Rename(path, filepath.Join(dst, name)); err != nil {
		return err
	}
	return nil
}
