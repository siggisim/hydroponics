#!/bin/sh

BAZEL_OS=linux
BAZEL_VERSION=0.9.0
BAZEL_ARGS="--output_base=$HOME/.cache/bazel --batch --host_jvm_args=-Xmx500m --host_jvm_args=-Xms500m"
BAZEL_TEST_ARGS="--config=ci --local_resources=400,1,1.0"
BAZEL_BUILD_ARGS="--config=ci --local_resources=400,1,1.0"

version() {
	if [[ -n $TRAVIS_TAG ]]; then
		version="$TRAVIS_TAG"
	else
		version=$(git describe --tags --exact-match 2> /dev/null)
		if [[ -z $version ]]; then
			version=$(git rev-parse HEAD | cut -b-7)
		fi
	fi
	echo $version
}

install_bazel() {
	wget -o install.sh \
		"https://github.com/bazelbuild/bazel/releases/download/${BAZEL_VERSION}/bazel-${BAZEL_VERSION}-installer-${BAZEL_OS}-x86_64.sh" || exit 1
	chmod +x install.sh
	./install.sh || exit 1
	rm install.sh
}

test_all() {
	bazel $BAZEL_ARGS test $BAZEL_TEST_ARGS //... || exit 1
}

build_release() {
	bazel $BAZEL_ARGS build $BAZEL_BUILD_ARGS //:release || exit 1
	mv bazel-bin/hydroponics.tar.gz "hydroponics-$(version).tar.gz"
}

case in "$1"
install)
	install_bazel
	;;
test)
	test_all
	;;
build)
	build_release
	;;
*)
	echo "usage: $0 [install|test|build]" >&2
	exit 1
	;;
esac
