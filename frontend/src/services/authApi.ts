import { Player } from '@/contexts/PlayerContext';

const BASE_API_URL = '/api';

// Helper function to handle API responses
const handleResponse = async (response: Response) => {
  const data = await response.json();
  
  if (!response.ok) {
    throw new Error(data.error || 'Something went wrong');
  }
  
  return data;
};

// Authentication API
export const authApi = {
  // Register a new user
  register: async (email: string, password: string, sessionId?: string): Promise<{ token: string; playerID: string; username: string }> => {
    const response = await fetch(`${BASE_API_URL}/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email,
        password,
        session_id: sessionId,
        redirect_url: `${window.location.origin}/verify-email-success`,
      }),
    });
    
    const data = await handleResponse(response);
    
    // Store token in localStorage
    localStorage.setItem('auth_token', data.token);
    
    // Clear any anonymous player data
    localStorage.removeItem('player');
    
    return data;
  },
  
  // Login user
  login: async (email: string, password: string): Promise<{ token: string; playerID: string; username: string }> => {
    const response = await fetch(`${BASE_API_URL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email,
        password,
      }),
    });
    
    const data = await handleResponse(response);
    
    // Store token in localStorage
    localStorage.setItem('auth_token', data.token);
    
    // Clear any anonymous player data
    localStorage.removeItem('player');
    
    return data;
  },
  
  // Google OAuth login
  googleLogin: async (accessToken: string, sessionId?: string): Promise<{ token: string; playerID: string; username: string }> => {
    const response = await fetch(`${BASE_API_URL}/auth/google`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        provider: 'google',
        access_token: accessToken,
        session_id: sessionId,
      }),
    });
    
    const data = await handleResponse(response);
    
    // Store token in localStorage
    localStorage.setItem('auth_token', data.token);
    
    // Clear any anonymous player data
    localStorage.removeItem('player');
    
    return data;
  },
  
  // Discord OAuth login
  discordLogin: async (accessToken: string, sessionId?: string): Promise<{ token: string; playerID: string; username: string }> => {
    const response = await fetch(`${BASE_API_URL}/auth/discord`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        provider: 'discord',
        access_token: accessToken,
        session_id: sessionId,
      }),
    });
    
    const data = await handleResponse(response);
    
    // Store token in localStorage
    localStorage.setItem('auth_token', data.token);
    
    // Clear any anonymous player data
    localStorage.removeItem('player');
    
    return data;
  },
  
  // Verify email
  verifyEmail: async (token: string): Promise<{ message: string }> => {
    const response = await fetch(`${BASE_API_URL}/auth/verify-email?token=${token}`, {
      method: 'GET',
    });
    
    return handleResponse(response);
  },
  
  // Get current user from token
  getCurrentUser: async (): Promise<Player | null> => {
    const token = localStorage.getItem('auth_token');
    
    if (!token) {
      return null;
    }
    
    const response = await fetch(`${BASE_API_URL}/auth/me`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });
    
    if (!response.ok) {
      localStorage.removeItem('auth_token');
      return null;
    }
    
    return handleResponse(response);
  },
  
  // Logout user
  logout: (): void => {
    localStorage.removeItem('auth_token');
  },
  
  // Check if user is authenticated
  isAuthenticated: (): boolean => {
    return !!localStorage.getItem('auth_token');
  },
}; 