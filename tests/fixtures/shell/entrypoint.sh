#!/usr/bin/env bash
set -e

export PORT=8080
LOG_DIR=/var/log/app

echo "starting on $PORT"
echo "writing to $LOG_DIR"
echo "connecting to $DATABASE_URL"

region="${AWS_REGION:-us-east-1}"
: "${API_KEY:?API_KEY must be set}"
echo "region=$region"
