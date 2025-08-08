import { api } from '@/lib/api/api-client';

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

class OAuthClient {
  /**
   * Get list of enabled OAuth providers
   */
  async getProviders(): Promise<OAuthProvider[]> {
    const response = await api.get<OAuthProvidersResponse>('/auth/oauth/providers');
    return response.providers;
  }

  /**
   * Initiate OAuth flow for a provider
   */
  async initiateOAuth(provider: OAuthProvider): Promise<void> {
    // For web flow, directly redirect
    const authUrl = `${api.baseUrl}/auth/oauth/${provider}/authorize`;
    window.location.href = authUrl;
  }

  /**
   * Initiate OAuth flow and get URL (for custom handling)
   */
  async getOAuthUrl(provider: OAuthProvider): Promise<string> {
    const response = await api.get<OAuthInitResponse>(
      `/auth/oauth/${provider}/authorize?redirect=false`
    );
    return response.auth_url;
  }

  /**
   * Handle OAuth callback tokens
   */
  handleCallback(accessToken: string, refreshToken: string): void {
    // Store tokens
    localStorage.setItem('access_token', accessToken);
    localStorage.setItem('refresh_token', refreshToken);
    
    // Redirect to dashboard or home
    window.location.href = '/';
  }

  /**
   * Get linked OAuth accounts for the current user
   */
  async getLinkedAccounts(): Promise<LinkedAccount[]> {
    const response = await api.get<{ linked_accounts: LinkedAccount[] }>(
      '/auth/oauth/linked'
    );
    return response.linked_accounts || [];
  }

  /**
   * Link an OAuth account to the current user
   */
  async linkAccount(provider: OAuthProvider): Promise<void> {
    const authUrl = `${api.baseUrl}/auth/oauth/${provider}/link`;
    window.location.href = authUrl;
  }

  /**
   * Unlink an OAuth account from the current user
   */
  async unlinkAccount(provider: OAuthProvider): Promise<void> {
    await api.delete(`/auth/oauth/${provider}/unlink`);
  }

  /**
   * Check if a provider is available
   */
  async isProviderAvailable(provider: OAuthProvider): Promise<boolean> {
    try {
      const providers = await this.getProviders();
      return providers.includes(provider);
    } catch {
      return false;
    }
  }
}

export const oauthClient = new OAuthClient();