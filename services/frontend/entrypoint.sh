#!/bin/sh
sed -i "s|<!-- APP_CONFIG_PLACEHOLDER -->|<script>window.__APP_NAME__=\"${APP_NAME:-Question Voting App}\";</script>|" /usr/share/nginx/html/index.html
exec nginx -g 'daemon off;'
