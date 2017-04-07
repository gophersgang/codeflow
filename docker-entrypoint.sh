#!/bin/sh
WORKDIR=${WORKDIR-.}
cd $WORKDIR
exec "$@"
