/**
 * Token utilities following security best practices
 */

export interface TokenValidationResult {
  isValid: boolean;
  error?: string;
  payload?: any;
}

/**
 * Validate JWT token format and basic structure
 */
export const validateJWT = (token: string): TokenValidationResult => {
  try {
    if (!token || typeof token !== 'string') {
      return { isValid: false, error: 'Token is required' };
    }

    const parts = token.split('.');
    if (parts.length !== 3) {
      return { isValid: false, error: 'Invalid token format' };
    }

    const [header, payload, signature] = parts;

    // Validate each part exists
    if (!header || !payload || !signature) {
      return { isValid: false, error: 'Token parts are missing' };
    }

    // Decode and validate header
    const decodedHeader = decodeBase64URL(header);
    if (!decodedHeader) {
      return { isValid: false, error: 'Invalid token header' };
    }

    // Decode and validate payload
    const decodedPayload = decodeBase64URL(payload);
    if (!decodedPayload) {
      return { isValid: false, error: 'Invalid token payload' };
    }

    // Basic payload validation
    const payloadObj = JSON.parse(decodedPayload);
    if (!payloadObj.sub || !payloadObj.exp) {
      return { isValid: false, error: 'Token missing required claims' };
    }

    // Check if token is expired
    const currentTime = Math.floor(Date.now() / 1000);
    if (payloadObj.exp < currentTime) {
      return { isValid: false, error: 'Token has expired' };
    }

    return { isValid: true, payload: payloadObj };
  } catch (error) {
    return { 
      isValid: false, 
      error: error instanceof Error ? error.message : 'Token validation failed' 
    };
  }
};

/**
 * Decode base64url encoded string
 */
const decodeBase64URL = (str: string): string | null => {
  try {
    // Convert base64url to base64
    const base64 = str.replace(/-/g, '+').replace(/_/g, '/');
    // Add padding if needed
    const padded = base64 + '='.repeat((4 - base64.length % 4) % 4);
    return atob(padded);
  } catch {
    return null;
  }
};

/**
 * Securely store authentication tokens
 */
// Helper function to check authentication status via API
export const checkAuthStatus = async (): Promise<boolean> => {
  try {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    const response = await fetch(`${apiUrl}/api/v1/auth/me`, {
      credentials: 'include' // Include cookies for authentication
    });
    return response.ok; // 200 means authenticated, 401 means not authenticated
  } catch (error) {
    console.error('Auth status check failed:', error);
    return false;
  }
};

// Legacy function for backward compatibility - tokens now managed via HTTP-only cookies
export const storeTokensSecurely = (_accessToken: string, _refreshToken: string): void => {
  console.warn('Tokens are now managed via HTTP-only cookies by the backend');
  // Tokens are now handled server-side via cookies, no client-side storage needed
};

/**
 * Clear stored authentication tokens
 */
export const clearStoredTokens = (): void => {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('token_stored_at');
  
  // Clear any session cookies (adjust cookie name as needed)
  document.cookie = 'session-token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; secure; samesite=strict';
};

/**
 * Get stored tokens with validation
 */
export const getStoredTokens = (): { accessToken: string; refreshToken: string } | null => {
  try {
    const accessToken = localStorage.getItem('access_token');
    const refreshToken = localStorage.getItem('refresh_token');

    if (!accessToken || !refreshToken) {
      return null;
    }

    // Validate tokens before returning
    const accessTokenValidation = validateJWT(accessToken);
    if (!accessTokenValidation.isValid) {
      clearStoredTokens();
      return null;
    }

    return { accessToken, refreshToken };
  } catch {
    clearStoredTokens();
    return null;
  }
};

/**
 * Check if user is authenticated with valid tokens
 */
export const isAuthenticated = (): boolean => {
  const tokens = getStoredTokens();
  return tokens !== null;
};

/**
 * Get access token if valid, otherwise return null
 */
// Note: Logout function moved to authApi in lib/api/auth.ts
// Use authApi.logout() instead of this deprecated function

// Legacy functions - tokens now managed via HTTP-only cookies
export const getValidAccessToken = (): string | null => {
  console.warn('Access tokens are now managed via HTTP-only cookies and not accessible from client-side');
  return null;
};