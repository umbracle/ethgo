#!/usr/bin/env bash

set -o errexit

echo "--> Build ENS"

ENS_ARTIFACTS=./contract/builtin/ens/artifacts
go run abigen/*.go --source ${ENS_ARTIFACTS}/ENS.abi,${ENS_ARTIFACTS}/Resolver.abi --output ./contract/builtin/ens --package ens

echo "--> Build ERC20"

ERC20_ARTIFACTS=./contract/builtin/erc20/artifacts
go run abigen/*.go --source ${ERC20_ARTIFACTS}/ERC20.abi --output ./contract/builtin/erc20 --package erc20

echo "--> Build Testdata"
go run abigen/*.go --source ./abigen/testdata/testdata.abi --output ./abigen/testdata --package testdata
