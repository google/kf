#!/usr/bin/env sh
while true; do
  kubectl logs -f -n kf $(kubectl -n kf get pods | grep con | head -n 1 | awk '{print $1}')
done
