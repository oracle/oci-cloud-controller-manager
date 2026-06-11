#!/usr/bin/env bash

# Copyright 2026 Oracle and/or its affiliates. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

VERSION_FILE=${VERSION_FILE:-VERSION}
KUBERNETES_MINOR=${KUBERNETES_MINOR:-}
KUBERNETES_PATCH_VERSION=${KUBERNETES_PATCH_VERSION:-}
UPDATE_VENDOR=${UPDATE_VENDOR:-true}
RUN_GOVULNCHECK=${RUN_GOVULNCHECK:-true}
INSTALL_GOVULNCHECK=${INSTALL_GOVULNCHECK:-true}
GOVULNCHECK_BIN=${GOVULNCHECK_BIN:-}
GOVULNCHECK_SCAN_LEVEL=${GOVULNCHECK_SCAN_LEVEL:-package}
DRY_RUN=false

function usage() {
    cat <<EOF
Usage: $0 [flags]

Updates Kubernetes release-module dependencies to the latest patch for the
minor version in VERSION, without changing the go directive in go.mod.

Flags:
  --dry-run            Print the commands that would run.
  --no-vendor          Do not run go mod vendor after dependency updates.
  --skip-govulncheck   Do not run govulncheck after dependency updates.
  --no-install-govulncheck
                       Do not install govulncheck automatically when missing.
  --govulncheck-scan-level LEVEL
                       Use govulncheck scan LEVEL: module, package, or symbol.
  --version-file FILE  Read the Kubernetes minor from FILE instead of VERSION.
  --minor VERSION      Use this Kubernetes minor instead of VERSION, e.g. 1.35.
  --patch PATCH        Use this patch number instead of discovering latest.
  -h, --help           Show this help.

Environment:
  VERSION_FILE                 Same as --version-file.
  KUBERNETES_MINOR             Same as --minor.
  KUBERNETES_PATCH_VERSION     Same as --patch.
  UPDATE_VENDOR=false          Same as --no-vendor.
  RUN_GOVULNCHECK=false        Same as --skip-govulncheck.
  INSTALL_GOVULNCHECK=false    Same as --no-install-govulncheck.
  GOVULNCHECK_BIN=/path/bin    Use this govulncheck binary.
  GOVULNCHECK_SCAN_LEVEL       Same as --govulncheck-scan-level.
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --no-vendor)
            UPDATE_VENDOR=false
            shift
            ;;
        --skip-govulncheck)
            RUN_GOVULNCHECK=false
            shift
            ;;
        --no-install-govulncheck)
            INSTALL_GOVULNCHECK=false
            shift
            ;;
        --govulncheck-scan-level)
            GOVULNCHECK_SCAN_LEVEL=$2
            shift 2
            ;;
        --version-file)
            VERSION_FILE=$2
            shift 2
            ;;
        --minor)
            KUBERNETES_MINOR=$2
            shift 2
            ;;
        --patch)
            KUBERNETES_PATCH_VERSION=$2
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "unknown argument: $1" >&2
            usage >&2
            exit 2
            ;;
    esac
done

function log() {
    echo "==> $*"
}

function die() {
    echo "ERROR: $*" >&2
    exit 1
}

function run() {
    echo "+ $*"
    if [[ "$DRY_RUN" != "true" ]]; then
        "$@"
    fi
}

function require_cmd() {
    command -v "$1" >/dev/null 2>&1 || die "'$1' is required"
}

function validate_govulncheck_scan_level() {
    case "$GOVULNCHECK_SCAN_LEVEL" in
        module|package|symbol)
            ;;
        *)
            die "invalid govulncheck scan level '${GOVULNCHECK_SCAN_LEVEL}'; expected one of: module, package, symbol"
            ;;
    esac
}

function repo_root() {
    git rev-parse --show-toplevel 2>/dev/null || pwd
}

function go_mod_flag() {
    env GOFLAGS=-mod=mod "$@"
}

function module_versions_without_replacements() {
    local module=$1
    local tmpdir
    local output
    local status

    tmpdir=$(mktemp -d "${TMPDIR:-/tmp}/k8s-module-versions.XXXXXX")
    set +o errexit
    output=$(
        cd "$tmpdir"
        env GOFLAGS= GOPATH="$tmpdir/gopath" GOMODCACHE="$tmpdir/gomodcache" GOCACHE="$tmpdir/gocache" go list -m -versions "$module"
    )
    status=$?
    set -o errexit

    rm -rf "$tmpdir"
    if [[ "$status" -ne 0 ]]; then
        return "$status"
    fi

    printf '%s\n' "$output"
}

function read_go_directive() {
    awk '$1 == "go" { print $2; exit }' go.mod
}

function normalize_minor() {
    local raw=${1#v}
    local major
    local minor

    IFS=. read -r major minor _ <<<"$raw"
    [[ "$major" =~ ^[0-9]+$ ]] || die "invalid Kubernetes version '$1'"
    [[ "$minor" =~ ^[0-9]+$ ]] || die "invalid Kubernetes version '$1'"
    [[ "$major" == "1" ]] || die "only Kubernetes 1.x versions are supported, got '$1'"

    echo "${major}.${minor}"
}

function version_from_file() {
    [[ -f "$VERSION_FILE" ]] || die "version file '$VERSION_FILE' does not exist"
    tr -d '[:space:]' < "$VERSION_FILE"
}

function latest_patch_for_minor() {
    local minor=$1
    local module_minor=${minor#*.}
    local versions
    local patch

    if [[ -n "$KUBERNETES_PATCH_VERSION" ]]; then
        [[ "$KUBERNETES_PATCH_VERSION" =~ ^[0-9]+$ ]] || die "invalid patch '$KUBERNETES_PATCH_VERSION'"
        echo "$KUBERNETES_PATCH_VERSION"
        return
    fi

    versions=$(module_versions_without_replacements k8s.io/api)
    patch=$(printf '%s\n' $versions | awk -v prefix="v0.${module_minor}." '
        index($0, prefix) == 1 {
            value = substr($0, length(prefix) + 1)
            if (value ~ /^[0-9]+$/ && value + 0 > max) {
                max = value + 0
                found = 1
            }
        }
        END {
            if (!found) {
                exit 1
            }
            print max
        }
    ') || die "could not find a k8s.io/api patch for Kubernetes ${minor}"

    echo "$patch"
}

function is_kubernetes_release_module() {
    case "$1" in
        k8s.io/klog|k8s.io/klog/v2|k8s.io/kube-openapi|k8s.io/utils|k8s.io/gengo|k8s.io/gengo/*|k8s.io/system-validators)
            return 1
            ;;
        k8s.io/*)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

function module_version_for() {
    local module=$1
    local kubernetes_version=$2
    local staging_version=$3

    if [[ "$module" == "k8s.io/kubernetes" ]]; then
        echo "$kubernetes_version"
    else
        echo "$staging_version"
    fi
}

function replace_lines() {
    awk '$2 == "=>" && $1 ~ /^k8s\.io\// { print $1, $3 }' go.mod
}

function required_kubernetes_modules() {
    awk '$1 ~ /^k8s\.io\// && $2 ~ /^v[0-9]/ { print $1 }' go.mod | sort -u
}

function direct_non_kubernetes_module_versions() {
    awk '
        $1 == "require" && $2 !~ /^\(/ && $3 ~ /^v[0-9]/ && $0 !~ /\/\/ indirect/ {
            if ($2 !~ /^k8s\.io\//) {
                print $2, $3
            }
            next
        }
        $1 !~ /^(module|go|toolchain|require|\)|replace)$/ && $2 ~ /^v[0-9]/ && $0 !~ /\/\/ indirect/ {
            if ($1 !~ /^k8s\.io\//) {
                print $1, $2
            }
        }
    ' go.mod | sort -u
}

function semver_sort_key() {
    local version=${1#v}
    local major
    local minor
    local patch

    version=${version%%+*}
    version=${version%%-*}
    IFS=. read -r major minor patch _ <<<"$version"
    patch=${patch:-0}

    [[ "$major" =~ ^[0-9]+$ ]] || return 1
    [[ "$minor" =~ ^[0-9]+$ ]] || return 1
    [[ "$patch" =~ ^[0-9]+$ ]] || return 1

    printf '%08d%08d%08d\n' "$major" "$minor" "$patch"
}

function version_is_downgrade() {
    local before=$1
    local after=$2
    local before_key
    local after_key

    before_key=$(semver_sort_key "$before") || return 1
    after_key=$(semver_sort_key "$after") || return 1

    [[ "$after_key" < "$before_key" ]]
}

function module_version_from_snapshot() {
    local snapshot=$1
    local module=$2

    awk -v module="$module" '$1 == module { print $2; exit }' <<<"$snapshot"
}

function report_direct_non_kubernetes_downgrades() {
    local before_snapshot=$1
    local after_snapshot=$2
    local kubernetes_minor=$3
    local module
    local before_version
    local after_version
    local found=false

    while read -r module before_version; do
        [[ -n "${module:-}" ]] || continue

        after_version=$(module_version_from_snapshot "$after_snapshot" "$module")
        if [[ -n "$after_version" && "$after_version" != "$before_version" ]] && version_is_downgrade "$before_version" "$after_version"; then
            if [[ "$found" != "true" ]]; then
                echo
                echo "Go downgraded direct non-Kubernetes modules while resolving Kubernetes ${kubernetes_minor}:"
                found=true
            fi
            echo "  ${module}: ${before_version} => ${after_version}"
        fi
    done <<<"$before_snapshot"

    if [[ "$found" == "true" ]]; then
        echo
        echo "This script pins Kubernetes release modules to the ${kubernetes_minor} minor from VERSION."
        echo "Those downgrades usually mean the previous module versions require a different Kubernetes minor."
        echo "Review the module requirements before accepting the compatible versions selected by Go."
    fi
}

function restore_go_directive() {
    local original_go=$1
    local current_go

    current_go=$(read_go_directive)
    if [[ "$current_go" != "$original_go" ]]; then
        log "Restoring go directive to ${original_go}; Go upgrades are reported but not applied."
        run go mod edit "-go=${original_go}"
    fi

    if grep -q '^toolchain ' go.mod; then
        log "Removing toolchain directive; Go upgrades are reported but not applied."
        run go mod edit -toolchain=none
    fi
}

function report_golang_upgrade_suggestion() {
    local scan_output=$1

    if printf '%s\n' "$scan_output" | grep -Eq 'Fixed in: .*@go[0-9]'; then
        echo
        echo "Go toolchain vulnerabilities remain. Upgrade Go outside this script, then rerun the scan."
        printf '%s\n' "$scan_output" | grep -E 'Found in: .*@go|Fixed in: .*@go' | sed 's/^/  /'
    fi
}

function govulncheck_module_scan_dir() {
    local preferred_dir=cmd/oci-cloud-controller-manager
    local fallback_dir=""
    local file
    local dir

    if [[ -n "$(find "$preferred_dir" -maxdepth 1 -type f -name '*.go' -print -quit 2>/dev/null)" ]]; then
        echo "$preferred_dir"
        return
    fi

    while IFS= read -r file; do
        [[ -n "$file" ]] || continue

        dir=${file%/*}
        if [[ "$dir" == "$file" ]]; then
            dir=.
        fi

        if [[ -z "$fallback_dir" ]]; then
            fallback_dir=$dir
        fi

        case "$dir" in
            hack|hack/*|test|test/*)
                ;;
            *)
                echo "$dir"
                return
                ;;
        esac
    done < <(git ls-files '*.go' ':!:vendor/**')

    [[ -n "$fallback_dir" ]] || die "could not find a Go package directory for govulncheck module scan"
    echo "$fallback_dir"
}

function go_install_bin_dir() {
    local gobin
    local gopath
    local first_gopath

    gobin=$(go env GOBIN)
    if [[ -n "$gobin" ]]; then
        echo "$gobin"
        return
    fi

    gopath=$(go env GOPATH)
    IFS=: read -r first_gopath _ <<<"$gopath"
    [[ -n "$first_gopath" ]] || die "could not determine GOPATH for govulncheck install"

    echo "${first_gopath}/bin"
}

function resolve_govulncheck_bin() {
    local gobin_path
    local install_path

    if [[ -n "$GOVULNCHECK_BIN" ]]; then
        echo "$GOVULNCHECK_BIN"
        return
    fi

    if gobin_path=$(command -v govulncheck 2>/dev/null); then
        echo "$gobin_path"
        return
    fi

    install_path="$(go_install_bin_dir)/govulncheck"
    if [[ -x "$install_path" ]]; then
        echo "$install_path"
        return
    fi

    return 1
}

function install_govulncheck() {
    local install_output
    local install_path
    local status

    if [[ "$INSTALL_GOVULNCHECK" != "true" ]]; then
        echo "govulncheck not found; install it with:"
        echo "  go install golang.org/x/vuln/cmd/govulncheck@latest"
        return 1
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        run env GOTOOLCHAIN=local go install golang.org/x/vuln/cmd/govulncheck@latest
        return 1
    fi

    log "govulncheck not found; installing golang.org/x/vuln/cmd/govulncheck@latest."
    set +o errexit
    install_output=$(env GOTOOLCHAIN=local go install golang.org/x/vuln/cmd/govulncheck@latest 2>&1)
    status=$?
    set -o errexit

    if [[ -n "$install_output" ]]; then
        printf '%s\n' "$install_output"
    fi

    if [[ "$status" -ne 0 ]]; then
        echo "Could not install govulncheck with the current Go toolchain." >&2
        echo "If the output above says a newer Go version is required, upgrade Go outside this script and rerun." >&2
        return 1
    fi

    install_path="$(go_install_bin_dir)/govulncheck"
    if [[ -x "$install_path" ]]; then
        GOVULNCHECK_BIN="$install_path"
        return 0
    fi

    if GOVULNCHECK_BIN=$(command -v govulncheck 2>/dev/null); then
        return 0
    fi

    echo "govulncheck was installed, but the binary could not be found at ${install_path}." >&2
    echo "Add the Go install bin directory to PATH or set GOVULNCHECK_BIN and rerun." >&2
    return 1
}

function run_govulncheck() {
    local govulncheck_bin
    local module_scan_dir
    local scan_output
    local status

    if [[ "$RUN_GOVULNCHECK" != "true" ]]; then
        log "Skipping govulncheck."
        return
    fi

    if [[ -n "$GOVULNCHECK_BIN" && ! -x "$GOVULNCHECK_BIN" ]]; then
        die "GOVULNCHECK_BIN '$GOVULNCHECK_BIN' is not executable"
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        if ! resolve_govulncheck_bin >/dev/null; then
            install_govulncheck || true
        fi
        log "Skipping govulncheck in dry-run mode."
        return
    fi

    if ! govulncheck_bin=$(resolve_govulncheck_bin); then
        if ! install_govulncheck; then
            echo "Skipping govulncheck ${GOVULNCHECK_SCAN_LEVEL} scan."
            return
        fi
        govulncheck_bin=$GOVULNCHECK_BIN
    fi

    if [[ -z "$govulncheck_bin" ]]; then
        echo "Skipping govulncheck ${GOVULNCHECK_SCAN_LEVEL} scan because no govulncheck binary was resolved."
        return
    fi

    if [[ "$GOVULNCHECK_SCAN_LEVEL" == "module" ]]; then
        module_scan_dir=$(govulncheck_module_scan_dir)
        log "Running govulncheck ${GOVULNCHECK_SCAN_LEVEL} scan from ${module_scan_dir}."
    else
        log "Running govulncheck ${GOVULNCHECK_SCAN_LEVEL} scan."
    fi

    set +o errexit
    if [[ "$GOVULNCHECK_SCAN_LEVEL" == "module" ]]; then
        scan_output=$(go_mod_flag "$govulncheck_bin" -C "$module_scan_dir" "-scan=${GOVULNCHECK_SCAN_LEVEL}" 2>&1)
    else
        scan_output=$(go_mod_flag "$govulncheck_bin" "-scan=${GOVULNCHECK_SCAN_LEVEL}" ./... 2>&1)
    fi
    status=$?
    set -o errexit

    printf '%s\n' "$scan_output"
    report_golang_upgrade_suggestion "$scan_output"

    if [[ "$status" -ne 0 ]]; then
        echo "govulncheck exited with status ${status}; review any remaining findings above."
    fi
}

require_cmd awk
require_cmd git
require_cmd go
validate_govulncheck_scan_level

cd "$(repo_root)"

[[ -f go.mod ]] || die "go.mod not found"

raw_minor=${KUBERNETES_MINOR:-$(version_from_file)}
kubernetes_minor=$(normalize_minor "$raw_minor")
patch=$(latest_patch_for_minor "$kubernetes_minor")
module_minor=${kubernetes_minor#*.}
staging_version="v0.${module_minor}.${patch}"
kubernetes_version="v${kubernetes_minor}.${patch}"
original_go=$(read_go_directive)
direct_non_kubernetes_before=$(direct_non_kubernetes_module_versions)
updates=()

log "Kubernetes minor: ${kubernetes_minor}"
log "Latest patch: ${patch}"
log "Kubernetes module version: ${kubernetes_version}"
log "Staging module version: ${staging_version}"
log "Preserving go directive: ${original_go}"

while read -r module replacement; do
    if is_kubernetes_release_module "$module"; then
        version=$(module_version_for "$module" "$kubernetes_version" "$staging_version")
        run go mod edit "-replace=${module}=${replacement}@${version}"
    fi
done < <(replace_lines)

while read -r module; do
    if is_kubernetes_release_module "$module"; then
        version=$(module_version_for "$module" "$kubernetes_version" "$staging_version")
        updates+=("${module}@${version}")
    fi
done < <(required_kubernetes_modules)

if [[ "${#updates[@]}" -gt 0 ]]; then
    run go get "${updates[@]}"
    direct_non_kubernetes_after=$(direct_non_kubernetes_module_versions)
    report_direct_non_kubernetes_downgrades "$direct_non_kubernetes_before" "$direct_non_kubernetes_after" "$kubernetes_minor"
else
    log "No Kubernetes release modules found in require directives."
fi

run go mod tidy
restore_go_directive "$original_go"

if [[ "$UPDATE_VENDOR" == "true" && -d vendor ]]; then
    run go mod vendor
else
    log "Skipping go mod vendor."
fi

run_govulncheck
