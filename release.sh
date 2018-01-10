#!/bin/bash
out="$1"; shift
root="$(cd "$(dirname "$0")"; pwd)"

if [[ -n $TRAVIS_TAG ]]; then
	version="$TRAVIS_TAG"
else
	cd "$root"
	version=$(git describe --tags --exact-match 2> /dev/null)
	if [[ -z $version ]]; then
		version=$(git rev-parse HEAD | cut -b-7)
	fi
	cd -
fi

name="hydroponics-${version}"

mkdir "$name"
for artifact in "$@"; do
	cp "$artifact" "$name"/
done
tar -czf "$out" "$name"
