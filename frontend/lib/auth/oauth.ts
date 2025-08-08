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
 * Handle OAuth callback tokens
 */
export const handleOAuthCallback = (accessToken: string, refreshToken: string): void => {
  // Store tokens
  localStorage.setItem('access_token', accessToken);
  localStorage.setItem('refresh_token', refreshToken);
  
  // Redirect to dashboard or home
  window.location.href = '/';
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