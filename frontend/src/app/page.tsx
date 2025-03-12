'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { usePlayer } from '@/contexts/PlayerContext';
import { fetchCharacterData } from '@/services/character';
import listApi from '@/services/listApi';
import { playerApi } from '@/services/api';

// Define list interface
interface List {
  id: string;
  name: string;
  description?: string;
  world: string;
  share_code: string;
  created_at: string;
  is_creator?: boolean;
  member_count?: number;
  character_name?: string;
  raw_data?: {
    creator_id: string;
    player_id_match?: boolean;
    member_is_creator: boolean;
    session_id?: string;
    member_session_id?: string;
  };
}

export default function Home() {
  const router = useRouter();
  const { player, loading: playerLoading, fetchAnonymousPlayer, isAnonymous } = usePlayer();
  
  // Character creation state
  const [characterName, setCharacterName] = useState('');
  const [characterError, setCharacterError] = useState<string | null>(null);
  const [selectedCharacterId, setSelectedCharacterId] = useState<string>('');
  const [showCharacterDropdown, setShowCharacterDropdown] = useState(false);
  const [inputFocused, setInputFocused] = useState(false);
  const characterInputRef = useState<HTMLInputElement | null>(null);
  
  // Join list state
  const [joinCode, setJoinCode] = useState('');
  const [joinError, setJoinError] = useState<string | null>(null);
  const [isJoining, setIsJoining] = useState(false);
  
  // Loading states
  const [isCreating, setIsCreating] = useState(false);
  
  // User lists state
  const [userLists, setUserLists] = useState<List[]>([]);
  const [listsLoading, setListsLoading] = useState(false);
  const [listsError, setListsError] = useState<string | null>(null);
  
  // Check for session ID on mount and try to fetch anonymous player
  useEffect(() => {
    const fetchPlayerOnce = async () => {
      if (!player) {
        const tempSessionId = localStorage.getItem('tempSessionId');
        // Only try to fetch anonymous player if we have a session ID (user has created a list)
        if (tempSessionId) {
          try {
            await fetchAnonymousPlayer(tempSessionId);
          } catch (err) {
            console.error('Failed to fetch anonymous player:', err);
          }
        }
      }
    };
    
    fetchPlayerOnce();
  }, [player, fetchAnonymousPlayer]);
  
  // Fetch user lists on component mount
  useEffect(() => {
    const fetchUserLists = async () => {
      setListsLoading(true);
      setListsError(null);
      
      try {
        let fetchedLists: List[] = [];
        let rawListsData: any[] = [];
        
        if (player) {
          // Fetch lists for logged-in user
          const response = await listApi.getLists(player.id);
          console.log('Raw lists data for logged-in user:', response);
          
          // Make sure response is an array
          rawListsData = Array.isArray(response) ? response : [];
          
          // Process lists to add isCreator flag
          fetchedLists = rawListsData.map(list => {
            // Check if the current player is the creator by ID
            console.log('single List:', list);
            const isCreatorById = list.creator_id === player.id;
            
            // Find the member entry for the current player
            const memberInfo = list.members.find(
              (member: any) => member.player_id === player.id
            );
            
            // Check if the player is marked as creator in the members array
            const isCreatorByMember = memberInfo?.is_creator === true;
            
            console.log('List:', list.name);
            console.log('Creator ID:', list.creator_id, 'Player ID:', player.id, 'Match:', isCreatorById);
            console.log('Member info:', memberInfo);
            console.log('Is creator from member info:', isCreatorByMember);
            
            return {
              id: list.id,
              name: list.name,
              description: list.description,
              world: list.world,
              share_code: list.share_code,
              created_at: list.created_at,
              // User is creator if they are the creator by ID
              is_creator: isCreatorById,
              member_count: list.members.length,
              // Get character name from member info
              character_name: memberInfo?.character_name || 'Unknown',
              // Store raw data for debugging
              raw_data: {
                creator_id: list.creator_id,
                player_id_match: isCreatorById,
                member_is_creator: isCreatorByMember
              }
            };
          });
        } else {
          // Check if we have a temporary session ID
          const tempSessionId = localStorage.getItem('tempSessionId');
          if (tempSessionId) {
            try {
              // Try to fetch the anonymous player first
              await fetchAnonymousPlayer(tempSessionId);
              
              // If we now have a player, fetch their lists
              const currentPlayer = await playerApi.getPlayerBySession(tempSessionId);
              if (currentPlayer) {
                const response = await listApi.getLists(currentPlayer.id);
                console.log('Raw lists data for anonymous user (via player):', response);
                
                // Make sure response is an array
                rawListsData = Array.isArray(response) ? response : [];
                
                // Process lists to add isCreator flag
                fetchedLists = rawListsData.map(list => {
                  // Check if the current player is the creator by ID
                  const isCreatorById = list.creator_id === currentPlayer.id;
                  
                  // Find the member with this session ID
                  const memberInfo = list.members.find(
                    (member: any) => member.player_id === currentPlayer.id
                  );
                  
                  console.log('List:', list.name);
                  console.log('Creator ID:', list.creator_id, 'Player ID:', currentPlayer.id, 'Match:', isCreatorById);
                  console.log('Member info:', memberInfo);
                  console.log('Is creator from member info:', memberInfo?.is_creator);
                  
                  return {
                    id: list.id,
                    name: list.name,
                    description: list.description,
                    world: list.world,
                    share_code: list.share_code,
                    created_at: list.created_at,
                    // User is creator if they are the creator by ID
                    is_creator: isCreatorById,
                    member_count: list.members.length,
                    // Get character name from member info
                    character_name: memberInfo?.character_name || 'Unknown',
                    // Store raw data for debugging
                    raw_data: {
                      creator_id: list.creator_id,
                      player_id_match: isCreatorById,
                      member_is_creator: memberInfo?.is_creator,
                      session_id: tempSessionId,
                      member_session_id: memberInfo?.session_id
                    }
                  };
                });
              } else {
                console.log('No player found for session ID:', tempSessionId);
              }
            } catch (err: any) {
              console.error('Error fetching anonymous player or lists:', err);
              // Don't set an error, just log it and continue with an empty list
            }
          }
        }
        
        console.log('Processed lists:', fetchedLists);
        setUserLists(fetchedLists);
        
        // Store raw data in localStorage for debugging
        localStorage.setItem('debugRawListsData', JSON.stringify(rawListsData));
      } catch (err: any) {
        console.error('Failed to fetch lists:', err);
        setListsError(err.message || 'Failed to fetch lists. Please try again.');
      } finally {
        setListsLoading(false);
      }
    };
    
    fetchUserLists();
  }, [player]);
  
  // Handle create list form submission
  const handleCreateList = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // If player exists and a valid character ID is selected
    if (player && selectedCharacterId) {
      // Find the selected character to pass its details
      const selectedCharacter = player.characters.find(char => char.id === selectedCharacterId);
      if (selectedCharacter) {
        console.log('Redirecting to create list page with character ID:', selectedCharacterId);
        
        // Clear any stored character data to avoid conflicts
        localStorage.removeItem('characterData');
        
        router.push(`/lists/create?character_id=${selectedCharacterId}`);
        return;
      }
    }
    
    // For new character entry (both anonymous and logged-in users)
    if (!characterName.trim()) {
      setCharacterError('Please enter a character name');
      return;
    }
    
    setIsCreating(true);
    setCharacterError(null);
    
    try {
      console.log('Validating character name:', characterName);
      
      // Validate character exists in TibiaData API
      const character = await fetchCharacterData(characterName);
      
      if (!character) {
        setCharacterError('Character not found. Please check the name and try again.');
        setIsCreating(false);
        return;
      }
      
      console.log('Character validated successfully:', character);
      
      // Create a temporary session ID if user doesn't have an account
      const sessionId = localStorage.getItem('tempSessionId') || Math.random().toString(36).substring(2, 15);
      localStorage.setItem('tempSessionId', sessionId);
      
      // Clear any previously stored character data to avoid conflicts
      localStorage.removeItem('characterData');
      
      // Store character data in localStorage for later use
      const characterData = {
        name: character.name,
        world: character.world,
        level: character.level,
        vocation: character.vocation,
        isNew: true // Flag to indicate this is a new character
      };
      
      console.log('Storing character data in localStorage:', characterData);
      localStorage.setItem('characterData', JSON.stringify(characterData));
      
      // Redirect to the create list page with character data
      console.log('Redirecting to create list page with new character');
      router.push('/lists/create');
      
      // Try to fetch the anonymous player after a short delay to allow the backend to create it
      if (!player) {
        setTimeout(() => {
          fetchAnonymousPlayer(sessionId).catch(err => {
            console.error('Failed to fetch anonymous player after list creation:', err);
          });
        }, 1000);
      }
    } catch (err: any) {
      console.error('Failed to create list:', err);
      setCharacterError(err.message || 'Failed to create list. Please try again.');
    } finally {
      setIsCreating(false);
    }
  };

  // Handle character selection from dropdown
  const handleCharacterSelect = (character: any) => {
    setSelectedCharacterId(character.id);
    setCharacterName(character.name);
    setShowCharacterDropdown(false);
  };

  // Handle input focus
  const handleInputFocus = () => {
    setInputFocused(true);
    if (player && player.characters && player.characters.length > 0) {
      setShowCharacterDropdown(true);
    }
  };

  // Handle input blur
  const handleInputBlur = () => {
    // Delay hiding the dropdown to allow for click events to register
    setTimeout(() => {
      setInputFocused(false);
      setShowCharacterDropdown(false);
    }, 200);
  };

  // Filter characters based on input
  const getFilteredCharacters = () => {
    if (!player || !player.characters) return [];
    
    if (!characterName.trim()) {
      return player.characters;
    }
    
    return player.characters.filter(character => 
      character.name.toLowerCase().includes(characterName.toLowerCase())
    );
  };

  // Handle join list form submission
  const handleJoinList = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!joinCode.trim()) {
      setJoinError('Please enter a join code');
      return;
    }
    
    setIsJoining(true);
    setJoinError(null);
    
    try {
      // If player exists, redirect to join page with code
      if (player) {
        router.push(`/lists/join?code=${encodeURIComponent(joinCode)}`);
        return;
      }
      
      // Otherwise, redirect to join page with code
      router.push(`/lists/join?code=${encodeURIComponent(joinCode)}&noAccount=true`);
    } catch (err: any) {
      console.error('Failed to join list:', err);
      setJoinError(err.message || 'Failed to join list. Please try again.');
    } finally {
      setIsJoining(false);
    }
  };

  // Set default character selection
  useEffect(() => {
    if (player && player.characters && player.characters.length > 0) {
      // Only set default if currently set to "new" (initial state)
      if (selectedCharacterId === 'new') {
        setSelectedCharacterId(player.characters[0].id);
      }
    } else {
      setSelectedCharacterId('new');
    }
  }, [player, selectedCharacterId]);

  return (
    <div className="max-w-4xl mx-auto">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 border border-amber-200 dark:border-amber-800">
        <h1 className="text-3xl font-bold text-amber-800 dark:text-amber-500 mb-6 text-center">
          Welcome to SoulPit Manager
        </h1>
        
        <div className="text-center mb-8">
          <p className="text-gray-600 dark:text-gray-400 mb-2">
            Track your Soul Pit progress with friends and guildmates
          </p>
          <div className="flex justify-center">
            <span className="inline-block px-3 py-1 bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-300 rounded-full text-sm font-medium">
              Tibia Soul Pit Tracking Tool
            </span>
          </div>
        </div>
        
        {/* Create and Join List Section */}
        <div className="grid md:grid-cols-2 gap-8 mb-8">
          {/* Create List Section */}
          <div className="bg-amber-50 dark:bg-amber-900/30 rounded-lg p-6 border border-amber-200 dark:border-amber-800">
            <h2 className="text-xl font-semibold text-amber-800 dark:text-amber-500 mb-4">
              Create a Soul Pit List
            </h2>
            
            <form onSubmit={handleCreateList}>
              <div className="mb-4 relative">
                <label htmlFor="characterName" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Character*
                </label>
                
                <div className="relative">
                  <input
                    type="text"
                    id="characterName"
                    value={characterName}
                    onChange={(e) => {
                      setCharacterName(e.target.value);
                      setSelectedCharacterId('');
                      if (player && player.characters && player.characters.length > 0) {
                        setShowCharacterDropdown(true);
                      }
                    }}
                    onFocus={handleInputFocus}
                    onBlur={handleInputBlur}
                    placeholder="Enter or select a character name"
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
                    disabled={isCreating}
                  />
                  
                  {player && player.characters && player.characters.length > 0 && (
                    <button 
                      type="button"
                      className="absolute right-2 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                      onClick={() => setShowCharacterDropdown(!showCharacterDropdown)}
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clipRule="evenodd" />
                      </svg>
                    </button>
                  )}
                </div>
                
                {player && player.characters && player.characters.length > 0 && showCharacterDropdown && (
                  <div className="absolute z-10 mt-1 w-full bg-white dark:bg-gray-800 shadow-lg rounded-md border border-gray-200 dark:border-gray-700 max-h-60 overflow-auto">
                    {getFilteredCharacters().length > 0 ? (
                      getFilteredCharacters().map(character => (
                        <div 
                          key={character.id}
                          className={`px-4 py-2 cursor-pointer hover:bg-amber-50 dark:hover:bg-amber-900/30 ${
                            selectedCharacterId === character.id ? 'bg-amber-100 dark:bg-amber-800/50' : ''
                          }`}
                          onClick={() => handleCharacterSelect(character)}
                        >
                          <div className="font-medium text-amber-800 dark:text-amber-400">{character.name}</div>
                          <div className="text-xs text-gray-500 dark:text-gray-400">
                            Level {character.level} {character.vocation}, {character.world}
                          </div>
                        </div>
                      ))
                    ) : (
                      <div className="px-4 py-2 text-gray-500 dark:text-gray-400 text-sm">
                        No matching characters found
                      </div>
                    )}
                  </div>
                )}
                
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {player && player.characters && player.characters.length > 0 
                    ? "Type to search or enter a new character name"
                    : "Enter your Tibia character name to create a Soul Pit list. No account required!"}
                </p>
              </div>
              
              {characterError && (
                <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/30 rounded-md border border-red-200 dark:border-red-800 text-sm text-red-700 dark:text-red-300">
                  {characterError}
                </div>
              )}
              
              <button
                type="submit"
                disabled={isCreating || !characterName.trim()}
                className="w-full px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isCreating ? (
                  <span className="flex items-center justify-center">
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Creating...
                  </span>
                ) : (
                  'Create Soul Pit List'
                )}
              </button>
            </form>
          </div>
          
          {/* Join List Section */}
          <div className="bg-amber-50 dark:bg-amber-900/30 rounded-lg p-6 border border-amber-200 dark:border-amber-800">
            <h2 className="text-xl font-semibold text-amber-800 dark:text-amber-500 mb-4">
              Join a Soul Pit List
            </h2>
            
            <form onSubmit={handleJoinList}>
              <div className="mb-4">
                <label htmlFor="joinCode" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Join Code or URL*
                </label>
                <input
                  type="text"
                  id="joinCode"
                  value={joinCode}
                  onChange={(e) => setJoinCode(e.target.value)}
                  placeholder="Enter join code or paste URL"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
                  disabled={isJoining}
                />
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  Enter the join code or URL shared with you
                </p>
              </div>
              
              {joinError && (
                <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/30 rounded-md border border-red-200 dark:border-red-800 text-sm text-red-700 dark:text-red-300">
                  {joinError}
                </div>
              )}
              
              <button
                type="submit"
                disabled={isJoining || !joinCode.trim()}
                className="w-full px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isJoining ? (
                  <span className="flex items-center justify-center">
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Joining...
                  </span>
                ) : (
                  'Join List'
                )}
              </button>
            </form>
          </div>
        </div>
        
        {/* User Lists Section */}
        {(userLists.length > 0 || player) && (
          <div className="mb-8">
            <div className="bg-amber-50 dark:bg-amber-900/30 rounded-lg p-6 border border-amber-200 dark:border-amber-800">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-semibold text-amber-800 dark:text-amber-500">
                  Your Soul Pit Lists
                </h2>
              </div>
              
              {/* Debug Information */}
              <div className="mb-4 p-3 bg-gray-100 dark:bg-gray-800 rounded-md text-xs font-mono overflow-auto">
                <h3 className="font-bold mb-2">Debug Info:</h3>
                <div>
                  <p><strong>Player:</strong> {player ? `ID: ${player.id}, Username: ${player.username}` : 'Anonymous'}</p>
                  {player && player.session_id && <p><strong>Player Session ID:</strong> {player.session_id}</p>}
                  {player && <p><strong>DB Anonymous Flag:</strong> {player.is_anonymous ? 'Yes' : 'No'}</p>}
                  {!player && <p><strong>Session ID:</strong> {localStorage.getItem('tempSessionId') || 'None'}</p>}
                  <p><strong>Session Auth Method:</strong> {isAnonymous ? 'Session ID' : 'Account Login'}</p>
                  <p><strong>Characters:</strong> {player ? player.characters.map(c => `${c.name} (${c.world})`).join(', ') : 'None'}</p>
                  <p><strong>Lists Count:</strong> {userLists.length}</p>
                  {userLists.length > 0 && (
                    <div className="mt-2">
                      <p><strong>Lists:</strong></p>
                      <ul className="pl-4 list-disc">
                        {userLists.map(list => (
                          <li key={list.id}>
                            {list.name} - {list.is_creator ? 'Owner' : 'Member'}
                            {list.raw_data && (
                              <div className="pl-4 text-gray-500 dark:text-gray-400">
                                <p>Creator ID: {list.raw_data.creator_id}</p>
                                {player && (
                                  <>
                                    <p>Player ID Match: {list.raw_data.player_id_match ? 'Yes' : 'No'}</p>
                                    <p>Member is Creator: {list.raw_data.member_is_creator ? 'Yes' : 'No'}</p>
                                  </>
                                )}
                                {!player && (
                                  <>
                                    <p>Session ID: {list.raw_data.session_id}</p>
                                    <p>Member Session ID: {list.raw_data.member_session_id}</p>
                                    <p>Member is Creator: {list.raw_data.member_is_creator ? 'Yes' : 'No'}</p>
                                  </>
                                )}
                              </div>
                            )}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              </div>
              
              {listsLoading ? (
                <div className="flex justify-center py-8">
                  <svg className="animate-spin h-8 w-8 text-amber-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                </div>
              ) : listsError ? (
                <div className="p-4 bg-red-50 dark:bg-red-900/30 rounded-md border border-red-200 dark:border-red-800 text-sm text-red-700 dark:text-red-300">
                  {listsError}
                </div>
              ) : userLists.length === 0 ? (
                <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                  <p className="mb-4">You haven't created or joined any lists yet.</p>
                  <Link
                    href="/lists/create"
                    className="px-4 py-2 bg-amber-600 hover:bg-amber-700 text-white rounded-md"
                  >
                    Create Your First List
                  </Link>
                </div>
              ) : (
                <div className="grid gap-4">
                  {userLists.map((list) => (
                    <Link 
                      key={list.id} 
                      href={`/lists/${list.id}`}
                      className="block p-4 bg-white dark:bg-gray-700 rounded-md border border-amber-100 dark:border-amber-800 hover:border-amber-300 dark:hover:border-amber-600 transition-colors"
                    >
                      <div className="flex justify-between items-start">
                        <div>
                          <h3 className="font-medium text-amber-800 dark:text-amber-400">{list.name}</h3>
                          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                            Character: {list.character_name} • World: {list.world} • Created: {new Date(list.created_at).toLocaleDateString()}
                          </p>
                          {list.description && (
                            <p className="text-sm text-gray-600 dark:text-gray-300 mt-2 line-clamp-2">
                              {list.description}
                            </p>
                          )}
                        </div>
                        <span className={`px-2 py-1 text-xs rounded-full ${
                          list.is_creator 
                            ? 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-300 font-medium' 
                            : 'bg-amber-100 dark:bg-amber-900 text-amber-800 dark:text-amber-300'
                        }`}>
                          {list.is_creator ? 'Owner' : 'Member'}
                        </span>
                      </div>
                    </Link>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
        
        {player ? (
          isAnonymous ? (
            <div className="mb-8 bg-blue-50 dark:bg-blue-900/20 rounded-lg p-6 border border-blue-200 dark:border-blue-800">
              <h3 className="text-lg font-semibold text-blue-800 dark:text-blue-400 mb-2">
                Save Your Progress
              </h3>
              <p className="text-gray-600 dark:text-gray-400 mb-4">
                You're currently using a temporary account. Register now to keep your progress and access more features!
              </p>
              <div className="flex flex-col sm:flex-row gap-3">
                <Link
                  href="/register"
                  className="inline-flex items-center justify-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                  Register Account
                </Link>
                <Link
                  href="/login"
                  className="inline-flex items-center justify-center px-4 py-2 border border-blue-300 dark:border-blue-700 rounded-md shadow-sm text-sm font-medium text-blue-700 dark:text-blue-300 bg-white dark:bg-gray-800 hover:bg-blue-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
                  </svg>
                  Log In
                </Link>
              </div>
            </div>
          ) : (
            <></>
          )
        ) : null}
        
        <div className="mt-8 pt-6 border-t border-gray-200 dark:border-gray-700">
          <h3 className="text-lg font-semibold text-amber-800 dark:text-amber-500 mb-3">
            About SoulPit Manager
          </h3>
          <p className="text-gray-600 dark:text-gray-400 mb-2">
            SoulPit Manager helps Tibia players track their Soul Pit progress, including:
          </p>
          <ul className="list-disc list-inside text-gray-600 dark:text-gray-400 mb-4 space-y-1">
            <li>Track which soul cores you've collected</li>
            <li>Share your progress with friends and guildmates</li>
            <li>Coordinate hunting efforts for missing creatures</li>
            <li>Get notifications when new creatures are added</li>
          </ul>
          <p className="text-gray-600 dark:text-gray-400 text-sm">
            SoulPit Manager is a fan-made tool and is not affiliated with CipSoft GmbH.
          </p>
        </div>
      </div>
    </div>
  );
}
