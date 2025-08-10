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

// Flag to prevent multiple simultaneous refresh attempts
let isRefreshing = false;
let refreshPromise: Promise<boolean> | null = null;

// Export refresh state for other components
export const getRefreshState = () => ({ isRefreshing, refreshPromise });

// Function to refresh token
const refreshToken = async (): Promise<boolean> => {
  if (isRefreshing && refreshPromise) {
    return refreshPromise;
  }

  isRefreshing = true;
  refreshPromise = (async () => {
    try {
      const response = await fetch(`${getBaseURL()}/api/v1/token/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      });

      const success = response.ok;
      
      if (!success) {
        console.error('Token refresh failed:', response.status);
      }
      
      return success;
    } catch (error) {
      console.error('Token refresh error:', error);
      return false;
    } finally {
      isRefreshing = false;
      refreshPromise = null;
    }
  })();

  return refreshPromise;
};

const request = async <T>(
  endpoint: string,
  options: RequestInit = {},
  _requiresAuth: boolean = false
): Promise<ApiResponse<T>> => {
  const makeRequest = async (retryCount = 0): Promise<ApiResponse<T>> => {
    try {
      const headers: HeadersInit = {
        'Content-Type': 'application/json',
        ...options.headers,
      };

      const response = await fetch(`${getBaseURL()}${endpoint}`, {
        headers,
        credentials: 'include',
        ...options,
      });

      // If we get a 401 and haven't already retried, try to refresh token
      if (response.status === 401 && retryCount === 0 && endpoint !== '/api/v1/token/refresh') {
        console.log('Received 401, attempting token refresh...');
        
        const refreshSuccess = await refreshToken();
        
        if (refreshSuccess) {
          console.log('Token refresh successful, retrying original request...');
          // Retry the original request
          return makeRequest(1);
        } else {
          console.log('Token refresh failed, returning 401 response');
          // If refresh fails, let the error propagate
          return handleResponse<T>(response);
        }
      }

      return handleResponse<T>(response);
    } catch (error) {
      return { error: error instanceof Error ? error.message : 'Network error' };
    }
  };

  return makeRequest();
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