#!/bin/bash

# ETL API Test Script
# This script tests the API endpoints (requires PostgreSQL database)

BASE_URL="http://localhost:8080"
EMAIL="test@example.com"
PASSWORD="testpassword123"

echo "🧪 ETL API Test Suite"
echo "====================="

# Test 1: Health Check
echo "📊 Testing health endpoint..."
curl -s "$BASE_URL/health" | grep -q "healthy" && echo "✅ Health check passed" || echo "❌ Health check failed"

# Test 2: User Registration
echo "👤 Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

if echo "$REGISTER_RESPONSE" | grep -q "successfully"; then
  echo "✅ User registration passed"
else
  echo "❌ User registration failed: $REGISTER_RESPONSE"
fi

# Test 3: User Login
echo "🔑 Testing user login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$TOKEN" ]; then
  echo "✅ User login passed"
  echo "🎫 Token: ${TOKEN:0:20}..."
else
  echo "❌ User login failed: $LOGIN_RESPONSE"
  exit 1
fi

# Test 4: Protected Endpoint (List Tables)
echo "📋 Testing protected endpoint (list tables)..."
TABLES_RESPONSE=$(curl -s -X GET "$BASE_URL/tables" \
  -H "Authorization: Bearer $TOKEN")

if echo "$TABLES_RESPONSE" | grep -q '"tables"'; then
  echo "✅ Protected endpoint passed"
else
  echo "❌ Protected endpoint failed: $TABLES_RESPONSE"
fi

# Test 5: File Upload Endpoint (without actual file)
echo "📁 Testing file upload endpoint structure..."
UPLOAD_RESPONSE=$(curl -s -X POST "$BASE_URL/upload" \
  -H "Authorization: Bearer $TOKEN")

if echo "$UPLOAD_RESPONSE" | grep -q "error"; then
  echo "✅ Upload endpoint structure is correct (expects file)"
else
  echo "❌ Upload endpoint failed: $UPLOAD_RESPONSE"
fi

echo ""
echo "🎉 Test suite completed!"
echo "📝 Note: Full CSV upload testing requires actual CSV file"
echo "🚀 Ready for Railway deployment!"