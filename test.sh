#!/usr/bin/env bash

echo Endpoint: https://movies-${OKTETO_NAMESPACE}.${OKTETO_DOMAIN}/catalog
while true
do
  curl -v http://catalog/catalog
  sleep 1
done