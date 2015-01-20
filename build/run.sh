#!/usr/bin/bash

set -eu -o pipefail

PG_ROOT=$(dirname "${BASH_SOURCE}")/..
source "$PG_ROOT/build/common.sh"

pg::build::verify

pg::build::build_image

readonly GOPATH=/go/src
readonly PROJECT_NAME=service-notify
readonly PROJECT_NAME_FULL=github.com/protogalaxy/service-notify
readonly PROJECT_DIR="${GOPATH}/${PROJECT_NAME_FULL}"

pg::build::run_build $PROJECT_DIR

pg::build::build_release_image "${PROJECT_NAME}"
