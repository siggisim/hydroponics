workspace(
    name = "com_zenreach_hydroponics",
)

rules_go_version = "985b08a07c3fe8c3a305bae8204da4e8c13fe17d"

rules_go_sha = "f599a0ec2149b440a48bbab3240b303d15b5a48175ed76a2db15064e0202e36a"

http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/archive/%s.tar.gz" % rules_go_version,
    strip_prefix = "rules_go-%s" % rules_go_version,
    sha256 = rules_go_sha,
)

load("@io_bazel_rules_go//go:def.bzl", "go_register_toolchains", "go_repository", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

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
