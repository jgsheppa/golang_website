#!/bin/sh

# Abort on any error (including if wait-for-it fails).
set -e

# Wait for the backend to be up, if we know where it is.
if [ -n "$db" ]; then
  /usr/src/app/wait-for-it.sh "$db:${db:-5432}"
fi

echo "working"

# Run the main container command.
exec "$@"
