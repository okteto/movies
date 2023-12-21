#!/usr/bin/env bash

echo Endpoint: https://movies-${OKTETO_NAMESPACE}.${OKTETO_DOMAIN}/catalog
while true
do
  curl -v https://movies-${OKTETO_NAMESPACE}.${OKTETO_DOMAIN}/catalog
  sleep 1
done