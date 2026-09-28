#!/usr/bin/env bash
find . -type f \( -name '*.go' -o -name '*.sql' -o -name '*.json' -o -name '*.sh' \) ! -path './.git/*' ! -name '.env' -print0 | while IFS= read -r -d '' file; do printf '\n\n===== FILE: %s =====\n\n' "$file"; cat "$file"; done > ~/downloads/mws-api/mws-api-source-go.txt
