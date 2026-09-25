#!/usr/bin/env bash
set -a; source .env; set +a

TERMS=(2267 2265 2263 2257 2255 2253)
CAREERS=(ugrd ce)
URL="http://localhost:6767"
ENDPOINT="/schedule"

while true; do
    echo "===== Starting test cycle $(date '+%Y-%m-%d %H:%M:%S') ====="

	for TERM in "${TERMS[@]}"; do
	    for CAREER in "${CAREERS[@]}"; do
	        REQUEST_COUNT=$((RANDOM % 4 + 1))
	        echo "$TERM/$CAREER: $REQUEST_COUNT request(s)"
	
	        PIDS=()
	
	        for ((i = 1; i <= REQUEST_COUNT; i++)); do
	            curl -sS -w "Request $i: HTTP %{http_code}, %{time_total}s\n" -o /dev/null -H "X-NDnerBYETrTEHL6F: $MWSAPI_AUTH_SECRET" "${URL}${ENDPOINT}?term=$TERM&career=$CAREER" &
	            PIDS+=($!)
	        done
	
	        for PID in "${PIDS[@]}"; do
	            wait "$PID"
	        done
	
	        sleep 0.5
	    done
	done

    DELAY=$((RANDOM % 6 + 2))
    echo "Cycle complete. Waiting ${DELAY}s..."
    sleep "$DELAY"
done
