#!/bin/bash
#
# ./hack/check-gofmt.sh checks that Go code is correctly formatted according to
# gofmt.

set -o errexit
set -o nounset
set -o pipefail

echo -n "Checking gofmt: "
ERRS=$(find "$@" -type f -name \*.go | xargs gofmt -l 2>&1 || true)
if [ -n "${ERRS}" ]; then
    echo "FAIL - the following files need to be gofmt'ed:"
    for e in ${ERRS}; do
        echo "    $e"
    done
    echo
    exit 1
fi
echo "PASS"
