package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"path/filepath"

	"github.com/umbracle/go-web3/compiler"
)

func main() {
	var source string
	var pckg string
	var output string
	var name string

	flag.StringVar(&source, "source", "", "List of abi files")
	flag.StringVar(&pckg, "package", "main", "Name of the package")
	flag.StringVar(&output, "output", "", "Output directory")
	flag.StringVar(&name, "name", "", "name of the contract")

	flag.Parse()

	config := &config{
		Package: pckg,
		Output:  output,
		Name:    name,
	}

	artifacts, err := process(source, config)
	if err != nil {
		fmt.Printf("Failed to parse sources: %v", err)
		os.Exit(0)
	}
	if err := gen(artifacts, config); err != nil {
		fmt.Printf("Failed to generate sources: %v", err)
		os.Exit(0)
	}
}

const (
	vyExt  = 0
	solExt = 1
	abiExt = 2
)

func process(sources string, config *config) (map[string]*compiler.Artifact, error) {
	files := strings.Split(sources, ",")
	if len(files) == 0 {
		return nil, fmt.Errorf("input not found")
	}

	prev := -1
	for _, f := range files {
		var ext int
		switch filepath.Ext(f) {
		case ".abi":
			ext = abiExt
		case ".sol":
			ext = solExt
		case ".vy", ".py":
			ext = vyExt
		default:
			return nil, fmt.Errorf("file extension not found")
		}

		if prev == -1 {
			prev = ext
		} else if ext != prev {
			return nil, fmt.Errorf("two file formats found")
		}
	}

	switch prev {
	case abiExt:
		return processAbi(files, config)
	case solExt:
		return processSolc(files)
	case vyExt:
		return processVyper(files)
	}

	return nil, nil
}

func processVyper(sources []string)  (map[string]*compiler.Artifact, error) {
	c, err := compiler.NewCompiler("vyper", "vyper")
	if err != nil {
		return nil, err
	}
	raw, err := c.Compile(sources...)
	if err != nil {
		return nil, err
	}
	res := map[string]*compiler.Artifact{}
	for name, entry := range raw {
		_, filename := filepath.Split(name)
		filename = strings.TrimSuffix(filename, ".vy")
		filename = strings.TrimSuffix(filename, ".v.py")
		res[filename] = entry
	}
	return res, nil
}

func processSolc(sources []string) (map[string]*compiler.Artifact, error) {
	c, err := compiler.NewCompiler("solidity", "solc")
	if err != nil {
		return nil, err
	}
	raw, err := c.Compile(sources...)
	if err != nil {
		return nil, err
	}
	res := map[string]*compiler.Artifact{}
	for rawName, entry := range raw {
		name := strings.Split(rawName, ":")[1]
		res[name] = entry
	}
	return res, nil
}

func processAbi(sources []string, config *config) (map[string]*compiler.Artifact, error) {
	artifacts := map[string]*compiler.Artifact{}

	for _, abiPath := range sources {
		content, err := ioutil.ReadFile(abiPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to read file (%s): %v", abiPath, err)
		}

		// name of the contract
		name := filepath.Base(abiPath)
		name = strings.TrimSuffix(name, filepath.Ext(name))

		if len(sources) == 1 && config.Name != "" {
			name = config.Name
		}
		artifacts[name] = &compiler.Artifact{
			Abi: string(content),
		}
	}
	return artifacts, nil
}
