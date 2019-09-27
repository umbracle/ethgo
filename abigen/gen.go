package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/umbracle/go-web3/abi"
	"github.com/umbracle/go-web3/compiler"
)

type config struct {
	Package string
	Output  string
	Name    string
}

func cleanName(str string) string {
	return strings.Trim(str, "_")
}

func outputArg(str string) string {
	if str == "" {

	}
	return str
}

func inputArg(str string) string {
	return str
}

func encodeArg(str interface{}) string {
	arg, ok := str.(*abi.Argument)
	if !ok {
		panic("bad")
	}

	switch arg.Type.Kind() {
	case abi.KindAddress:
		return "web3.Address"

	case abi.KindString:
		return "string"

	case abi.KindBool:
		return "bool"

	case abi.KindInt:
		return arg.Type.GoType().String()

	case abi.KindUInt:
		return arg.Type.GoType().String()

	case abi.KindFixedBytes:
		return fmt.Sprintf("[%d]byte", arg.Type.Size())

	case abi.KindBytes:
		return "[]byte"

	default:
		return fmt.Sprintf("input not done for type: %s", arg.Type.String())
	}
}

func gen(artifacts map[string]*compiler.Artifact, config *config) error {
	funcMap := template.FuncMap{
		"title":     strings.Title,
		"clean":     cleanName,
		"arg":       encodeArg,
		"outputArg": outputArg,
	}
	tmplAbi, err := template.New("eth-abi").Funcs(funcMap).Parse(templateAbiStr)
	if err != nil {
		return err
	}
	tmplBin, err := template.New("eth-abi").Funcs(funcMap).Parse(templateBinStr)
	if err != nil {
		return err
	}

	for name, artifact := range artifacts {
		// parse abi
		abi, err := abi.NewABI(artifact.Abi)
		if err != nil {
			return err
		}
		input := map[string]interface{}{
			"Ptr":      "a",
			"Config":   config,
			"Contract": artifact,
			"Abi":      abi,
			"Name":     name,
		}

		filename := strings.ToLower(name)

		var b bytes.Buffer
		if err := tmplAbi.Execute(&b, input); err != nil {
			return err
		}
		if err := ioutil.WriteFile(filepath.Join(config.Output, filename+".go"), []byte(b.Bytes()), 0644); err != nil {
			return err
		}

		b.Reset()
		if err := tmplBin.Execute(&b, input); err != nil {
			return err
		}
		if err := ioutil.WriteFile(filepath.Join(config.Output, filename+"_artifacts.go"), []byte(b.Bytes()), 0644); err != nil {
			return err
		}
	}
	return nil
}

var templateAbiStr = `package {{.Config.Package}}

import (
	"fmt"
	"math/big"

	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/contract"
	"github.com/umbracle/go-web3/jsonrpc"
)

var (
	_ = big.NewInt
)

// {{.Name}} is a solidity contract
type {{.Name}} struct {
	c *contract.Contract
}
{{if .Contract.Bin}}
// Deploy{{.Name}} deploys a new {{.Name}} contract
func Deploy{{.Name}}(provider *jsonrpc.Client, from web3.Address, args ...interface{}) *contract.Txn {
	return contract.DeployContract(provider, from, abi{{.Name}}, bin{{.Name}}, args...)
}
{{end}}
// New{{.Name}} creates a new instance of the contract at a specific address
func New{{.Name}}(addr web3.Address, provider *jsonrpc.Client) *{{.Name}} {
	return &{{.Name}}{c: contract.NewContract(addr, abi{{.Name}}, provider)}
}

// Contract returns the contract object
func ({{.Ptr}} *{{.Name}}) Contract() *contract.Contract {
	return {{.Ptr}}.c
}

// calls
{{range $key, $value := .Abi.Methods}}{{if .Const}}
// {{title $key}} calls the {{$key}} method in the solidity contract
func ({{$.Ptr}} *{{$.Name}}) {{title $key}}({{range $index, $val := .Inputs}}{{if .Name}}{{clean .Name}}{{else}}val{{$index}}{{end}} {{arg .}}, {{end}}block ...web3.BlockNumber) ({{range $index, $val := .Outputs}}val{{$index}} {{arg .}}, {{end}}err error) {
	var out map[string]interface{}
	{{ $length := len .Outputs }}{{ if ne $length 0 }}var ok bool{{ end }}

	out, err = {{$.Ptr}}.c.Call("{{$key}}", web3.EncodeBlock(block...){{range $index, $val := .Inputs}}, {{if .Name}}{{clean .Name}}{{else}}val{{$index}}{{end}}{{end}})
	if err != nil {
		return
	}

	// decode outputs
	{{range $index, $val := .Outputs}}val{{$index}}, ok = out["{{if .Name}}{{.Name}}{{else}}{{$index}}{{end}}"].({{arg .}})
	if !ok {
		err = fmt.Errorf("failed to encode output at index {{$index}}")
		return
	}
	{{end}}
	return
}
{{end}}{{end}}

// txns
{{range $key, $value := .Abi.Methods}}{{if not .Const}}
// {{title $key}} sends a {{$key}} transaction in the solidity contract
func ({{$.Ptr}} *{{$.Name}}) {{title $key}}({{range $index, $input := .Inputs}}{{if $index}}, {{end}}{{clean .Name}} {{arg .}}{{end}}) *contract.Txn {
	return {{$.Ptr}}.c.Txn("{{$key}}"{{range $index, $elem := .Inputs}}, {{clean $elem.Name}}{{end}})
}
{{end}}{{end}}`

var templateBinStr = `package {{.Config.Package}}

import (
	"encoding/hex"
	"fmt"

	"github.com/umbracle/go-web3/abi"
)

var abi{{.Name}} *abi.ABI

// {{.Name}}Abi returns the abi of the {{.Name}} contract
func {{.Name}}Abi() *abi.ABI {
	return abi{{.Name}}
}

var bin{{.Name}} []byte
{{if .Contract.Bin}}
// {{.Name}}Bin returns the bin of the {{.Name}} contract
func {{.Name}}Bin() []byte {
	return bin{{.Name}}
}
{{end}}
func init() {
	var err error
	abi{{.Name}}, err = abi.NewABI(abi{{.Name}}Str)
	if err != nil {
		panic(fmt.Errorf("cannot parse {{.Name}} abi: %v", err))
	}
	if len(bin{{.Name}}Str) != 0 {
		bin{{.Name}}, err = hex.DecodeString(bin{{.Name}}Str[2:])
		if err != nil {
			panic(fmt.Errorf("cannot parse {{.Name}} bin: %v", err))
		}
	}
}

var bin{{.Name}}Str = "{{.Contract.Bin}}"

var abi{{.Name}}Str = ` + "`" + `{{.Contract.Abi}}` + "`\n"
