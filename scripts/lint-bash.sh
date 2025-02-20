#!/bin/bash
# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

# mapfile reads stdin lines into array, -t trims newlines
mapfile -t files < <(
  find scripts e2e -type f \( -name '*.sh' -o -name '*.bash' \)
)
mapfile -t check_files < <(
  echo "${files[@]}" |
    tr ' ' '\n' |
    sort |
    uniq -u
)

# Handle bats tests
bats_tmp="$(mktemp -d lint-bash-XXXXXX)"
function cleanup() {
  rm -rf "${bats_tmp}"
}
trap cleanup EXIT

readonly linter=koalaman/shellcheck:v0.6.0

if ! docker image inspect "$linter" &>/dev/null; then
  docker pull "$linter"
fi

cmd=(docker run -v "$(pwd):/mnt")
if [ -t 1 ]; then
  cmd+=(--tty)
fi
cmd+=(
  --rm
  "$linter" "${check_files[@]}"
)

echo "Linting scripts..."
"${cmd[@]}"
echo "PASS"
