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

  if (playerLoading) {
    return (
      <div className="flex justify-center py-8">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-amber-500"></div>
      </div>
    );
  }

  if (!player) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center border border-amber-200 dark:border-amber-800">
        <svg className="h-16 w-16 text-amber-500 mx-auto mb-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
        </svg>
        <h3 className="text-lg font-semibold mb-2">No Account Found</h3>
        <p className="text-gray-500 dark:text-gray-400 mb-4">
          You need to create an account before viewing lists.
        </p>
        <Link
          href="/"
          className="inline-flex items-center px-4 py-2 rounded-md text-white font-medium bg-amber-600 hover:bg-amber-700"
        >
          Create Account
        </Link>
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-amber-800 dark:text-amber-500">Your Lists</h1>
        <div className="flex space-x-2">
          <button
            onClick={() => setShowJoinModal(true)}
            className="inline-flex items-center px-4 py-2 rounded-md text-white font-medium bg-amber-600 hover:bg-amber-700"
          >
            <svg className="h-5 w-5 mr-2" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path d="M8 9a3 3 0 100-6 3 3 0 000 6zM8 11a6 6 0 016 6H2a6 6 0 016-6zM16 7a1 1 0 10-2 0v1h-1a1 1 0 100 2h1v1a1 1 0 102 0v-1h1a1 1 0 100-2h-1V7z" />
            </svg>
            Join List
          </button>
          <Link
            href="/lists/create"
            className="inline-flex items-center px-4 py-2 rounded-md text-white font-medium bg-amber-600 hover:bg-amber-700"
          >
            <svg className="h-5 w-5 mr-2" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clipRule="evenodd" />
            </svg>
            Create List
          </Link>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-900 border-l-4 border-red-500 p-4 mb-6">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-500" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <p className="text-sm text-red-700 dark:text-red-300">{error}</p>
            </div>
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="flex justify-center py-8">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-amber-500"></div>
        </div>
      ) : lists.length === 0 ? (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center border border-amber-200 dark:border-amber-800">
          <svg className="h-16 w-16 text-amber-500 mx-auto mb-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
          <h3 className="text-lg font-semibold mb-2">No Lists Found</h3>
          <p className="text-gray-500 dark:text-gray-400 mb-4">
            You haven't created or joined any lists yet.
          </p>
          <div className="flex justify-center space-x-4">
            <button
              onClick={() => setShowJoinModal(true)}
              className="inline-flex items-center px-4 py-2 rounded-md text-white font-medium bg-amber-600 hover:bg-amber-700"
            >
              Join a List
            </button>
            <Link
              href="/lists/create"
              className="inline-flex items-center px-4 py-2 rounded-md text-white font-medium bg-amber-600 hover:bg-amber-700"
            >
              Create a List
            </Link>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {lists.map((list) => (
            <ListCard key={list.id} list={list} />
          ))}
        </div>
      )}

      {/* Join List Modal */}
      {showJoinModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg max-w-md w-full p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Join a List</h3>
            
            <form onSubmit={handleJoinList}>
              <div className="mb-4">
                <label htmlFor="joinCode" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Enter List Code
                </label>
                <input
                  type="text"
                  id="joinCode"
                  value={joinCode}
                  onChange={(e) => setJoinCode(e.target.value)}
                  placeholder="Enter the list code"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
                />
                {joinError && (
                  <p className="mt-2 text-sm text-red-600 dark:text-red-400">{joinError}</p>
                )}
              </div>
              
              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setShowJoinModal(false)}
                  className="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={joinLoading}
                  className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {joinLoading ? (
                    <>
                      <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Joining...
                    </>
                  ) : (
                    'Join List'
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
} 