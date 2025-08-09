type ApiResponse<T> = {
  data?: T;
  error?: string;
};

const getBaseURL = () => process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8888';

const handleResponse = async <T>(response: Response): Promise<ApiResponse<T>> => {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'An error occurred' }));
    return { error: errorData.error || `HTTP error! status: ${response.status}` };
  }

  try {
    const data = await response.json();
    return { data };
  } catch (_) {
    return { error: 'Failed to parse response' };
  }
};

const getAuthToken = () => {
  if (typeof window !== 'undefined') {
    return localStorage.getItem('access_token');
  }
  return null;
};

const request = async <T>(
  endpoint: string,
  options: RequestInit = {},
  requiresAuth: boolean = false
): Promise<ApiResponse<T>> => {
  try {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    // Add auth token if required
    if (requiresAuth) {
      const token = getAuthToken();
      if (token) {
        (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
      }
    }

    const response = await fetch(`${getBaseURL()}${endpoint}`, {
      headers,
      credentials: 'include',
      ...options,
    });

    return handleResponse<T>(response);
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Network error' };
  }
};

export const apiClient = {
  // Public endpoints (no auth required)
  post: <T>(endpoint: string, body?: any) =>
    request<T>(endpoint, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    }, false),

  get: <T>(endpoint: string) =>
    request<T>(endpoint, {
      method: 'GET',
    }, false),

  put: <T>(endpoint: string, body?: any) =>
    request<T>(endpoint, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    }, false),

  delete: <T>(endpoint: string) =>
    request<T>(endpoint, {
      method: 'DELETE',
    }, false),

  // Protected endpoints (auth required)
  auth: {
    post: <T>(endpoint: string, body?: any) =>
      request<T>(endpoint, {
        method: 'POST',
        body: body ? JSON.stringify(body) : undefined,
      }, true),

    get: <T>(endpoint: string) =>
      request<T>(endpoint, {
        method: 'GET',
      }, true),

    put: <T>(endpoint: string, body?: any) =>
      request<T>(endpoint, {
        method: 'PUT',
        body: body ? JSON.stringify(body) : undefined,
      }, true),

    delete: <T>(endpoint: string) =>
      request<T>(endpoint, {
        method: 'DELETE',
      }, true),
  },
};