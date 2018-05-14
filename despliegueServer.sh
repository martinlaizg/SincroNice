#!/bin/bash

echo 'export GOPATH='$HOME'/Escritorio' >> $HOME/.bashrc
go get golang.org/x/crypto/scrypt
go get github.com/howeyc/gopass