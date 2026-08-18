#!/bin/bash -u

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"

# Protobuf definitions
PROTO=${1:-"$ROOT/proto"}

function main() {
  checks

  set -e
  pushd "$ROOT" >/dev/null

  generate_proto

  popd >/dev/null

  echo "generate.sh - `date` - `whoami`" > $ROOT/pb/last_generate.txt
  echo "streamingfast/dregistry revision: `GIT_DIR=$ROOT/.git git rev-parse HEAD`" >> $ROOT/pb/last_generate.txt

  echo "Done"
}

function generate_proto() {
  echo "Generating dregistry Protobuf bindings via 'buf'"
  buf generate proto
}

function checks() {
  result=`printf "" | buf --version 2>&1 | grep -Eo '1\.(1[0-9]+|[2-9][0-9]+)\.'`
  if [[ "$result" == "" ]]; then
    echo "The 'buf' binary is either missing or is not recent enough (at `which buf || echo N/A`)."
    echo ""
    echo "To fix your problem, on Mac OS, perform this command:"
    echo ""
    echo "  brew install bufbuild/buf/buf"
    echo ""
    echo "On other system, refers to https://docs.buf.build/installation"
    echo ""
    echo "If everything is working as expected, the command:"
    echo ""
    echo "  buf --version"
    echo ""
    echo "Should print '1.11.0' (or newer)"
    exit 1
  fi
}

main "$@"
