module github.com/umbracle/ethgo/cmd

go 1.17

require (
	github.com/mitchellh/cli v1.1.2
	github.com/umbracle/ethgo v0.0.0-20220303093617-1621d9ff042b
)

require github.com/spf13/pflag v1.0.5

replace github.com/umbracle/ethgo => ../
