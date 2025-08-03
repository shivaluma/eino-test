#!/bin/bash

# Simple API test script
# Usage: ./scripts/test-api.sh [base_url]

BASE_URL=${1:-"http://localhost:8888"}
API_URL="$BASE_URL/api/v1"

echo "Testing AI Food Agent API at $BASE_URL"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test health endpoint
echo -e "\n${YELLOW}1. Testing health endpoint...${NC}"
HEALTH_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" "$BASE_URL/health")
HEALTH_BODY=$(echo $HEALTH_RESPONSE | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
HEALTH_STATUS=$(echo $HEALTH_RESPONSE | tr -d '\n' | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')

if [ "$HEALTH_STATUS" -eq "200" ]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
    echo "Response: $HEALTH_BODY"
else
    echo -e "${RED}✗ Health check failed (HTTP $HEALTH_STATUS)${NC}"
    echo "Response: $HEALTH_BODY"
    exit 1
fi

# Test user registration
echo -e "\n${YELLOW}2. Testing user registration...${NC}"
REGISTER_DATA='{"username":"testuser","email":"test@example.com","password":"password123"}'
REGISTER_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST "$API_URL/register" \
    -H "Content-Type: application/json" \
    -d "$REGISTER_DATA")

REGISTER_BODY=$(echo $REGISTER_RESPONSE | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
REGISTER_STATUS=$(echo $REGISTER_RESPONSE | tr -d '\n' | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')

if [ "$REGISTER_STATUS" -eq "201" ] || [ "$REGISTER_STATUS" -eq "409" ]; then
    if [ "$REGISTER_STATUS" -eq "409" ]; then
        echo -e "${YELLOW}⚠ User already exists (continuing with login)${NC}"
    else
        echo -e "${GREEN}✓ User registration successful${NC}"
    fi
    echo "Response: $REGISTER_BODY"
else
    echo -e "${RED}✗ User registration failed (HTTP $REGISTER_STATUS)${NC}"
    echo "Response: $REGISTER_BODY"
    exit 1
fi

# Test user login
echo -e "\n${YELLOW}3. Testing user login...${NC}"
LOGIN_DATA='{"email":"test@example.com","password":"password123"}'
LOGIN_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST "$API_URL/login" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_DATA")

LOGIN_BODY=$(echo $LOGIN_RESPONSE | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
LOGIN_STATUS=$(echo $LOGIN_RESPONSE | tr -d '\n' | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')

if [ "$LOGIN_STATUS" -eq "200" ]; then
    echo -e "${GREEN}✓ User login successful${NC}"
    
    # Extract access token
    ACCESS_TOKEN=$(echo $LOGIN_BODY | grep -o '"access_token":"[^"]*' | sed 's/"access_token":"//')
    
    if [ -n "$ACCESS_TOKEN" ]; then
        echo "Access token received (length: ${#ACCESS_TOKEN})"
    else
        echo -e "${RED}✗ No access token in response${NC}"
        exit 1
    fi
else
    echo -e "${RED}✗ User login failed (HTTP $LOGIN_STATUS)${NC}"
    echo "Response: $LOGIN_BODY"
    exit 1
fi

# Test protected endpoint (conversations)
echo -e "\n${YELLOW}4. Testing protected endpoint (conversations)...${NC}"
CONVERSATIONS_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X GET "$API_URL/conversations" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

CONVERSATIONS_BODY=$(echo $CONVERSATIONS_RESPONSE | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
CONVERSATIONS_STATUS=$(echo $CONVERSATIONS_RESPONSE | tr -d '\n' | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')

if [ "$CONVERSATIONS_STATUS" -eq "200" ]; then
    echo -e "${GREEN}✓ Protected endpoint access successful${NC}"
    echo "Response: $CONVERSATIONS_BODY"
else
    echo -e "${RED}✗ Protected endpoint access failed (HTTP $CONVERSATIONS_STATUS)${NC}"
    echo "Response: $CONVERSATIONS_BODY"
    exit 1
fi

# Test creating a conversation
echo -e "\n${YELLOW}5. Testing conversation creation...${NC}"
CREATE_CONV_DATA='{"title":"Test Conversation"}'
CREATE_CONV_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST "$API_URL/conversations" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "$CREATE_CONV_DATA")

CREATE_CONV_BODY=$(echo $CREATE_CONV_RESPONSE | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
CREATE_CONV_STATUS=$(echo $CREATE_CONV_RESPONSE | tr -d '\n' | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')

if [ "$CREATE_CONV_STATUS" -eq "201" ]; then
    echo -e "${GREEN}✓ Conversation creation successful${NC}"
    echo "Response: $CREATE_CONV_BODY"
else
    echo -e "${RED}✗ Conversation creation failed (HTTP $CREATE_CONV_STATUS)${NC}"
    echo "Response: $CREATE_CONV_BODY"
    exit 1
fi

echo -e "\n${GREEN}🎉 All API tests passed successfully!${NC}"
echo -e "\nAPI is ready for development. You can now:"
echo "  • Register new users"
echo "  • Authenticate with JWT tokens"
echo "  • Create and manage conversations"
echo "  • Access protected endpoints"