import { headers as getHeaders } from 'next/headers';

type BackendUser = {
  id: string;
  username: string;
  email: string;
  created_at: string;
  updated_at: string;
};

export type UISessionUser = {
  id: string;
  name: string;
  email: string;
  image: string | null;
};

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

function extractBearerFromHeadersOrCookies(incoming: Headers): string | null {
  const auth = incoming.get('authorization') || incoming.get('Authorization');
  if (auth && auth.startsWith('Bearer ')) return auth;
  const cookieHeader = incoming.get('cookie') || incoming.get('Cookie');
  if (!cookieHeader) return null;
  const token = cookieHeader
    .split(';')
    .map((c) => c.trim())
    .find((c) => c.startsWith('access_token='))
    ?.split('=')[1];
  return token ? `Bearer ${decodeURIComponent(token)}` : null;
}

export const auth = {
  api: {
    async getSession(params?: { headers?: Headers }): Promise<{ session: { user: UISessionUser }; user: UISessionUser }> {
      const incomingHeaders = params?.headers ?? (await getHeaders());
      const bearer = extractBearerFromHeadersOrCookies(incomingHeaders);
      if (!bearer) {
        throw new Error('Unauthenticated');
      }

      const res = await fetch(`${API_BASE_URL}/api/v1/auth/me`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          Authorization: bearer,
        },
        // Do not forward cookies by default; we rely on Bearer
      });

      if (!res.ok) {
        throw new Error('Unauthenticated');
      }

      const data: BackendUser = await res.json();
      const uiUser: UISessionUser = {
        id: data.id,
        name: data.username,
        email: data.email,
        image: null,
      };
      return { session: { user: uiUser }, user: uiUser };
    },
  },
};

export type AuthServer = typeof auth;

