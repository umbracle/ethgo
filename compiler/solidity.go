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

// SolcOutput is the output of the compilation.
type SolcOutput struct {
	Contracts map[string]*SolcContract
	Version   string
}

// SolcContract refers to a solidity compiled contract
type SolcContract struct {
	BinRuntime    string `json:"bin-runtime"`
	SrcMapRuntime string `json:"srcmap-runtime"`
	Bin           string
	SrcMap        string
	Abi           string
	Devdoc        string
	Userdoc       string
	Metadata      string
}

// Solidity is the solidity compiler
type Solidity struct {
	path string
}

// NewSolidityCompiler instantiates a new solidity compiler
func NewSolidityCompiler(path string) Compiler {
	return &Solidity{path}
}

// Compile implements the compiler interface
func (s *Solidity) Compile(code string) (interface{}, error) {
	args := []string{
		"--combined-json",
		"bin,bin-runtime,srcmap,srcmap-runtime,abi,userdoc,devdoc",
		"-",
	}

	cmd := exec.Command(s.path, args...)
	cmd.Stdin = strings.NewReader(code)

	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%v\n%s", err, stderr.Bytes())
	}

	var output *SolcOutput
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
