#!/bin/bash

set -eu -o pipefail

PG_ROOT=$(dirname "${BASH_SOURCE}")/..
source "$PG_ROOT/build/common.sh"

pg::build::verify

pg::build::build_image

readonly GOPATH=/go/src
readonly PROJECT_NAME=service-notify
readonly PROJECT_PATH="github.com/protogalaxy/${PROJECT_NAME}"
readonly PROJECT_DIR="${GOPATH}/${PROJECT_PATH}"

pg::build::run_command build/test.sh
pg::build::run_command build/style.sh
pg::build::run_build $PROJECT_DIR

pg::build::build_release_image "${PROJECT_NAME}"
