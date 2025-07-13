#!/bin/bash

API_ENDPOINT="http://localhost:8080/api/v1/version"
NUM_REQUESTS=1000

for _ in $(seq 1 $NUM_REQUESTS); do
  resp=$(curl -s -X 'GET' -H 'accept: application/json' $API_ENDPOINT)
  echo $resp
done
