#!/bin/bash

function pg::golang::find_dirs() {
  find . -not \( \
    -path ./target \
    -o -path '*/Godeps/*' \
  \) -name '*.go' -print0 | xargs -0n1 dirname | sort -u
}

function pg::golang::find_test_dirs() {
  find . -not \( \
    -path ./target \
    -o -path '*/Godeps/*' \
  \) -name '*_test.go' -print0 | xargs -0n1 dirname | sort -u
}

function pg::golang::find_files() {
  find . -not \( \
    -path ./target \
    -o -path '*/Godeps/*' \
  \) -name '*.go'
}
