#!/usr/bin/env bash
# This file allows launching the application in a Cloud Foundry compatible way
# where arguments to the start command are passed to /lifecycle/launcher like
# they would be in Diego.
set -e

if [[ "$@" == "" ]]; then
  exec /lifecycle/launcher "/home/vcap/app" "" ""
else
  exec /lifecycle/launcher "/home/vcap/app" "$@" ""
fi