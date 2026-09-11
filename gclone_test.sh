#!/usr/bin/env bash
# Smoke tests for gclone (pure bash port). Mirrors main_test.go vectors.
# Usage: ./gclone_test.sh
cd "$(dirname "$0")" || exit 1
fail=0

# Stub git: records target dir, creates it, succeeds.
FAKEBIN="$(mktemp -d)"
trap 'rm -rf "$FAKEBIN" "$WS"' EXIT
printf '#!/usr/bin/env bash\nmkdir -p "${@: -1}"\nexit 0\n' >"$FAKEBIN/git"
chmod +x "$FAKEBIN/git"
export WS
WS="$(mktemp -d)"
export WORKSPACE="$WS"

expect_dir() { # <url> <expected-suffix>
  local out
  out=$(PATH="$FAKEBIN:$PATH" ./gclone "$1" 2>&1)
  case "$out" in *"$2"*) echo "ok: $1" ;; *)
    echo "FAIL: $1 -> $out"; fail=1 ;;
  esac
}

expect_reject() { # <url>
  if PATH="$FAKEBIN:$PATH" ./gclone "$1" 2>&1 | grep -q "error parsing"; then
    echo "ok reject: $1"
  else
    echo "FAIL accept: $1"; fail=1
  fi
}

while IFS='|' read -r url expected; do
  expect_dir "$url" "$expected"
done <<'EOF'
https://github.com/Owner/Repo.git|github.com/Owner/Repo
git@github.com:Owner/Repo.git|github.com/Owner/Repo
ssh://git@example.com:2222/team/repo.git|example.com/team/repo
git://github.com/owner/repo.git|github.com/owner/repo
https://gitlab.com/group/subgroup/repo.git|gitlab.com/group/subgroup/repo
file:///tmp/repo|local/tmp/repo
file:///C:/tmp/repo|local/C/tmp/repo
file://server/share/repo|server/share/repo
EOF

expect_reject "https://github.com/../repo.git"
expect_reject "https://github.com/owner/../../repo.git"
expect_reject "file:///../repo"
expect_reject "file:///tmp/../repo"
expect_reject "file:///tmp%2f..%2frepo"

# Real end-to-end clone of a local repo (uses real git).
if command -v git >/dev/null; then
  SRC="$(mktemp -d)"
  git init -q "$SRC" && git -C "$SRC" config user.email t@t \
    && git -C "$SRC" config user.name t \
    && echo hi >"$SRC/f.txt" && git -C "$SRC" add . && git -C "$SRC" commit -qm init
  if ./gclone -f "file://$SRC" >/dev/null 2>&1 \
    && [ -f "$WS/local$SRC/f.txt" ]; then
    echo "ok: real file:// clone"
  else
    echo "FAIL: real file:// clone"; fail=1
  fi
  rm -rf "$SRC"
fi

if [ "$fail" = 0 ]; then echo "ALL TESTS PASS"; else echo "TESTS FAILED"; fi
exit "$fail"
