package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/umbracle/go-web3/abi"
)

type artifact struct {
	Name   string
	ABI    *abi.ABI
	ABIStr string
	Bin    string
}

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
	arg, ok := str.(*abi.ArgumentObj)
	if !ok {
		panic("bad")
	}

	switch arg.Type.Kind() {
	case abi.KindAddress:
		return "[20]byte"

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

func gen(artifacts []*artifact, config *config) error {
	funcMap := template.FuncMap{
		"title":     strings.Title,
		"clean":     cleanName,
		"arg":       encodeArg,
		"outputArg": outputArg,
	}
	tmpl, err := template.New("eth-abi").Funcs(funcMap).Parse(templateAbiStr)
	if err != nil {
		panic(err)
	}

	for _, artifact := range artifacts {
		input := map[string]interface{}{
			"Ptr":      "a",
			"Config":   config,
			"Contract": artifact,
		}
		var b bytes.Buffer
		if err := tmpl.Execute(&b, input); err != nil {
			panic(err)
		}

		str := string(b.Bytes())
		if !strings.Contains(str, "*big.Int") {
			// remove the math/big library
			str = strings.Replace(str, "\"math/big\"\n", "", -1)
		}

		if err := ioutil.WriteFile(filepath.Join(config.Output, artifact.Name+".go"), []byte(str), 0644); err != nil {
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

// {{.Contract.Name}} is a solidity contract
type {{.Contract.Name}} struct {
	c *contract.Contract
}

// New{{.Contract.Name}} creates a new instance of the contract at a specific address
func New{{.Contract.Name}}(addr string, provider *jsonrpc.Client) *{{.Contract.Name}}{
	return &{{.Contract.Name}}{c: contract.NewContract(addr, abi{{.Contract.Name}}, provider)}
}

// Contract returns the contract object
func ({{.Ptr}}* {{.Contract.Name}}) Contract() *contract.Contract {
	return {{.Ptr}}.c
}

// calls
{{range $key, $value := .Contract.ABI.Methods}}{{if .Const}}
// {{title $key}} calls the {{$key}} method in the solidity contract
func ({{$.Ptr}}* {{$.Contract.Name}}) {{title $key}}({{range .Inputs}}{{clean .Name}} {{arg .}}, {{end}}block ...web3.BlockNumber) ({{range $index, $val := .Outputs}}val{{$index}} {{arg .}}, {{end}}err error) {
	var out map[string]interface{}
	{{ $length := len .Outputs }}{{ if ne $length 0 }}var ok bool{{ end }}

	out, err = {{$.Ptr}}.c.Call("{{$key}}", web3.EncodeBlock(block...){{range .Inputs}}, {{clean .Name}}{{end}})
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
{{range $key, $value := .Contract.ABI.Methods}}{{if not .Const}}
// {{title $key}} sends a {{$key}} transaction in the solidity contract
func ({{$.Ptr}}* {{$.Contract.Name}}) {{title $key}}({{range $index, $input := .Inputs}}{{if $index}}, {{end}}{{clean .Name}} {{arg .}}{{end}}) *contract.Txn {
	return {{$.Ptr}}.c.Txn("{{$key}}"{{range $index, $elem := .Inputs}}, {{clean $elem.Name}}{{end}})
}
{{end}}{{end}}

var abi{{.Contract.Name}} *abi.ABI

func init() {
	var err error
	abi{{.Contract.Name}}, err = abi.NewABI(abi{{.Contract.Name}}Str)
	if err != nil {
		panic(fmt.Errorf("cannot parse {{.Contract.Name}} abi: %v", err))
	}
}

var bin{{.Contract.Name}} = []byte{}

var abi{{.Contract.Name}}Str = ` + "`" + `{{.Contract.ABIStr}}` + "`"
