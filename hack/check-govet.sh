#!/bin/bash
#
# ./hack/check-govet.sh checks that go vet reports no suspicious constructs.

set -o errexit
set -o nounset
set -o pipefail

TARGETS=$(for d in "$@"; do echo ./$d/...; done)

echo -n "Checking go vet: "
ERRS=$(go vet ${TARGETS} 2>&1 || true)
if [ -n "${ERRS}" ]; then
    echo "FAIL"
    echo "${ERRS}"
    echo
    exit 1
fi
echo "PASS"
