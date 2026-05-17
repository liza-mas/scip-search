#!/bin/sh
set -eu

REPO_OWNER=${SCIP_SEARCH_REPO_OWNER:-liza-mas}
REPO_NAME=${SCIP_SEARCH_REPO_NAME:-scip-search}
INSTALL_DIR=${INSTALL_DIR:-"$HOME/.local/bin"}
RELEASES_FILE=${SCIP_SEARCH_RELEASES_FILE:-}
if [ -n "${SCIP_SEARCH_RELEASES_URL:-}" ]; then
	RELEASES_URL=$SCIP_SEARCH_RELEASES_URL
elif [ -n "${VERSION:-}" ]; then
	RELEASES_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/releases.tsv"
else
	RELEASES_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/releases.tsv"
fi

fail() {
	printf 'scip-search install: %s\n' "$*" >&2
	exit 1
}

normalize_os() {
	case "$1" in
	Linux | linux)
		printf 'linux\n'
		;;
	Darwin | darwin)
		printf 'darwin\n'
		;;
	*)
		printf '%s\n' "$1"
		;;
	esac
}

normalize_arch() {
	case "$1" in
	x86_64 | amd64)
		printf 'amd64\n'
		;;
	arm64 | aarch64)
		printf 'arm64\n'
		;;
	*)
		printf '%s\n' "$1"
		;;
	esac
}

detect_platform() {
	raw_os=${SCIP_SEARCH_INSTALL_OS:-}
	raw_arch=${SCIP_SEARCH_INSTALL_ARCH:-}

	if [ -z "$raw_os" ]; then
		raw_os=$(uname -s)
	fi
	if [ -z "$raw_arch" ]; then
		raw_arch=$(uname -m)
	fi

	platform_os=$(normalize_os "$raw_os")
	platform_arch=$(normalize_arch "$raw_arch")

	case "$platform_os/$platform_arch" in
	linux/amd64 | linux/arm64 | darwin/amd64 | darwin/arm64)
		return 0
		;;
	*)
		fail "unsupported platform $platform_os/$platform_arch"
		;;
	esac
}

metadata_source() {
	if [ -n "$RELEASES_FILE" ]; then
		[ -r "$RELEASES_FILE" ] || fail "release metadata file is not readable: $RELEASES_FILE"
		printf '%s\n' "$RELEASES_FILE"
		return 0
	fi

	tmp_file=$(mktemp)
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$RELEASES_URL" -o "$tmp_file" || fail "latest release metadata unavailable at $RELEASES_URL"
	else
		fail "curl is required to fetch release metadata unless SCIP_SEARCH_RELEASES_FILE is set"
	fi
	printf '%s\n' "$tmp_file"
}

select_release() {
	metadata=$1
	requested=${VERSION:-}
	best_version=
	best_published=
	best_artifact=

	while IFS='	' read -r release_version published_at release_os release_arch artifact_path; do
		[ -n "$release_version" ] || continue
		[ "$release_os" = "$platform_os" ] || continue
		[ "$release_arch" = "$platform_arch" ] || continue

		if [ -n "$requested" ]; then
			if [ "$release_version" = "$requested" ]; then
				best_version=$release_version
				best_artifact=$artifact_path
				break
			fi
			continue
		fi

		if [ -z "$best_published" ] || [ "$published_at" \> "$best_published" ]; then
			best_version=$release_version
			best_published=$published_at
			best_artifact=$artifact_path
		fi
	done <"$metadata"

	if [ -z "$best_version" ]; then
		if [ -n "$requested" ]; then
			fail "VERSION=$requested is unavailable for platform $platform_os/$platform_arch"
		fi
		fail "latest release is unavailable for platform $platform_os/$platform_arch"
	fi

	selected_version=$best_version
	selected_artifact=$best_artifact
}

install_artifact() {
	dest=$INSTALL_DIR/scip-search
	mkdir -p "$INSTALL_DIR" || fail "INSTALL_DIR is not usable: $INSTALL_DIR"

	case "$selected_artifact" in
	http://* | https://*)
		if command -v curl >/dev/null 2>&1; then
			curl -fsSL "$selected_artifact" -o "$dest" || fail "release artifact is unavailable for $selected_version"
		else
			fail "curl is required to fetch release artifacts"
		fi
		;;
	*)
		[ -r "$selected_artifact" ] || fail "release artifact is unavailable for $selected_version: $selected_artifact"
		cp "$selected_artifact" "$dest" || fail "INSTALL_DIR is not usable: $INSTALL_DIR"
		;;
	esac

	chmod 0755 "$dest" || fail "INSTALL_DIR is not usable: $INSTALL_DIR"
	printf 'Installed scip-search %s to %s\n' "$selected_version" "$dest"
}

detect_platform
metadata=$(metadata_source)
select_release "$metadata"
install_artifact
