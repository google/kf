#!/usr/bin/env bash

set -euo pipefail

cd upstream
COMMIT=$(git rev-parse HEAD)

cd ..
git clone downstream merged

cd merged
git config --local user.name "$GIT_USER_NAME"
git config --local user.email ${GIT_USER_EMAIL}

git remote add upstream ../upstream
git fetch upstream --no-tags

git merge --no-ff --log --no-edit ${COMMIT}
