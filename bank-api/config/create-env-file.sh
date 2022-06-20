#!/bin/sh

basedir=$(dirname "$0")
envfile="$basedir/env/.env"
environment="$1"

rm -f "$envfile"
touch "$envfile"

for file in $(find "$basedir" -type f -iname "*.$environment.env.example" | sort); do
  cat "$file" >> "$envfile"
  echo >> "$envfile"
done

echo "$envfile" created
