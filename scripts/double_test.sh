#!/usr/bin/env bash
set -a; source .env; set +a

TERM="2267"
#CAREER="ugrd"
CAREER="${1:-ugrd}"
URL="http://localhost:6767"
ENDPOINT="/schedule"

curl -sS -w "Request 1: HTTP %{http_code}, %{time_total}s\n" -o /tmp/mws-api-test-1.out -H "X-NDnerBYETrTEHL6F: $MWSAPI_AUTH_SECRET" "${URL}${ENDPOINT}?term=$TERM&career=$CAREER" &
PID1=$!

curl -sS -w "Request 2: HTTP %{http_code}, %{time_total}s\n" -o /tmp/mws-api-test-2.out -H "X-NDnerBYETrTEHL6F: $MWSAPI_AUTH_SECRET" "${URL}${ENDPOINT}?term=$TERM&career=$CAREER" &
PID2=$!

wait $PID1
RESULT1=$?

wait $PID2
RESULT2=$?

echo "Request 1: exit $RESULT1, $(wc -c < /tmp/mws-api-test-1.out) bytes"
echo "Request 2: exit $RESULT2, $(wc -c < /tmp/mws-api-test-2.out) bytes"
#cat /tmp/mws*
echo "Test complete!"
