#!/bin/sh
set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish; do
	go run cmd/rsr/rsr.go completion "$sh" > "completions/rsr.$sh"
done