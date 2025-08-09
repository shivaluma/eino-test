import { useMutation } from '@tanstack/react-query';
import { authApi, LoginRequest, LoginResponse, RegisterRequest, User, CheckEmailRequest } from '@/lib/api/auth';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';

const STORAGE_KEY = 'auth_tokens';
const USER_STORAGE_KEY = 'auth_user';

const saveAuthData = (_tokens: LoginResponse) => {
  // Tokens are set by HttpOnly cookies; optionally cache user for UI responsiveness
  if (typeof window !== 'undefined' && _tokens.user) {
    localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(_tokens.user));
  }
};

const clearAuthData = () => {
  if (typeof window !== 'undefined') {
    localStorage.removeItem(STORAGE_KEY);
    localStorage.removeItem(USER_STORAGE_KEY);
  }
};

export const getTokens = (): LoginResponse | null => {
  if (typeof window === 'undefined') return null;
  
  const stored = localStorage.getItem(STORAGE_KEY);
  if (!stored) return null;
  
  try {
    return JSON.parse(stored);
  } catch {
    return null;
  }
};

export const getUser = (): User | null => {
  if (typeof window === 'undefined') return null;
  
  const stored = localStorage.getItem(USER_STORAGE_KEY);
  if (!stored) return null;
  
  try {
    return JSON.parse(stored);
  } catch {
    return null;
  }
};

export const useCheckEmailMutation = () => {
  return useMutation({
    mutationFn: async (data: CheckEmailRequest) => {
      const response = await authApi.checkEmail(data);
      
      if (response.error) {
        throw new Error(response.error);
      }
      
      if (!response.data) {
        throw new Error('No data received from server');
      }
      
      return response.data;
    },
  });
};

export const useRegisterMutation = () => {
  const router = useRouter();

  return useMutation({
    mutationFn: async (data: RegisterRequest) => {
      const response = await authApi.register(data);
      
      if (response.error) {
        throw new Error(response.error);
      }
      
      if (!response.data) {
        throw new Error('No data received from server');
      }
      
      // After successful registration, log the user in
      const loginResponse = await authApi.login({
        email: data.email,
        password: data.password,
      });
      
      if (loginResponse.error) {
        throw new Error(loginResponse.error);
      }
      
      if (!loginResponse.data) {
        throw new Error('Failed to log in after registration');
      }
      
      return loginResponse.data;
    },
    onSuccess: (data: LoginResponse) => {
      saveAuthData(data);
      toast.success('Account created successfully!');
      router.push('/');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to create account');
    },
  });
};

export const useLoginMutation = () => {
  const router = useRouter();

  return useMutation({
    mutationFn: async (data: LoginRequest) => {
      const response = await authApi.login(data);
      
      if (response.error) {
        throw new Error(response.error);
      }
      
      if (!response.data) {
        throw new Error('No data received from server');
      }
      
      return response.data;
    },
    onSuccess: (data: LoginResponse) => {
      saveAuthData(data);
      toast.success('Successfully logged in!');
      router.push('/');
      router.refresh();
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Invalid email or password');
    },
  });
};

export const useLogout = () => {
  const router = useRouter();
  
  return () => {
    clearAuthData();
    router.push('/sign-in');
    toast.success('Successfully logged out');
  };
};