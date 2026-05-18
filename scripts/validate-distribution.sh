#!/bin/sh
set -u

GO_CMD=${GO:-go}

run_distribution_check() {
	category=$1
	shift

	printf 'distribution validation: %s\n' "$category"
	"$@"
	status=$?
	if [ "$status" -eq 0 ]; then
		return 0
	fi

	printf 'distribution validation failed: %s\n' "$category" >&2
	return "$status"
}

run_distribution_check "docs drift" "$GO_CMD" test ./tests/docs || exit $?
run_distribution_check "version smoke" "$GO_CMD" test ./tests/e2e -run '^TestVersionExecutableSmokeReleaseAndSourceBuildsWithoutIndex$' || exit $?
run_distribution_check "release install packaging" "$GO_CMD" test ./tests/e2e -run '^TestReleaseInstallerEndToEnd' || exit $?
run_distribution_check "source install packaging" "$GO_CMD" test ./tests/e2e -run '^TestSourceBranchInstallEndToEnd' || exit $?
run_distribution_check "local clone source install" "$GO_CMD" test ./tests/install -run '^TestMakeInstall' || exit $?
