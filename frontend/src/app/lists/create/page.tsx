'use client';

import { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { usePlayer } from '@/contexts/PlayerContext';
import listApi from '@/services/listApi';

function CreateListContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const fromHome = searchParams.get('fromHome') === 'true';
  const characterIdParam = searchParams.get('character_id');
  
  const { player, characters, loading: playerLoading, fetchCharacters, fetchAnonymousPlayer } = usePlayer();
  
  // Form state
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [selectedCharacterId, setSelectedCharacterId] = useState<string>('');
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [characterData, setCharacterData] = useState<any>(null);
  const [newCharName, setNewCharName] = useState('');
  const [newCharWorld, setNewCharWorld] = useState('');
  const [newCharLevel, setNewCharLevel] = useState(0);
  const [newCharVocation, setNewCharVocation] = useState('');
  const [isCreatingCharacter, setIsCreatingCharacter] = useState(false);
  const [characterError, setCharacterError] = useState<string | null>(null);
  const [displayCharacter, setDisplayCharacter] = useState<any>(null);
  
  // Fetch characters when component mounts
  useEffect(() => {
    if (player) {
      fetchCharacters();
    }
  }, [player, fetchCharacters]);
  
  // Set character ID from URL parameter if provided
  useEffect(() => {
    if (characterIdParam) {
      console.log('Setting character ID from URL parameter:', characterIdParam);
      setSelectedCharacterId(characterIdParam);
    }
  }, [characterIdParam]);
  
  // Set display character data
  useEffect(() => {
    console.log('Running display character effect with:', {
      player: player ? 'exists' : 'null',
      selectedCharacterId,
      charactersCount: characters.length,
      characterData: characterData ? 'exists' : 'null'
    });
    
    // Priority 1: Use new character data from localStorage if it exists and has isNew flag
    if (characterData && characterData.isNew) {
      console.log('PRIORITY 1: Setting display character from localStorage (new character):', characterData);
      setDisplayCharacter({
        name: characterData.name,
        world: characterData.world,
        level: characterData.level,
        vocation: characterData.vocation,
        isNew: true
      });
    }
    // Priority 2: Use existing character selected from dropdown
    else if (player && selectedCharacterId) {
      const selectedChar = characters.find(char => char.id === selectedCharacterId);
      if (selectedChar) {
        console.log('PRIORITY 2: Setting display character from selected character:', selectedChar);
        setDisplayCharacter({
          name: selectedChar.name,
          world: selectedChar.world,
          level: selectedChar.level,
          vocation: selectedChar.vocation,
          isExisting: true
        });
      }
    } 
    // Priority 3: Use any other character data from localStorage
    else if (characterData) {
      console.log('PRIORITY 3: Setting display character from localStorage (not new):', characterData);
      setDisplayCharacter({
        name: characterData.name,
        world: characterData.world,
        level: characterData.level,
        vocation: characterData.vocation,
        isNew: characterData.isNew || false
      });
    } else {
      console.log('No character data available to set display character');
    }
  }, [player, selectedCharacterId, characters, characterData]);
  
  // Component mount effect
  useEffect(() => {
    console.log('Component mounted');
    
    // First, check for character data in localStorage
    const storedCharacterData = localStorage.getItem('characterData');
    if (storedCharacterData) {
      try {
        const parsedData = JSON.parse(storedCharacterData);
        console.log('Found character data in localStorage on mount:', parsedData);
        
        // Set character data and clear any selected character ID
        setCharacterData(parsedData);
        if (parsedData.isNew) {
          console.log('Clearing selectedCharacterId on mount to prioritize new character');
          setSelectedCharacterId('');
        }
      } catch (err) {
        console.error('Failed to parse character data on mount:', err);
      }
    } else if (characterIdParam) {
      console.log('Using character ID from URL parameter on mount:', characterIdParam);
      setSelectedCharacterId(characterIdParam);
    }
  }, []); // Empty dependency array means this runs once on mount
  
  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!name.trim()) {
      setError('Please enter a list name');
      return;
    }
    
    console.log('Form submission - Current state:');
    console.log('- player:', player ? `ID: ${player.id}` : 'null');
    console.log('- selectedCharacterId:', selectedCharacterId);
    console.log('- displayCharacter:', displayCharacter);
    console.log('- characterData:', characterData);
    
    // Ensure we have a display character
    if (!displayCharacter) {
      setError('Character data is missing. Please go back to the home page.');
      return;
    }
    
    setIsCreating(true);
    setError(null);
    
    try {
      let newList: any;
      let requestData: any;
      let tempSessionId: string | null = null;
      
      // Priority 1: If we have a new character (from localStorage), use it
      if (displayCharacter.isNew) {
        console.log('PRIORITY 1: Using new character for list creation:', displayCharacter.name);
        
        if (player) {
          // For logged-in user with new character
          requestData = {
            name,
            description: description.trim() || undefined,
            character_name: displayCharacter.name,
            world: displayCharacter.world,
            player_id: player.id
          };
          console.log('Creating list for logged-in user with new character:', requestData);
        } else {
          // For anonymous user with new character
          tempSessionId = localStorage.getItem('tempSessionId') || Math.random().toString(36).substring(2, 15);
          
          requestData = {
            name,
            description: description.trim() || undefined,
            character_name: displayCharacter.name,
            world: displayCharacter.world,
            session_id: tempSessionId
          };
          console.log('Creating list for anonymous user with new character:', requestData);
          
          // Store the session ID if it's new
          if (!localStorage.getItem('tempSessionId')) {
            localStorage.setItem('tempSessionId', tempSessionId);
          }
        }
      }
      // Priority 2: If we have a selected character ID, use it
      else if (player && selectedCharacterId) {
        console.log('PRIORITY 2: Using existing character with ID:', selectedCharacterId);
        
        requestData = {
          name,
          description: description.trim() || undefined,
          character_id: selectedCharacterId
        };
        console.log('Creating list for logged-in user with existing character:', requestData);
      }
      // Priority 3: Use any other display character data
      else {
        console.log('PRIORITY 3: Using display character for list creation:', displayCharacter.name);
        
        tempSessionId = localStorage.getItem('tempSessionId') || Math.random().toString(36).substring(2, 15);
        
        requestData = {
          name,
          description: description.trim() || undefined,
          character_name: displayCharacter.name,
          world: displayCharacter.world,
          session_id: tempSessionId
        };
        console.log('Creating list for user with display character:', requestData);
        
        // Store the session ID if it's new
        if (!localStorage.getItem('tempSessionId')) {
          localStorage.setItem('tempSessionId', tempSessionId);
        }
      }
      
      // Make the API call with the prepared request data
      newList = await listApi.createList(requestData);
      
      console.log('List created successfully:', newList);
      
      // Clear character data from localStorage after successful creation
      console.log('Clearing character data from localStorage');
      localStorage.removeItem('characterData');
      
      // If this is an anonymous user, fetch the player data before redirecting
      if (!player && tempSessionId) {
        console.log('Fetching anonymous player data with session ID:', tempSessionId);
        await fetchAnonymousPlayer(tempSessionId);
      }
      
      // Navigate to the new list
      router.push(`/lists/${newList.id}`);
    } catch (err: any) {
      console.error('Failed to create list:', err);
      setError(err.message || 'Failed to create list. Please try again.');
    } finally {
      setIsCreating(false);
    }
  };
  
  const handleCreateCharacter = async () => {
    if (!newCharName || !newCharWorld || !newCharLevel || !newCharVocation) {
      setCharacterError('Please fill out all fields');
      return;
    }

    setIsCreatingCharacter(true);
    setCharacterError(null);

    try {
      const newCharacter = {
        name: newCharName,
        world: newCharWorld,
        level: newCharLevel,
        vocation: newCharVocation
      };

      const response = await listApi.createCharacter(newCharacter) as { id?: string };

      if (response && response.id) {
        setSelectedCharacterId(response.id);
        setNewCharName('');
        setNewCharWorld('');
        setNewCharLevel(0);
        setNewCharVocation('');
      } else {
        setCharacterError('Failed to create character');
      }
    } catch (err: any) {
      console.error('Failed to create character:', err);
      setCharacterError(err.message || 'Failed to create character');
    } finally {
      setIsCreatingCharacter(false);
    }
  };
  
  if (playerLoading) {
    return (
      <div className="flex justify-center items-center py-12">
        <svg className="animate-spin h-8 w-8 text-amber-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span className="ml-2 text-amber-800 dark:text-amber-400">Loading...</span>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-amber-200 dark:border-amber-800">
        <div className="flex items-center mb-6">
          <Link
            href="/lists"
            className="mr-2 text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
          >
            <svg className="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
          </Link>
          <h1 className="text-2xl font-bold text-amber-800 dark:text-amber-500">Create Soul Pit List</h1>
        </div>
        
        {!displayCharacter ? (
          <div className="text-center py-8">
            <div className="text-amber-600 dark:text-amber-400 mb-4">
              <svg className="mx-auto h-12 w-12" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-amber-800 dark:text-amber-500 mb-2">Character Information Missing</h3>
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              Please go back to the home page and enter a character name or select a character.
            </p>
            <Link
              href="/"
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
            >
              Back to Lists
            </Link>
          </div>
        ) : (
          <form onSubmit={handleSubmit}>
            {/* Character Information Display */}
            <div className="mb-6 p-4 bg-amber-50 dark:bg-amber-900/30 rounded-md border border-amber-200 dark:border-amber-800">
              <h3 className="text-sm font-medium text-amber-800 dark:text-amber-500 mb-2">Character Information</h3>
              <div className="flex items-center">
                <div className="flex-1">
                  <p className="text-lg font-medium text-gray-800 dark:text-gray-200">{displayCharacter.name}</p>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Level {displayCharacter.level} {displayCharacter.vocation}, {displayCharacter.world}
                  </p>
                </div>
                <span className={`px-2 py-1 text-xs rounded-full ${
                  displayCharacter.isNew 
                    ? 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-300' 
                    : 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-300'
                }`}>
                  {displayCharacter.isNew ? 'New Character' : 'Existing Character'}
                </span>
              </div>
            </div>
            
            {/* List Name */}
            <div className="mb-4">
              <label htmlFor="name" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                List Name*
              </label>
              <input
                type="text"
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Enter a name for your Soul Pit list"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
                disabled={isCreating}
              />
            </div>
            
            {/* List Description */}
            <div className="mb-6">
              <label htmlFor="description" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Description (Optional)
              </label>
              <textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Enter a description for your Soul Pit list"
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
                disabled={isCreating}
              ></textarea>
            </div>
            
            {error && (
              <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/30 rounded-md border border-red-200 dark:border-red-800 text-sm text-red-700 dark:text-red-300">
                {error}
              </div>
            )}
            
            <div className="flex justify-end space-x-3">
              <Link
                href="/"
                className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
              >
                Cancel
              </Link>
              <button
                type="submit"
                disabled={isCreating || !name.trim()}
                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isCreating ? 'Creating...' : 'Create List'}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}

export default function CreateListPage() {
  return (
    <Suspense fallback={<div className="container mx-auto p-4">Loading...</div>}>
      <CreateListContent />
    </Suspense>
  );
}