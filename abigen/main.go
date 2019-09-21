package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"path/filepath"

	"github.com/umbracle/go-web3/abi"
)

func main() {
	var abisStr string
	var pckg string
	var output string
	var name string

	flag.StringVar(&abisStr, "abis", "", "List of abi files")
	flag.StringVar(&pckg, "package", "main", "Name of the package")
	flag.StringVar(&output, "output", "", "Output directory")
	flag.StringVar(&name, "name", "", "name of the contract")

	flag.Parse()

	config := &config{
		Package: pckg,
		Output:  output,
		Name:    name,
	}
	if err := process(abisStr, config); err != nil {
		panic(err)
	}
}

func process(abisStr string, config *config) error {
	abis := strings.Split(abisStr, ",")

	// check if all the abis exist
	for _, abi := range abis {
		_, err := os.Stat(abi)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("ABI file does not exists (%s)", abi)
			}
			return fmt.Errorf("Failed to stat (%s): %v", abi, err)
		}
	}

	artifacts := []*artifact{}

	// read the content
	for _, abiPath := range abis {
		content, err := ioutil.ReadFile(abiPath)
		if err != nil {
			return fmt.Errorf("Failed to read file (%s): %v", abiPath, err)
		}

		// name of the contract
		name := filepath.Base(abiPath)
		name = strings.TrimSuffix(name, filepath.Ext(name))

		if len(abis) == 1 && config.Name != "" {
			name = config.Name
		}

		a, err := abi.NewABI(string(content))
		if err != nil {
			return err
		}
		artifact := &artifact{
			Name:   name,
			ABI:    a,
			ABIStr: string(content),
		}
		artifacts = append(artifacts, artifact)
	}

	gen(artifacts, config)
	return nil
}
