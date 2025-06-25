#!/bin/bash

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
URL="http://localhost:9200"

echo "Creating Elasticsearch index with mapping..."
curl -X PUT "${URL}/users" -H "Content-Type: application/json" -d @"${SCRIPT_DIR}/mapping.json"
echo

echo "Loading test data..."
curl -X POST "${URL}/_bulk" -H "Content-Type: application/json" --data-binary @"${SCRIPT_DIR}/test_users.ndjson"
echo

echo "Migration completed!"
