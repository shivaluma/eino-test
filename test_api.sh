#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# API base URL
BASE_URL="http://localhost:8080/api/v1"

echo -e "${YELLOW}Testing AI Agent API${NC}\n"

# 1. Register a test user
echo -e "${GREEN}1. Registering user...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123456!",
    "name": "Test User"
  }')
echo "Response: $REGISTER_RESPONSE"
echo

# 2. Login to get token
echo -e "${GREEN}2. Logging in...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123456!"
  }')
echo "Response: $LOGIN_RESPONSE"

# Extract access token
ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
echo "Access Token: $ACCESS_TOKEN"
echo

# 3. Send a message (create new conversation)
echo -e "${GREEN}3. Sending first message (creates new conversation)...${NC}"
MESSAGE_RESPONSE=$(curl -s -X POST "$BASE_URL/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "message": "Xin chào, bạn khỏe không?",
    "stream": false
  }')
echo "Response: $MESSAGE_RESPONSE"

# Extract conversation ID
CONVERSATION_ID=$(echo $MESSAGE_RESPONSE | grep -o '"conversation_id":"[^"]*' | cut -d'"' -f4)
echo "Conversation ID: $CONVERSATION_ID"
echo

# 4. Send follow-up message to existing conversation
echo -e "${GREEN}4. Sending follow-up message...${NC}"
FOLLOWUP_RESPONSE=$(curl -s -X POST "$BASE_URL/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"message\": \"Tôi muốn học nấu phở, bạn có thể chỉ cho tôi không?\",
    \"conversation_id\": \"$CONVERSATION_ID\",
    \"stream\": false
  }")
echo "Response: $FOLLOWUP_RESPONSE"
echo

# 5. Test streaming
echo -e "${GREEN}5. Testing streaming response...${NC}"
echo "Sending streaming request..."
curl -N -X POST "$BASE_URL/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"message\": \"Kể cho tôi nghe một câu chuyện ngắn về một con chó\",
    \"conversation_id\": \"$CONVERSATION_ID\",
    \"stream\": true
  }"
echo
echo

# 6. Get all conversations
echo -e "${GREEN}6. Getting all conversations...${NC}"
CONVERSATIONS=$(curl -s -X GET "$BASE_URL/conversations" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "Conversations: $CONVERSATIONS"
echo

# 7. Get messages from conversation
echo -e "${GREEN}7. Getting messages from conversation...${NC}"
MESSAGES=$(curl -s -X GET "$BASE_URL/conversations/$CONVERSATION_ID/messages" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "Messages: $MESSAGES"
echo

echo -e "${YELLOW}Test completed!${NC}"