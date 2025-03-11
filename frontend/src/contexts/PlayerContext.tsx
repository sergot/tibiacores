'use client';

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { playerApi } from '@/services/api';

// Define types
export interface Character {
  id: string;
  name: string;
  world: string;
  level: number;
  vocation: string;
  created_at: string;
  updated_at: string;
}

export interface Player {
  id: string;
  username: string;
  characters: Character[];
  session_id?: string;
  is_anonymous?: boolean;
  created_at: string;
  updated_at: string;
}

export interface PlayerContextType {
  player: Player | null;
  characters: Character[];
  loading: boolean;
  error: string | null;
  createPlayer: (data: { username: string; characterName: string; world: string }) => Promise<void>;
  addCharacter: (data: { name: string; world: string; level: number; vocation: string }) => Promise<void>;
  fetchCharacters: () => Promise<void>;
  deleteCharacter: (characterId: string) => Promise<void>;
  logout: () => void;
  fetchAnonymousPlayer: (sessionId: string) => Promise<void>;
  isAnonymous: boolean;
}

// Create context
const PlayerContext = createContext<PlayerContextType | undefined>(undefined);

// Provider component
export const PlayerProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [player, setPlayer] = useState<Player | null>(null);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isAnonymous, setIsAnonymous] = useState(false);

  // Load player from localStorage on initial render
  useEffect(() => {
    console.log('PlayerContext initializing, checking localStorage for player data...');
    const storedPlayer = localStorage.getItem('player');
    if (storedPlayer) {
      try {
        const parsedPlayer = JSON.parse(storedPlayer);
        console.log('Found player data in localStorage:', parsedPlayer);
        setPlayer(parsedPlayer);
        setIsAnonymous(false);
      } catch (err) {
        console.error('Failed to parse stored player:', err);
        localStorage.removeItem('player');
      }
    } else {
      console.log('No player data found in localStorage');
      
      // Check if we have a temporary session ID
      const tempSessionId = localStorage.getItem('tempSessionId');
      // Only try to fetch anonymous player if we have a session ID (user has created a list)
      if (tempSessionId) {
        console.log('Session ID found:', tempSessionId);
        // Automatically fetch anonymous player if session ID exists
        fetchAnonymousPlayer(tempSessionId)
          .then(() => {
            console.log('Successfully fetched anonymous player with session ID');
          })
          .catch(err => {
            console.error('Failed to fetch anonymous player:', err);
          });
      } else {
        console.log('No session ID found, user needs to log in or create an account');
      }
    }
    setLoading(false);
  }, []);

  // Fetch anonymous player by session ID
  const fetchAnonymousPlayer = async (sessionId: string) => {
    setLoading(true);
    setError(null);
    
    try {
      // Call the API to get player by session ID
      const response = await playerApi.getPlayerBySession(sessionId);
      
      if (response) {
        setPlayer(response);
        setIsAnonymous(true);
        console.log('Fetched anonymous player:', response);
      }
    } catch (err: any) {
      console.error('Failed to fetch anonymous player:', err);
      setError(err.message || 'Failed to fetch anonymous player. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  // Create a new player
  const createPlayer = async (data: { username: string; characterName: string; world: string }) => {
    setLoading(true);
    setError(null);
    
    try {
      const apiData = {
        username: data.username,
        character_name: data.characterName,
        world: data.world
      };
      
      const response = await playerApi.createPlayer(apiData);
      setPlayer(response);
      localStorage.setItem('player', JSON.stringify(response));
    } catch (err: any) {
      console.error('Failed to create player:', err);
      setError(err.message || 'Failed to create player. Please try again.');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // Fetch characters for the current player
  const fetchCharacters = useCallback(async () => {
    if (!player) return;
    
    setLoading(true);
    setError(null);
    
    try {
      const response = await playerApi.getCharacters(player.id);
      setCharacters(response);
    } catch (err: any) {
      console.error('Failed to fetch characters:', err);
      setError(err.message || 'Failed to fetch characters. Please try again.');
    } finally {
      setLoading(false);
    }
  }, [player]);

  // Add a new character
  const addCharacter = async (data: { name: string; world: string; level: number; vocation: string }) => {
    if (!player) {
      throw new Error('You need to create an account first');
    }
    
    setLoading(true);
    setError(null);
    
    try {
      // Convert to snake_case for API
      const characterData = {
        character_name: data.name,
        world: data.world
      };
      
      const newCharacter = await playerApi.addCharacter(player.id, characterData);
      setCharacters(prev => [...prev, newCharacter]);
    } catch (err: any) {
      console.error('Failed to add character:', err);
      setError(err.message || 'Failed to add character. Please try again.');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // Delete a character
  const deleteCharacter = async (characterId: string) => {
    if (!player) {
      throw new Error('You need to create an account first');
    }
    
    setLoading(true);
    setError(null);
    
    try {
      await playerApi.deleteCharacter(player.id, characterId);
      setCharacters(prev => prev.filter(char => char.id !== characterId));
    } catch (err: any) {
      console.error('Failed to delete character:', err);
      setError(err.message || 'Failed to delete character. Please try again.');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // Logout
  const logout = () => {
    setPlayer(null);
    setCharacters([]);
    localStorage.removeItem('player');
  };

  return (
    <PlayerContext.Provider
      value={{
        player,
        characters,
        loading,
        error,
        createPlayer,
        addCharacter,
        fetchCharacters,
        deleteCharacter,
        logout,
        fetchAnonymousPlayer,
        isAnonymous
      }}
    >
      {children}
    </PlayerContext.Provider>
  );
};

// Custom hook to use the player context
export const usePlayer = () => {
  const context = useContext(PlayerContext);
  if (context === undefined) {
    throw new Error('usePlayer must be used within a PlayerProvider');
  }
  return context;
}; 