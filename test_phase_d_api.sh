#!/bin/bash

# Test Phase D API endpoints

API_URL="http://localhost:8080"
CSV_PATH="$(pwd)/testdata/solarfire/testcase-1-transit.csv"

# JN natal data (from phase_d_validation_test.go)
JN_JD="2450800.900729"
JN_LAT="-31.9333"
JN_LON="115.8833"

echo "Testing Phase D API endpoints..."
echo "================================"
echo ""

# Test 1: Single validator (Tr-Na)
echo "Test 1: Single validator (Tr-Na)"
echo "POST /api/v1/validation/timeline"
curl -s -X POST "$API_URL/api/v1/validation/timeline" \
  -H "Content-Type: application/json" \
  -d "{
    \"csv_path\": \"$CSV_PATH\",
    \"natal_jd\": $JN_JD,
    \"natal_lat\": $JN_LAT,
    \"natal_lon\": $JN_LON,
    \"validator_type\": \"tr-na\",
    \"format\": \"json\"
  }" | jq '.'

echo ""
echo "---"
echo ""

# Test 2: Aggregated all validators
echo "Test 2: Aggregated all validators"
echo "POST /api/v1/validation/phase-d"
curl -s -X POST "$API_URL/api/v1/validation/phase-d" \
  -H "Content-Type: application/json" \
  -d "{
    \"csv_path\": \"$CSV_PATH\",
    \"natal_jd\": $JN_JD,
    \"natal_lat\": $JN_LAT,
    \"natal_lon\": $JN_LON,
    \"format\": \"json\"
  }" | jq '.'

echo ""
echo "Test complete"
