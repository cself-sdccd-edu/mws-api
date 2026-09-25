#!/usr/bin/env	bash
set -a; source .env; set +a

TERM="2267"
CAREER="ugrd"
URL="http://localhost:6767"
ENDPOINT="/schedule"

curl -H "X-NDnerBYETrTEHL6F: $MWSAPI_AUTH_SECRET" "${URL}${ENDPOINT}?term=$TERM&career=$CAREER"

