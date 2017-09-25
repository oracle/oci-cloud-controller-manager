#!/bin/bash
#
# ./hack/check-golint.sh checks the coding style of our Go code.

set -o errexit
set -o nounset
set -o pipefail

TARGETS=$(for d in "$@"; do echo ./$d/...; done)

echo -n "Checking golint: "
ERRS=$(golint ${TARGETS} 2>&1 || true)
if [ -n "${ERRS}" ]; then
    echo "FAIL"
    echo "${ERRS}"
    echo
    exit 1
fi
echo "PASS"
