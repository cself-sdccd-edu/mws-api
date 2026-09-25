#!/usr/bin/env	bash
set -a; source .env; set +a
RUN_VERSION="${1:-code}"
echo "Run version: $RUN_VERSION"
if [[ "$RUN_VERSION" == "code" ]]; then
  go run ./cmd/mws-api
elif [[ "$RUN_VERSION" == "dist" ]]; then
  ./dist/mws-api
else
  echo "Only code/dist run versions supported"
fi

