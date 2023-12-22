#!/usr/bin/env bash

echo Endpoint: http://catalog/catalog
while true
do
  curl -v http://catalog:8080/catalog
  sleep 1
done