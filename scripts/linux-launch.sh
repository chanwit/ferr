#!/bin/sh

DIR="$(cd "$(dirname "$0")" ; pwd -P)"
FERR_EXE=$DIR/bin/ferr.bin
export PATH=$DIR/bin:$PATH
exec "$FERR_EXE" $@
