'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { usePlayer, Character } from '@/contexts/PlayerContext';
import listApi from '@/services/listApi';
import { playerApi } from '@/services/api';
import ListCard from '@/components/lists/ListCard';

interface List {
  id: string;
  name: string;
  description?: string;
  is_creator: boolean;
  member_count: number;
  created_at: string;
  updated_at: string;
  world: string;
}

interface CharacterData {
  name: string;
  world: string;
  level: number;
  vocation: string;
}

export default function ListsPage() {
  const router = useRouter();
  const { player, loading: playerLoading } = usePlayer();
  
  const [lists, setLists] = useState<List[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [joinCode, setJoinCode] = useState('');
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [joinError, setJoinError] = useState<string | null>(null);
  const [joinLoading, setJoinLoading] = useState(false);
  const [showAccountPrompt, setShowAccountPrompt] = useState(false);
  const [mainCharacter, setMainCharacter] = useState<Character | null>(null);
  const [characterData, setCharacterData] = useState<CharacterData | null>(null);

  // Load lists on component mount
  const fetchLists = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      let fetchedLists: List[] = [];
      
      if (player) {
        // Fetch lists for logged-in user
        const userLists = await listApi.getLists(player.id) as List[];
        fetchedLists = userLists;
      } else {
        // Check if we have a temporary session ID
        const tempSessionId = localStorage.getItem('tempSessionId');
        if (tempSessionId) {
          try {
            // Try to fetch the anonymous player first
            const currentPlayer = await playerApi.getPlayerBySession(tempSessionId);
            if (currentPlayer) {
              // Fetch lists for anonymous user
              const anonymousLists = await listApi.getLists(currentPlayer.id) as List[];
              fetchedLists = anonymousLists;
              
              // Show account prompt for anonymous users
              setShowAccountPrompt(true);
            }
          } catch (err) {
            console.error('Failed to fetch anonymous player:', err);
            // Clear the invalid session ID from localStorage to prevent infinite loops
            console.log('Clearing invalid session ID from localStorage in lists page');
            localStorage.removeItem('tempSessionId');
          }
        }
      }
      
      setLists(fetchedLists);
    } catch (err: any) {
      console.error('Failed to fetch lists:', err);
      setError(err.message || 'Failed to fetch lists. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchLists();
  }, [player]);

  // Set main character when player changes
  useEffect(() => {
    if (player && player.characters.length > 0) {
      // Just use the first character for now
      setMainCharacter(player.characters[0]);
    }
  }, [player]);

  // Handle join list
  const handleJoinList = async () => {
    if (!joinCode) {
      setJoinError('Please enter a join code');
      return;
    }

    if (player && !mainCharacter) {
      setJoinError('Please create a character first');
      return;
    }

    setJoinLoading(true);
    setJoinError(null);

    try {
      if (player && mainCharacter) {
        // Join list as logged-in user
        await listApi.joinList(joinCode, {
          characterId: mainCharacter.id,
        });
      } else if (characterData) {
        // Join list as anonymous user
        const tempSessionId = localStorage.getItem('tempSessionId');
        if (!tempSessionId) {
          setJoinError('Session ID not found');
          return;
        }

        await listApi.joinList(joinCode, {
          characterName: characterData.name,
          world: characterData.world,
          sessionId: tempSessionId,
        });
      } else {
        setJoinError('Please enter character details');
        return;
      }

      // Close modal and refresh lists
      setShowJoinModal(false);
      fetchLists();
    } catch (err: any) {
      console.error('Failed to join list:', err);
      setJoinError(err.message || 'Failed to join list. Please try again.');
    } finally {
      setJoinLoading(false);
    }
  };

  useEffect(() => {
    // Redirect to home page
    router.push('/');
  }, [router]);
  
  // Return empty div while redirecting
  return <div></div>;
} 