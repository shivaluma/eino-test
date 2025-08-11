'use client';

import { useState, useRef, useCallback } from 'react';

// Types to match your backend models
export interface Message {
  id: number | string;
  conversation_id: string;
  sender_id: string;
  sender_type: 'USER' | 'AGENT';
  content: string;
  metadata?: Record<string, any>;
  created_at: string;
  role: 'user' | 'assistant';
}

export interface UseChatOptions {
  api?: string;
  id?: string;
  initialMessages?: Message[];
  onChunk?: (chunk: string) => void;
  onFinish?: (message: Message) => void;
  onError?: (error: Error) => void;
  prepareRequestBody?: (options: {
    messages: Message[];
    message: string;
    conversationId?: string;
  }) => Record<string, any>;
  headers?: Record<string, string>;
}

export interface StreamChunk {
  type: 'init' | 'chunk' | 'complete' | 'error';
  content?: string;
  conversation_id?: string;
  message_id?: string | number;
  error?: string;
}

export interface UseChatResult {
  messages: Message[];
  input: string;
  setInput: (input: string) => void;
  isLoading: boolean;
  error: Error | null;
  append: (message: string) => Promise<void>;
  reload: () => Promise<void>;
  stop: () => void;
  sendMessage: (message: string) => Promise<void>;
}

const createMessage = (
  content: string,
  role: 'user' | 'assistant',
  conversationId?: string,
  id?: string | number
): Message => ({
  id: id || Date.now().toString(),
  conversation_id: conversationId || '',
  sender_id: role === 'user' ? 'user' : 'assistant',
  sender_type: role === 'user' ? 'USER' : 'AGENT',
  content,
  created_at: new Date().toISOString(),
  role,
});

export function useChat(options: UseChatOptions = {}): UseChatResult {
  const {
    api = '/api/messages',
    id,
    initialMessages = [],
    onChunk,
    onFinish,
    onError,
    prepareRequestBody,
    headers = {},
  } = options;

  const [messages, setMessages] = useState<Message[]>(initialMessages);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [conversationId, setConversationId] = useState<string | undefined>(id);

  const abortControllerRef = useRef<AbortController | null>(null);

  const stop = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
    setIsLoading(false);
  }, []);

  const sendMessage = useCallback(async (message: string) => {
    if (!message.trim()) return;
    
    setError(null);
    setIsLoading(true);

    // Create user message
    const userMessage = createMessage(message, 'user', conversationId);
    setMessages(prev => [...prev, userMessage]);

    // Prepare abort controller
    abortControllerRef.current = new AbortController();

    try {
      // Prepare request body
      const defaultBody = {
        message: message.trim(),
        conversation_id: conversationId || undefined,
        stream: true,
      };

      const requestBody = prepareRequestBody
        ? prepareRequestBody({
            messages,
            message: message.trim(),
            conversationId,
          })
        : defaultBody;

      // Make API call
      const response = await fetch(api, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...headers,
        },
        body: JSON.stringify(requestBody),
        signal: abortControllerRef.current.signal,
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const reader = response.body?.getReader();
      if (!reader) {
        throw new Error('No response body reader available');
      }

      let aiMessageContent = '';
      let aiMessageId: string | number | undefined;
      let tempConversationId = conversationId;

      try {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          // Decode the chunk
          const chunk = new TextDecoder().decode(value);
          const lines = chunk.split('\n');

          for (const line of lines) {
            if (line.startsWith('data: ')) {
              const jsonStr = line.slice(6); // Remove 'data: ' prefix
              if (jsonStr.trim()) {
                try {
                  const data: StreamChunk = JSON.parse(jsonStr);
                  
                  switch (data.type) {
                    case 'init':
                      // Set conversation ID from the response
                      if (data.conversation_id) {
                        tempConversationId = data.conversation_id;
                        setConversationId(data.conversation_id);
                      }
                      break;
                      
                    case 'chunk':
                      if (data.content) {
                        aiMessageContent += data.content;
                        onChunk?.(data.content);
                        
                        // Update the assistant message in real-time
                        setMessages(prev => {
                          const newMessages = [...prev];
                          const lastMessage = newMessages[newMessages.length - 1];
                          
                          if (lastMessage && lastMessage.role === 'assistant') {
                            // Update existing assistant message
                            lastMessage.content = aiMessageContent;
                          } else {
                            // Create new assistant message
                            const assistantMessage = createMessage(
                              aiMessageContent,
                              'assistant',
                              tempConversationId
                            );
                            newMessages.push(assistantMessage);
                          }
                          
                          return newMessages;
                        });
                      }
                      break;
                      
                    case 'complete':
                      aiMessageId = data.message_id;
                      
                      // Finalize the assistant message
                      setMessages(prev => {
                        const newMessages = [...prev];
                        const lastMessage = newMessages[newMessages.length - 1];
                        
                        if (lastMessage && lastMessage.role === 'assistant') {
                          const finalMessage = {
                            ...lastMessage,
                            id: aiMessageId || lastMessage.id,
                            content: aiMessageContent,
                          };
                          
                          onFinish?.(finalMessage);
                          newMessages[newMessages.length - 1] = finalMessage;
                        }
                        
                        return newMessages;
                      });
                      break;
                      
                    case 'error':
                      throw new Error(data.error || 'Stream error');
                  }
                } catch (parseError) {
                  console.error('Error parsing stream data:', parseError);
                }
              }
            }
          }
        }
      } finally {
        reader.releaseLock();
      }
      
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Unknown error');
      
      if (error.name === 'AbortError') {
        // Request was cancelled
        return;
      }
      
      console.error('Chat error:', error);
      setError(error);
      onError?.(error);
      
      // Remove the incomplete message on error
      setMessages(prev => prev.slice(0, -1));
    } finally {
      setIsLoading(false);
      abortControllerRef.current = null;
    }
  }, [
    api,
    conversationId,
    messages,
    onChunk,
    onFinish,
    onError,
    prepareRequestBody,
    headers,
  ]);

  const append = useCallback(async (message: string) => {
    await sendMessage(message);
  }, [sendMessage]);

  const reload = useCallback(async () => {
    if (messages.length === 0) return;
    
    // Find the last user message and resend it
    const lastUserMessage = [...messages].reverse().find(msg => msg.role === 'user');
    if (lastUserMessage) {
      // Remove messages after the last user message
      const lastUserIndex = messages.findLastIndex(msg => msg.role === 'user');
      if (lastUserIndex >= 0) {
        setMessages(prev => prev.slice(0, lastUserIndex + 1));
        await sendMessage(lastUserMessage.content);
      }
    }
  }, [messages, sendMessage]);

  return {
    messages,
    input,
    setInput,
    isLoading,
    error,
    append,
    reload,
    stop,
    sendMessage,
  };
}