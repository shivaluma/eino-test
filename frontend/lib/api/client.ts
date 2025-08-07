type ApiResponse<T> = {
  data?: T;
  error?: string;
};

const getBaseURL = () => process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

const handleResponse = async <T>(response: Response): Promise<ApiResponse<T>> => {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'An error occurred' }));
    return { error: errorData.error || `HTTP error! status: ${response.status}` };
  }

  try {
    const data = await response.json();
    return { data };
  } catch (error) {
    return { error: 'Failed to parse response' };
  }
};

const request = async <T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<ApiResponse<T>> => {
  try {
    const response = await fetch(`${getBaseURL()}${endpoint}`, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    });

    return handleResponse<T>(response);
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Network error' };
  }
};

export const apiClient = {
  post: <T>(endpoint: string, body: any) =>
    request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  get: <T>(endpoint: string) =>
    request<T>(endpoint, {
      method: 'GET',
    }),

  put: <T>(endpoint: string, body: any) =>
    request<T>(endpoint, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),

  delete: <T>(endpoint: string) =>
    request<T>(endpoint, {
      method: 'DELETE',
    }),
};