#!/usr/bin/env	bash
set -a; source .env; set +a

TERM="2273"
CAREER="ugrd"
URL="http://localhost:6767"
#URL="https://mws-api.staging.sdccd.edu"
ENDPOINT="/"
echo "${URL}${ENDPOINT}?term=$TERM&career=$CAREER"
curl -H "X-NDnerBYETrTEHL6F: $MWSAPI_AUTH_SECRET" "${URL}${ENDPOINT}?term=$TERM&career=$CAREER" | tail -20

