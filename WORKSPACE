workspace(
    name = "com_zenreach_hydroponics",
)

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    urls = ["https://github.com/bazelbuild/rules_go/releases/download/0.16.0/rules_go-0.16.0.tar.gz"],
    sha256 = "ee5fe78fe417c685ecb77a0a725dc9f6040ae5beb44a0ba4ddb55453aad23a8a",
)

load("@io_bazel_rules_go//go:def.bzl", "go_register_toolchains", "go_rules_dependencies")
go_rules_dependencies()
go_register_toolchains()

http_archive(
    name = "bazel_gazelle",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/releases/download/0.14.0/bazel-gazelle-0.14.0.tar.gz"],
    sha256 = "c0a5739d12c6d05b6c1ad56f2200cb0b57c5a70e03ebd2f7b87ce88cabf09c7b",
)
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
gazelle_dependencies()

go_repository(
    name = "com_github_aws_aws_sdk_go",
    commit = "ae53883b2478fd8e2bdca2748367d5b5fa27ca22",
    importpath = "github.com/aws/aws-sdk-go",
)

go_repository(
    name = "com_github_caarlos0_env",
    commit = "0cf029d5748c52beb2c9d20c81880cb4bdf8f788",
    importpath = "github.com/caarlos0/env",
)

go_repository(
    name = "com_github_eawsy_aws_lambda_go_core",
    commit = "e26eed6aa244a3d45aa693816a9c5faf39390fcd",
    importpath = "github.com/eawsy/aws-lambda-go-core",
)

go_repository(
    name = "com_github_eawsy_aws_lambda_go_event",
    commit = "e888a5ec6428554de39d49d6eda94f60027cfb81",
    importpath = "github.com/eawsy/aws-lambda-go-event",
)

go_repository(
    name = "com_github_go_ini_ini",
    commit = "32e4c1e6bc4e7d0d8451aa6b75200d19e37a536a",
    importpath = "github.com/go-ini/ini",
)

go_repository(
    name = "com_github_jmespath_go_jmespath",
    commit = "3433f3ea46d9f8019119e7dd41274e112a2359a9",
    importpath = "github.com/jmespath/go-jmespath",
)

go_repository(
    name = "com_github_pkg_errors",
    commit = "c605e284fe17294bda444b34710735b29d1a9d90",
    importpath = "github.com/pkg/errors",
)

go_repository(
    name = "com_github_zenreach_hatchet",
    commit = "a3c0cee4131339eda3c6b75cd0097c0b45b20bac",
    importpath = "github.com/zenreach/hatchet",
)

go_repository(
    name = "com_github_golang_groupcache",
    commit = "84a468cf14b4376def5d68c722b139b881c450a4",
    importpath = "github.com/golang/groupcache",
)
