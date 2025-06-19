#!/bin/sh
# Wait for the PostgreSQL database to be available before starting the service

set -e

host="$1"
port="$2"
shift 2
cmd="$@"

echo "Waiting for $host:$port to be available..."

while ! nc -z "$host" "$port"; do
  sleep 1
done

echo "$host:$port is available. Starting service..."
exec $cmd
