#!/usr/bin/env	bash
set -a; source .env; set +a
RUN_VERSION="${1:-code}"
CONFIG_FILE="${2:-./config/app.json}"
echo "Run version: $RUN_VERSION"
if [[ "$RUN_VERSION" == "code" ]]; then
  go run ./cmd/mws-api -config "$CONFIG_FILE"
elif [[ "$RUN_VERSION" == "dist" ]]; then
  ./dist/mws-api -config "$CONFIG_FILE"
else
  echo "Only code/dist run versions supported"
fi

