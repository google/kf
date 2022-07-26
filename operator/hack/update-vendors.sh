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

set -eu

source "$(dirname "${BASH_SOURCE[0]}")/../vendor/knative.dev/hack/library.sh"
source "${REPO_ROOT_DIR}/hack/library.sh"

export FROM_UPDATE_VENDORS=true

pushd "${REPO_ROOT_DIR}"

ORIGINAL_SHA=$(git rev-parse @)

if ! hack/upgrade-forks.sh; then
  set +x
  if [[ $ORIGINAL_SHA == $(git rev-parse @) ]]; then
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
Hello ${USER}!

hack/upgrade-forks.sh failed, but your HEAD is unchanged.
So you can probably recover on your own.
Good luck"
    exit 1
  else
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
Hello ${USER}!

hack/upgrade-forks.sh failed and your HEAD is changed.
If you just aborted, please 'git reset --hard' to get back.
Please don't continue to update vendor without getting hack/upgrade-forks.sh to succeed first.
Reset your head and try again and/or fix any existing problem?"
    exit 1
  fi
fi

while ! hack/update-deps.sh --upgrade; do
  set +x
  echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
Hello ${USER}!

update-deps.sh --upgrade failed. I'm going to rerun it for you.
Please either attempt fix in another window, or Ctrl-Z this, make a fix, and 'fg'
to return to this program, then press <Enter> to try again.
If you are giving up, Ctrl-C will quit.
gl;hf"

  read -p "Press <Enter> to retry update-deps"
done

while ! hack/update-codegen.sh; do
  set +x
  echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
Hello ${USER}!

update-codegen.sh failed. I'm going to rerun it for you.
Please either attempt fix in another window, or Ctrl-Z this, make a fix, and 'fg'
to return to this program, then press <Enter> to try again.
If you are giving up, Ctrl-C will quit.
gl;hf"

  read -p "Press <Enter> to retry update-codegen"
done

set +x
echo "By this time, both update-deps --upgrade and update-codegen should have succeeded
I'm going to create a commit now with all files present.
If you have any scratch files lying around, please delete them before continuing.
(again, either another window or ^Z, make changes, 'fg')"

read -p "Press <Enter> to continue and make commit"
set -x

git add -A
git commit

PRE_SQUASH_SHA=$(git rev-parse @)

set +x
echo "Commit is ready. If you want to do any local testing, do it now.

When ready to continue to squashing commits, press <Enter>"

read -p "Press <Enter> to continue and squash commits for review"

echo "I'm going to squash commits so you have only one to submit"

for i in {0..10}; do
  if [[ $ORIGINAL_SHA == $(git rev-parse HEAD^) ]]; then
    echo "Only one commit exists, you are ready to push for review!"
    exit 0
  fi
  # need "bash" because vendored things don't have executable bit
  bash scripts/shared/squash-commits.sh
  set +x
done

echo "Could not squash to one commit, either too many forks or we over-squashed or something"
git log $PRE_SQUASH_SHA -n 10
exit 1
