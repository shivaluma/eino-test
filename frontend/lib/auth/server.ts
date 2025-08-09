import { headers as getHeaders } from 'next/headers';

// Types for API response
type ApiUser = {
  id: string;
  username: string;
  email: string;
  created_at: string;
  updated_at: string;
};

// Type for UI/Session use
export type SessionUser = {
  id: string;
  name: string;
  email: string;
  image: string | null;
};

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const auth = {
  api: {
    async getSession(): Promise<{ session: { user: SessionUser }; user: SessionUser }> {
      // Get the headers including cookies
      const requestHeaders = await getHeaders();
      
      // Create a new Headers object with the cookie header
      const headers = new Headers({
        'Content-Type': 'application/json',
      });
      
      // Forward the cookie header if it exists
      const cookieHeader = requestHeaders.get('cookie');
      if (cookieHeader) {
        headers.set('Cookie', cookieHeader);
      }

      const res = await fetch(`${API_BASE_URL}/api/v1/auth/me`, {
        method: 'GET',
        headers,
        // No need for credentials since we're manually forwarding cookies
      });

      if (!res.ok) {
        throw new Error('Unauthenticated');
      }

      const data: ApiUser = await res.json();
      const sessionUser: SessionUser = {
        id: data.id,
        name: data.username,
        email: data.email,
        image: null,
      };
      
      return { 
        session: { user: sessionUser }, 
        user: sessionUser 
      };
    },
  },
};

export type AuthServer = typeof auth;

