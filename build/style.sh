#!/bin/bash

set -eu -o pipefail

PG_ROOT=$(dirname "${BASH_SOURCE}")/..
source "$PG_ROOT/build/common.sh"

diff=$(gofmt -d `pg::golang::find_files`)
if [[ -n ${diff} ]]; then
  echo "+++ Unformatted Go files:"
  echo "${diff}"
  exit 1
fi

go vet `pg::golang::find_dirs`
