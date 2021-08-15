#!/usr/bin/env bash

# set -o errexit

install_geth () {
    VERSION=1.9.1-b7b2f60f

    NAME=geth-linux-amd64-$VERSION
    RELEASE=https://gethstore.blob.core.windows.net/builds/$NAME.tar.gz

    curl -sfSO "${RELEASE}"
    tar -xvzf $NAME.tar.gz

    mv $NAME/geth /usr/local/bin/geth
}

install_solidity() {
    VERSION="0.5.5"
    DOWNLOAD=https://github.com/ethereum/solidity/releases/download/v${VERSION}/solc-static-linux

    curl -L $DOWNLOAD > /tmp/solc
    chmod +x /tmp/solc
    mv /tmp/solc /usr/local/bin/solc
}

install_vyper() {
    sudo apt-get update
    sudo apt-get install -y -qq python3.6 virtualenv

    virtualenv -p python3 --no-site-packages ~/vyper-venv
    source ~/vyper-venv/bin/activate

    git clone https://github.com/ethereum/vyper.git
    cd vyper
    make

    cp ~/vyper-venv/bin/vyper /usr/local/bin
}

#install_geth
install_solidity
