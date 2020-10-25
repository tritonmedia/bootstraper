#!/usr/bin/env bash
#
# Run a golang binary using gobin
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

# shellcheck source=./lib/logging.sh
source "$DIR/lib/logging.sh"

get_root_module() {
  # shellcheck disable=SC2155
  local modules=$(go list -m -f '{{ if not .Indirect }}{{ .Path }}{{ end }}' all)
  IFS='/' read -r -a paths <<<"$1"

  local matching=""
  local i=0
  while [[ $i -ne ${#paths[@]} ]]; do
    unset "paths[-1]"
    local path=""
    for elem in "${paths[@]}"; do
      path+="/$elem"
    done

    # If we're at root, then stop
    if [[ $path == "" ]] || [[ $path == "/" ]]; then
      break
    fi

    path="$(sed 's/\///' <<<"$path")"
    if grep -q "$path" <<<"$modules"; then
      matching=$(grep "$path" <<<"$modules")
      break
    fi

    i=$((i + 1))
  done

  # If we have multiple matches, just stop
  echo "$matching" | head -n1
}

GOBINVERSION=v0.0.14
GOBINPATH="/usr/local/bin/gobin"
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

PRINT_PATH=false
if [[ $1 == "-p" ]]; then
  PRINT_PATH=true
  shift
fi

if [[ -z $1 ]] || [[ $1 =~ ^(--help|-h)$ ]]; then
  echo "Usage: $0 [-p|-h|--help] <package> [args]" >&2
  exit 1
fi

# Install a global version of gobin, if neccessary.
if ! command -v gobin >/dev/null 2>&1; then
  mkdir -p "$(dirname "$GOBINPATH")"
  info "installing gobin into '$GOBINPATH'" >&2
  curl -L -o "/tmp/gobin" "https://github.com/myitcv/gobin/releases/download/$GOBINVERSION/$GOOS-$GOARCH" >&2
  chmod +x "/tmp/gobin"

  if [[ ! -w $GOBINPATH ]]; then
    sudo mv "/tmp/gobin" "$GOBINPATH"
  else
    mv /tmp/gobin "$GOBINPATH"
  fi
fi

package="$1"
shift

# If we don't have a hardcoded version, attempt to look it
# up in go.mod and .tool-versions (asdf compatible)
if ! grep "@" <<<"$package" >/dev/null 2>&1; then
  # First find the version inside of our go.mod
  possibleModuleName=$(get_root_module "$package")
  goModVersion=$(go list -m -f '{{ .Version }}' "$possibleModuleName" 2>/dev/null || true)
  if [[ -n $goModVersion ]]; then
    version="$goModVersion"
  else
    # If not in go.mod, look up the version inside of .tool-versions
    if [[ -e "$DIR/../.tool-versions" ]]; then
      version=$(grep "$package" "$DIR/../.tool-versions" | awk '{ print $2 }')
      if [[ -n $version ]]; then
        package="$package@$version"
      else
        error "failed to find version of tool '$package'"
      fi
    fi
  fi
fi

if [[ $PRINT_PATH == "true" ]]; then
  exec "$GOBINPATH" -p "$package"
fi

exec "$GOBINPATH" -run "$package" "$@"
