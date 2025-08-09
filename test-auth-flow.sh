#!/bin/bash

# Test authentication flow with the new cookie-based system

API_URL="http://localhost:8080/api/v1"
EMAIL="test@example.com"
PASSWORD="password123"

echo "Testing Authentication Flow"
echo "============================"
echo ""

# Test 1: Login and get cookies
echo "1. Testing Login..."
RESPONSE=$(curl -s -c cookies.txt -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
  -w "\nHTTP_STATUS:%{http_code}")

HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$HTTP_STATUS" = "200" ]; then
  echo "✅ Login successful"
  echo "Response body (should only contain user data, not tokens):"
  echo "$BODY" | jq '.'
else
  echo "❌ Login failed with status $HTTP_STATUS"
  echo "$BODY"
fi

echo ""

# Test 2: Access protected endpoint with cookies
echo "2. Testing Protected Endpoint (/auth/me)..."
ME_RESPONSE=$(curl -s -b cookies.txt -X GET "$API_URL/auth/me" \
  -H "Content-Type: application/json" \
  -w "\nHTTP_STATUS:%{http_code}")

ME_STATUS=$(echo "$ME_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
ME_BODY=$(echo "$ME_RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$ME_STATUS" = "200" ]; then
  echo "✅ Protected endpoint accessed successfully with cookies"
  echo "User profile:"
  echo "$ME_BODY" | jq '.'
else
  echo "❌ Failed to access protected endpoint with status $ME_STATUS"
  echo "$ME_BODY"
fi

echo ""

# Test 3: Check cookie contents
echo "3. Checking Cookies..."
if [ -f cookies.txt ]; then
  echo "Cookies set by server:"
  grep -E "access_token|refresh_token" cookies.txt | while read line; do
    if echo "$line" | grep -q "HttpOnly"; then
      echo "✅ $(echo "$line" | awk '{print $6}') is HttpOnly"
    else
      echo "❌ Cookie missing HttpOnly flag"
    fi
  done
else
  echo "❌ No cookies file found"
fi

# Cleanup
rm -f cookies.txt

echo ""
echo "Test complete!"