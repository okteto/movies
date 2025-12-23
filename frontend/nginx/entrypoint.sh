#!/bin/sh
set -e

envsubst < /usr/share/nginx/html/config.js.template > /usr/share/nginx/html/config.js
exec nginx -g 'daemon off;'