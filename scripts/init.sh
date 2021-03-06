#!/bin/sh

set -eu

echo "Checking DB connection..."

i=0
until [ $i -ge 10 ]
do
  nc -z postgres 5432 && break

  i=$(( i + 1 ))

  echo "$i: Waiting for DB 3 second..."
  sleep 3
done

if [ $i -eq 10 ]
then
  echo "DB connection refused, terminating..."
  exit 1
fi

echo "DB is up..."

/bin/parser