#!/bin/bash

set -e

os=""
arch="$(uname -m)"
repo="reposaur/reposaur"
tag="v0.2.0"
filename=""

if uname -a | grep Msys > /dev/null; then
  os="Windows"
elif uname -a | grep Darwin > /dev/null; then
  os="Darwin"
elif uname -a | grep Linux > /dev/null; then
  os="Linux"
fi

url="https://github.com/$repo/releases/download/$tag/reposaur_${tag#v}_${os}_${arch}.tar.gz"

curl -L "$url" > reposaur.tar.gz
tar xvf reposaur.tar.gz
rm reposaur.tar.gz
