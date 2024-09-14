import { API_ORDER_URL } from './config';

interface LoginResponse {
  token: string;
}

interface LoginRequest {
  email: string;
  password: string;
}

export const loginUser = async (credentials: LoginRequest): Promise<string> => {
  const response = await fetch(`${API_ORDER_URL}/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(credentials),
  });

  if (!response.ok) {
    throw new Error('Login failed');
  }

  const data: LoginResponse = await response.json();
  return data.token;
};

interface RegisterRequest {
  email: string;
  password: string;
}

export const registerUser = async (credentials: RegisterRequest): Promise<void> => {
  const response = await fetch(`${API_ORDER_URL}/auth/register`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(credentials),
  });

  if (!response.ok) {
    throw new Error('Registration failed');
  }
};

export const forgotPassword = async (email: string): Promise<void> => {
  const response = await fetch(`${API_ORDER_URL}/auth/forgot-password`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email }),
  });

  if (!response.ok) {
    throw new Error('Failed to process forgot password request');
  }
};

export const resetPassword = async (email: string, resetToken: string, newPassword: string): Promise<void> => {
  const response = await fetch(`${API_ORDER_URL}/auth/reset-password`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email, resetToken, newPassword }),
  });

  if (!response.ok) {
    throw new Error('Failed to reset password');
  }
};
