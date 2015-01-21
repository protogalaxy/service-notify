#!/bin/bash

set -eu -o pipefail

PG_ROOT=$(dirname "${BASH_SOURCE}")/..
source "$PG_ROOT/build/common.sh"

function pg::test::find_dirs() {
  find . -not \( \
    -path ./target \
    -o -path '*/Godeps/*' \
  \) -name '*_test.go' -print0 | xargs -0n1 dirname | sort -u
}

mkdir -p "${TARGET_TEST_COVERAGE}"
readonly COVERAGE_FILE="${TARGET_TEST_COVERAGE}/coverage.out"

rm -f "${COVERAGE_FILE}"
echo "mode: count" > "${COVERAGE_FILE}"

for dir in `pg::test::find_dirs`; do
  profile="${dir}/coverage.out"
  godep go test -covermode=count -race -coverprofile="${profile}" "${dir}"
  if [[ -f $profile ]]; then
    cat "${profile}" | tail -n +2 >> ${COVERAGE_FILE}
    rm "${profile}"
  fi
done

go tool cover -html="${COVERAGE_FILE}" -o "${TARGET_TEST_COVERAGE}/coverage.html"
