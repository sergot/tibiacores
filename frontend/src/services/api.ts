// API service for interacting with the backend

import { Player, Character } from '@/contexts/PlayerContext';
import { camelCaseToSnakeCase, snakeCaseToCamelCase } from '@/utils/caseConversion';

// Base API URL - environment-specific
const API_URL = process.env.NEXT_PUBLIC_API_URL || '/api';

// Helper function to handle API responses
export const handleResponse = async <T>(response: Response): Promise<T> => {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    console.error('API Error Response:', {
      status: response.status,
      statusText: response.statusText,
      data: errorData
    });
    const errorMessage = errorData.error || `Error: ${response.status} ${response.statusText}`;
    throw new Error(errorMessage);
  }
  
  // Return the data directly without case conversion
  const data = await response.json();
  return data as T;
};

// Player API
export const playerApi = {
  // Create a new player
  createPlayer: async (data: { username: string; character_name: string; world: string }): Promise<Player> => {
    try {
      const response = await fetch(`${API_URL}/players`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });
      return handleResponse<Player>(response);
    } catch (error) {
      console.error('Error creating player:', error);
      throw error;
    }
  },

  // Get a player by ID
  getPlayer: async (playerId: string): Promise<Player> => {
    try {
      const response = await fetch(`${API_URL}/players/${playerId}`);
      return handleResponse<Player>(response);
    } catch (error) {
      console.error('Error getting player:', error);
      throw error;
    }
  },
  
  // Get a player by session ID
  getPlayerBySession: async (sessionId: string): Promise<Player> => {
    try {
      // Only make the request if we have a valid session ID
      if (!sessionId) {
        throw new Error('Session ID is required to fetch player');
      }
      
      const response = await fetch(`${API_URL}/players/session/${sessionId}`);
      return handleResponse<Player>(response);
    } catch (error) {
      console.error('Error getting player by session:', error);
      throw error;
    }
  },

  // Add a character to a player
  addCharacter: async (playerId: string, data: { character_name: string; world: string }): Promise<Character> => {
    try {
      const response = await fetch(`${API_URL}/players/${playerId}/characters`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });
      return handleResponse<Character>(response);
    } catch (error) {
      console.error('Error adding character:', error);
      throw error;
    }
  },

  // Get all characters for a player
  getCharacters: async (playerId: string): Promise<Character[]> => {
    try {
      const response = await fetch(`${API_URL}/players/${playerId}/characters`);
      return handleResponse<Character[]>(response);
    } catch (error) {
      console.error('Error getting characters:', error);
      throw error;
    }
  },

  // Delete a character
  deleteCharacter: async (playerId: string, characterId: string): Promise<void> => {
    try {
      const response = await fetch(`${API_URL}/players/${playerId}/characters/${characterId}`, {
        method: 'DELETE',
      });
      return handleResponse<void>(response);
    } catch (error) {
      console.error('Error deleting character:', error);
      throw error;
    }
  },

  updateUsername: async (playerId: string, username: string): Promise<Player> => {
    try {
      const response = await fetch(`${API_URL}/players/${playerId}/username`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username }),
      });
      
      return handleResponse<Player>(response);
    } catch (error) {
      console.error('Error updating username:', error);
      throw error;
    }
  },
};

// List API
export const listApi = {
  // Get all lists for a player or session
  getLists: async <T>(playerId?: string, sessionId?: string): Promise<T> => {
    try {
      let url = `${API_URL}/lists`;
      
      // Add query parameters if provided
      const params = new URLSearchParams();
      if (playerId) params.append('player_id', playerId);
      if (sessionId) params.append('session_id', sessionId);
      
      // Add query string if we have parameters
      if ([...params].length > 0) {
        url += `?${params.toString()}`;
      }
      
      const response = await fetch(url);
      return handleResponse<T>(response);
    } catch (error) {
      console.error('Error getting lists:', error);
      throw error;
    }
  },
  
  // Create a new list
  createList: async <T>(data: { 
    name: string; 
    description?: string; 
    player_id?: string;
    character_id?: string;
    character_name?: string;
    world?: string;
    session_id?: string;
  }): Promise<T> => {
    try {
      const response = await fetch(`${API_URL}/lists`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });
      
      return handleResponse<T>(response);
    } catch (error) {
      console.error('Error creating list:', error);
      throw error;
    }
  },
  
  // Join a list
  joinList: async <T>(joinCode: string, data: { 
    player_id?: string;
    character_id?: string;
    character_name?: string;
    world?: string;
    session_id?: string;
  }): Promise<T> => {
    try {
      const response = await fetch(`${API_URL}/lists/join?code=${joinCode}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });
      
      return handleResponse<T>(response);
    } catch (error) {
      console.error('Error joining list:', error);
      throw error;
    }
  },
};

// Export a default API object with all services
const api = {
  player: playerApi,
  list: listApi,
};

export default api; 