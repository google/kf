#!/usr/bin/env bash

require_env() {
  local name=$1
  local value=$(eval echo '$'$name)
  if [ "$value" == '' ]; then
    echo "environment variable $name must be set"
    exit 1
  fi
}
