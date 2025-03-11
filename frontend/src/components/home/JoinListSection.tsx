'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

interface Character {
  id: string;
  name: string;
  world: string;
  level: number;
  vocation: string;
}

export default function JoinListSection() {
  const router = useRouter();
  const [selectedCharacter, setSelectedCharacter] = useState<Character | null>(null);
  const [share_code, setShareCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load selected character on component mount
  useEffect(() => {
    const loadSelectedCharacter = () => {
      try {
        // Get characters from localStorage
        const storedCharacters = localStorage.getItem('characters');
        if (storedCharacters) {
          const characters = JSON.parse(storedCharacters);
          if (characters.length > 0) {
            setSelectedCharacter(characters[0]);
          }
        }
      } catch (err) {
        console.error('Failed to load selected character:', err);
      }
    };

    loadSelectedCharacter();
  }, []);

  // Function to join a list
  const joinList = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!share_code || !selectedCharacter) return;
    
    setIsLoading(true);
    setError(null);
    
    try {
      // In a real implementation, we would call the API
      // For now, we'll simulate it
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // Navigate to the join page with the share code
      router.push(`/join/${share_code}`);
    } catch (err: any) {
      setError(err.message || 'Failed to join list');
      setIsLoading(false);
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-amber-200 dark:border-amber-800">
      <h2 className="text-xl font-bold text-amber-800 dark:text-amber-500 mb-4">Join a List</h2>
      
      <form onSubmit={joinList} className="mb-4">
        <div className="mb-4">
          <label htmlFor="share_code" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Share Code
          </label>
          <input
            id="share_code"
            type="text"
            value={share_code}
            onChange={(e) => setShareCode(e.target.value)}
            placeholder="Enter the share code"
            disabled={isLoading || !selectedCharacter}
            required
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
          />
        </div>
        
        <button
          type="submit"
          disabled={!share_code || isLoading || !selectedCharacter}
          className={`w-full px-4 py-2 rounded-md text-white font-medium ${
            !share_code || isLoading || !selectedCharacter
              ? 'bg-gray-400 cursor-not-allowed'
              : 'bg-amber-600 hover:bg-amber-700'
          }`}
        >
          {isLoading ? 'Joining...' : 'Join List'}
        </button>
      </form>
      
      {!selectedCharacter && (
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
          <div className="flex">
            <svg className="h-5 w-5 text-blue-500 mr-2" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
            </svg>
            <div>
              <h3 className="text-sm font-medium text-blue-800 dark:text-blue-500">Character Required</h3>
              <p className="text-sm text-blue-700 dark:text-blue-400">Please select a character before joining a list.</p>
            </div>
          </div>
        </div>
      )}
      
      {error && (
        <div className="mt-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <div className="flex">
            <svg className="h-5 w-5 text-red-500 mr-2" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            <div>
              <h3 className="text-sm font-medium text-red-800 dark:text-red-500">Error</h3>
              <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
} 