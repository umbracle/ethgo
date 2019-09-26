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
	tmpl, err := template.New("eth-abi").Funcs(funcMap).Parse(templateAbiStr)
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
		var b bytes.Buffer
		if err := tmpl.Execute(&b, input); err != nil {
			return err
		}

		str := string(b.Bytes())
		if !strings.Contains(str, "*big.Int") {
			// remove the math/big library
			str = strings.Replace(str, "\"math/big\"\n", "", -1)
		}
		if err := ioutil.WriteFile(filepath.Join(config.Output, name+".go"), []byte(str), 0644); err != nil {
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
	"github.com/umbracle/go-web3/abi"
	"github.com/umbracle/go-web3/contract"
	"github.com/umbracle/go-web3/jsonrpc"
)

// {{.Name}} is a solidity contract
type {{.Name}} struct {
	c *contract.Contract
}

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
{{end}}{{end}}

var abi{{.Name}} *abi.ABI

func init() {
	var err error
	abi{{.Name}}, err = abi.NewABI(abi{{.Name}}Str)
	if err != nil {
		panic(fmt.Errorf("cannot parse {{.Name}} abi: %v", err))
	}
}

var abi{{.Name}}Str = ` + "`" + `{{.Contract.Abi}}` + "`"
