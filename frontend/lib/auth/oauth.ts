import { apiClient } from '@/lib/api/client';

export type OAuthProvider = 'github' | 'google';

export interface OAuthProvidersResponse {
  providers: OAuthProvider[];
}

export interface OAuthInitResponse {
  auth_url: string;
  state: string;
}

export interface LinkedAccount {
  provider: string;
  username?: string;
  email?: string;
  avatar_url?: string;
  created_at: string;
}

const getBaseURL = () => process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8888';

/**
 * Get list of enabled OAuth providers
 */
export const getOAuthProviders = async (): Promise<OAuthProvider[]> => {
  const response = await apiClient.get<OAuthProvidersResponse>('/api/v1/auth/oauth/providers');
  if (response.error) {
    throw new Error(response.error);
  }
  return response.data?.providers || [];
};

/**
 * Initiate OAuth flow for a provider
 */
export const initiateOAuth = (provider: OAuthProvider): void => {
  // For web flow, directly redirect
  const authUrl = `${getBaseURL()}/api/v1/auth/oauth/${provider}/authorize`;
  window.location.href = authUrl;
};

/**
 * Initiate OAuth flow and get URL (for custom handling)
 */
export const getOAuthUrl = async (provider: OAuthProvider): Promise<string> => {
  const response = await apiClient.get<OAuthInitResponse>(
    `/api/v1/auth/oauth/${provider}/authorize?redirect=false`
  );
  if (response.error) {
    throw new Error(response.error);
  }
  return response.data?.auth_url || '';
};

/**
 * Handle OAuth callback tokens with security best practices
 */
export const handleOAuthCallback = async (accessToken: string, refreshToken: string): Promise<void> => {
  try {
    // Validate token format before storing
    if (!isValidJWT(accessToken) || !isValidJWT(refreshToken)) {
      throw new Error('Invalid token format received');
    }

    // Store tokens securely
    localStorage.setItem('access_token', accessToken);
    localStorage.setItem('refresh_token', refreshToken);
    
    // Set httpOnly cookie for enhanced security (if backend supports it)
    try {
      await fetch('/api/auth/set-session', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ accessToken, refreshToken }),
      });
    } catch (cookieError) {
      console.warn('Could not set secure session cookie:', cookieError);
      // Continue anyway as localStorage is still available
    }

    // Clear any previous error states
    localStorage.removeItem('auth_error');
    
  } catch (error) {
    // Clean up on failure
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.setItem('auth_error', error instanceof Error ? error.message : 'Authentication failed');
    throw error;
  }
};

/**
 * Validate JWT token format
 */
const isValidJWT = (token: string): boolean => {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return false;
    
    // Validate each part can be decoded
    parts.forEach(part => {
      if (!part) throw new Error('Empty JWT part');
      // Convert URL-safe base64 to regular base64
      const base64 = part.replace(/-/g, '+').replace(/_/g, '/');
      atob(base64);
    });
    
    return true;
  } catch {
    return false;
  }
};

/**
 * Get linked OAuth accounts for the current user
 */
export const getLinkedAccounts = async (): Promise<LinkedAccount[]> => {
  const response = await apiClient.auth.get<{ linked_accounts: LinkedAccount[] }>(
    '/api/v1/auth/oauth/linked'
  );
  if (response.error) {
    throw new Error(response.error);
  }
  return response.data?.linked_accounts || [];
};

/**
 * Link an OAuth account to the current user
 */
export const linkOAuthAccount = (provider: OAuthProvider): void => {
  const authUrl = `${getBaseURL()}/api/v1/auth/oauth/${provider}/link`;
  window.location.href = authUrl;
};

/**
 * Unlink an OAuth account from the current user
 */
export const unlinkOAuthAccount = async (provider: OAuthProvider): Promise<void> => {
  const response = await apiClient.auth.delete(`/api/v1/auth/oauth/${provider}/unlink`);
  if (response.error) {
    throw new Error(response.error);
  }
};

/**
 * Check if a provider is available
 */
export const isOAuthProviderAvailable = async (provider: OAuthProvider): Promise<boolean> => {
  try {
    const providers = await getOAuthProviders();
    return providers.includes(provider);
  } catch {
    return false;
  }
};