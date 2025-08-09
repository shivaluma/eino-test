#!/bin/bash

# Test logout functionality

API_URL="http://localhost:8080/api/v1"
EMAIL="test@example.com"
PASSWORD="password123"

echo "Testing Logout Functionality"
echo "============================="
echo ""

# Step 1: Login to get cookies
echo "1. Logging in to get authentication cookies..."
LOGIN_RESPONSE=$(curl -s -c cookies.txt -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
  -w "\nHTTP_STATUS:%{http_code}")

LOGIN_STATUS=$(echo "$LOGIN_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)

if [ "$LOGIN_STATUS" = "200" ]; then
  echo "✅ Login successful"
  echo "Cookies saved to cookies.txt"
else
  echo "❌ Login failed with status $LOGIN_STATUS"
  echo "$LOGIN_RESPONSE"
  exit 1
fi

echo ""

# Step 2: Verify we can access protected endpoint
echo "2. Verifying access to protected endpoint..."
ME_RESPONSE=$(curl -s -b cookies.txt -X GET "$API_URL/auth/me" \
  -H "Content-Type: application/json" \
  -w "\nHTTP_STATUS:%{http_code}")

ME_STATUS=$(echo "$ME_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)

if [ "$ME_STATUS" = "200" ]; then
  echo "✅ Can access protected endpoint with cookies"
else
  echo "❌ Cannot access protected endpoint with status $ME_STATUS"
  exit 1
fi

echo ""

# Step 3: Test logout
echo "3. Testing logout..."
LOGOUT_RESPONSE=$(curl -s -b cookies.txt -c cookies_after_logout.txt -X POST "$API_URL/auth/logout" \
  -H "Content-Type: application/json" \
  -w "\nHTTP_STATUS:%{http_code}")

LOGOUT_STATUS=$(echo "$LOGOUT_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
LOGOUT_BODY=$(echo "$LOGOUT_RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$LOGOUT_STATUS" = "200" ]; then
  echo "✅ Logout successful"
  echo "Response: $LOGOUT_BODY"
else
  echo "❌ Logout failed with status $LOGOUT_STATUS"
  echo "$LOGOUT_RESPONSE"
fi

echo ""

# Step 4: Verify cookies are cleared/expired
echo "4. Verifying cookies are cleared..."
if [ -f cookies_after_logout.txt ]; then
  echo "Checking cookies after logout:"
  if grep -q "access_token.*expires.*1970" cookies_after_logout.txt || grep -q "access_token.*Max-Age.*-1" cookies_after_logout.txt; then
    echo "✅ Access token cookie properly expired"
  else
    echo "❓ Access token cookie status unclear"
  fi
  
  if grep -q "refresh_token.*expires.*1970" cookies_after_logout.txt || grep -q "refresh_token.*Max-Age.*-1" cookies_after_logout.txt; then
    echo "✅ Refresh token cookie properly expired"
  else
    echo "❓ Refresh token cookie status unclear"
  fi
fi

echo ""

# Step 5: Verify we can no longer access protected endpoint
echo "5. Verifying logout by testing protected endpoint access..."
ME_AFTER_LOGOUT=$(curl -s -b cookies_after_logout.txt -X GET "$API_URL/auth/me" \
  -H "Content-Type: application/json" \
  -w "\nHTTP_STATUS:%{http_code}")

ME_AFTER_STATUS=$(echo "$ME_AFTER_LOGOUT" | grep "HTTP_STATUS:" | cut -d: -f2)

if [ "$ME_AFTER_STATUS" = "401" ]; then
  echo "✅ Cannot access protected endpoint after logout (401 Unauthorized)"
  echo "Logout is working correctly!"
else
  echo "❌ Still can access protected endpoint after logout (status $ME_AFTER_STATUS)"
  echo "This indicates logout may not be working properly"
fi

# Cleanup
rm -f cookies.txt cookies_after_logout.txt

echo ""
echo "Logout test complete!"