#!/bin/bash

set -eu -o pipefail

PG_ROOT=$(dirname "${BASH_SOURCE}")/..
source "$PG_ROOT/build/common.sh"

mkdir -p "${TARGET_TEST_COVERAGE}"
readonly COVERAGE_FILE="${TARGET_TEST_COVERAGE}/coverage.out"

rm -f "${COVERAGE_FILE}"
echo "mode: count" > "${COVERAGE_FILE}"

for dir in `pg::golang::find_test_dirs`; do
  profile="${dir}/coverage.out"
  godep go test -covermode=count -race -coverprofile="${profile}" "${dir}"
  if [[ -f $profile ]]; then
    cat "${profile}" | tail -n +2 >> ${COVERAGE_FILE}
    rm "${profile}"
  fi
done

go tool cover -html="${COVERAGE_FILE}" -o "${TARGET_TEST_COVERAGE}/coverage.html"
