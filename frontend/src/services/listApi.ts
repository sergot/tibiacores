import { handleResponse } from './api';

// Base API URL - use direct API path
const API_URL = process.env.NEXT_PUBLIC_API_URL || '/api';

// List API service
const listApi = {
  // Get all lists for a player
  getLists: async (playerId?: string) => {
    let url = `${API_URL}/lists`;
    
    // Add query parameter if provided
    if (playerId) {
      url += `?player_id=${playerId}`;
    }
    
    try {
      const response = await fetch(url);
      return handleResponse(response);
    } catch (error) {
      console.error('Error getting lists:', error);
      // Return an empty array instead of throwing an error
      return [];
    }
  },
  
  // Get a specific list by ID
  getList: async (listId: string) => {
    return fetch(`${API_URL}/lists/${listId}`)
      .then(handleResponse);
  },
  
  // Get a list by share code
  getListByShareCode: async (share_code: string) => {
    return fetch(`${API_URL}/lists/share/${share_code}`)
      .then(handleResponse);
  },
  
  // Create a new list
  createList: async (data: any) => {
    return fetch(`${API_URL}/lists`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    }).then(handleResponse);
  },
  
  // Join a list
  joinList: async (joinCode: string, data: any) => {
    return fetch(`${API_URL}/lists/join`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    }).then(handleResponse);
  },
  
  // Update a list
  updateList: async (listId: string, data: any) => {
    return fetch(`${API_URL}/lists/${listId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    }).then(handleResponse);
  },
  
  // Delete a list
  deleteList: async (listId: string) => {
    return fetch(`${API_URL}/lists/${listId}`, {
      method: 'DELETE',
    }).then(handleResponse);
  },
  
  // Update a soul core in a list
  updateSoulCore: async (listId: string, soulCoreId: string, data: any) => {
    return fetch(`${API_URL}/lists/${listId}/soul-cores/${soulCoreId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    }).then(handleResponse);
  },
  
  // Add a soul core to a list
  addSoulCore: async (listId: string, data: any) => {
    return fetch(`${API_URL}/lists/${listId}/soul-cores`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    }).then(handleResponse);
  },
  
  // Create a new character
  createCharacter: async (data: any) => {
    return fetch(`${API_URL}/characters`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    }).then(handleResponse);
  },
  
  // Get all creatures
  getCreatures: async () => {
    return fetch(`${API_URL}/creatures`)
      .then(handleResponse);
  },

  // Get lists and soul cores by character ID
  getListsByCharacterID: async (characterId: string) => {
    try {
      const response = await fetch(`${API_URL}/characters/${characterId}/lists`);
      return handleResponse(response);
    } catch (error) {
      console.error('Error getting lists by character ID:', error);
      throw error;
    }
  },
};

export default listApi; 