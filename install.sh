#!/bin/bash

set -e

os=""
arch="$(uname -m)"
repo="reposaur/reposaur"
tag="v0.5.0"
filename=""

if uname -a | grep Msys > /dev/null; then
  os="Windows"
elif uname -a | grep Darwin > /dev/null; then
  os="Darwin"
elif uname -a | grep Linux > /dev/null; then
  os="Linux"
fi

url="https://github.com/$repo/releases/download/$tag/rsr_${tag#v}_${os}_${arch}.tar.gz"

echo "OS: $os ($arch)"
echo "Tag: $tag"

curl -sL "$url" > rsr.tar.gz
tar zxf rsr.tar.gz
rm rsr.tar.gz
