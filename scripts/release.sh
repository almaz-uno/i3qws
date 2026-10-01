#!/bin/sh
# GitHub release on a version tag (specs/001-releases): i3qws-linux-amd64 built
# from a temporary worktree at the tag, SHA256SUMS, and as text the section of
# RELEASE-NOTES.adoc at the tag. The author pushes the tag; the script only
# checks it.
set -eu

tag=${1:-}
case $tag in
v[0-9]*.[0-9]*.[0-9]*) ;;
*) echo "usage: make release TAG=vX.Y.Z" >&2; exit 2 ;;
esac
version=${tag#v}

test "$(git cat-file -t "$tag" 2>/dev/null)" = tag ||
	{ echo "$tag is not an annotated tag" >&2; exit 1; }
git ls-remote --exit-code --tags origin "refs/tags/$tag" >/dev/null ||
	{ echo "tag $tag is not pushed to origin: git push origin $tag" >&2; exit 1; }
if gh release view "$tag" >/dev/null 2>&1; then
	echo "release $tag exists; to replace it: gh release delete $tag" >&2
	exit 1
fi

work=$(mktemp -d)
trap 'git worktree remove --force "$work/src" >/dev/null 2>&1 || true; rm -rf "$work"' EXIT INT TERM
git worktree add --quiet --detach "$work/src" "$tag"
make -s -C "$work/src" build

# The version of the build is exactly the tag: the worktree is clean, and
# git describe gives the tag
got=$("$work/src/i3qws" version | awk '{print $NF}')
test "$got" = "$tag" || { echo "the build reports version $got, not $tag" >&2; exit 1; }

mkdir "$work/dist"
cp "$work/src/i3qws" "$work/dist/i3qws-linux-amd64"
(cd "$work/dist" && sha256sum i3qws-linux-amd64 >SHA256SUMS)

url=$(gh repo view --json url -q .url)
git show "$tag:RELEASE-NOTES.adoc" >"$work/notes.adoc"
go run ./tools/relnotes "$work/notes.adoc" "$version" "$url/blob/$tag/" >"$work/notes.md"

gh release create "$tag" --verify-tag --title "i3qws $version" --notes-file "$work/notes.md" "$work/dist"/*
