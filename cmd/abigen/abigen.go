package abigen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"path/filepath"

	"github.com/umbracle/ethgo/compiler"
)

func Parse(sources string, pckg string, output string) error {
	config := &config{
		Package: pckg,
		Output:  output,
	}

	if sources == "" {
		return fmt.Errorf("no source")
	}

	for _, source := range strings.Split(sources, ",") {
		matches, err := filepath.Glob(source)
		if err != nil {
			fmt.Printf("Failed to read files: %v", err)
			os.Exit(1)
		}
		if len(matches) == 0 {
			fmt.Printf("No match for source: %s\n", source)
			continue
		}
		for _, source := range matches {
			artifacts, err := process(source, config)
			if err != nil {
				fmt.Printf("Failed to parse sources: %v", err)
				os.Exit(1)
			}

			// hash the source file
			raw := sha256.Sum256([]byte(source))
			hash := hex.EncodeToString(raw[:])

			if err := gen(artifacts, config, hash); err != nil {
				fmt.Printf("Failed to generate sources: %v", err)
				os.Exit(1)
			}
		}
	}
	return nil
}

const (
	solExt  = 1
	abiExt  = 2
	jsonExt = 3
)

func process(sources string, config *config) (map[string]*compiler.Artifact, error) {
	files := strings.Split(sources, ",")
	if len(files) == 0 {
		return nil, fmt.Errorf("input not found")
	}

	prev := -1
	for _, f := range files {
		var ext int
		switch extt := filepath.Ext(f); extt {
		case ".abi":
			ext = abiExt
		case ".sol":
			ext = solExt
		case ".json":
			ext = jsonExt
		default:
			return nil, fmt.Errorf("file extension '%s' not found", extt)
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
	case jsonExt:
		return processJson(files)
	}

	return nil, nil
}

func processSolc(sources []string) (map[string]*compiler.Artifact, error) {
	c := compiler.NewSolidityCompiler("solc")
	raw, err := c.Compile(sources...)
	if err != nil {
		return nil, err
	}
	res := map[string]*compiler.Artifact{}
	for rawName, entry := range raw.Contracts {
		name := strings.Split(rawName, ":")[1]
		res[strings.Title(name)] = entry
	}
	return res, nil
}

func processAbi(sources []string, config *config) (map[string]*compiler.Artifact, error) {
	artifacts := map[string]*compiler.Artifact{}

	for _, abiPath := range sources {
		content, err := ioutil.ReadFile(abiPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read abi file (%s): %v", abiPath, err)
		}

		// Use the name of the file to name the contract
		path, name := filepath.Split(abiPath)

		name = strings.TrimSuffix(name, filepath.Ext(name))
		binPath := filepath.Join(path, name+".bin")

		bin, err := ioutil.ReadFile(binPath)
		if err != nil {
			// bin not found
			bin = []byte{}
		}
		artifacts[strings.Title(name)] = &compiler.Artifact{
			Abi: string(content),
			Bin: string(bin),
		}
	}
	return artifacts, nil
}

type JSONArtifact struct {
	Bytecode string          `json:"bytecode"`
	Abi      json.RawMessage `json:"abi"`
}

func processJson(sources []string) (map[string]*compiler.Artifact, error) {
	artifacts := map[string]*compiler.Artifact{}

	for _, jsonPath := range sources {
		content, err := ioutil.ReadFile(jsonPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read abi file (%s): %v", jsonPath, err)
		}

		// Use the name of the file to name the contract
		_, name := filepath.Split(jsonPath)
		name = strings.TrimSuffix(name, ".json")

		var art *JSONArtifact
		if err := json.Unmarshal(content, &art); err != nil {
			return nil, err
		}

		artifacts[strings.Title(name)] = &compiler.Artifact{
			Abi: string(art.Abi),
			Bin: "0x" + art.Bytecode,
		}
	}
	return artifacts, nil
}
