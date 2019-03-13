package compiler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
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
