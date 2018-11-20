#!/usr/bin/env groovy

@Library('zenreach@master') _

zenreachPipeline([
    name:       "hydroponics",
    build:      params.build,
    release:    params.release,
    build_prep_script: """
        #!/usr/bin/env bash
        set -e
    """,
    build_script: """
        #!/usr/bin/env bash
        set -e

        bazel build ...
        bazel test ...
	""",
    components: []
])
