package compiler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-getter"
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
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &getter.Client{
		Ctx:     ctx,
		Src:     url,
		Dst:     dst,
		Pwd:     pwd,
		Mode:    getter.ClientModeAny,
		Options: []getter.ClientOption{},
	}
	if err := client.Get(); err != nil {
		return err
	}

	// rename binary
	name := "solidity"
	if renameDst {
		name += "-" + version
	}
	dstPath := filepath.Join(dst, name)
	if err := os.Rename(filepath.Join(dst, "solc-static-linux"), dstPath); err != nil {
		return err
	}

	// make binary executable
	if err := os.Chmod(dstPath, 0755); err != nil {
		return err
	}
	return nil
}
