"use client";

import React, { useState, useEffect, useRef, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { usePlayer } from "@/contexts/PlayerContext";
import { fetchCharacterData } from "@/services/character";
import listApi from "@/services/listApi";

// Create a wrapper component that uses useSearchParams
function JoinListContent() {
  const searchParams = useSearchParams();
  const shareCode = searchParams.get("code") || "";
  
  const router = useRouter();
  const { player, loading: playerLoading, error: playerError, fetchAnonymousPlayer, createPlayer, characters, fetchCharacters } = usePlayer();
  
  // Form state
  const [characterName, setCharacterName] = useState('');
  const [selectedCharacterId, setSelectedCharacterId] = useState<string>('new');
  const [isJoining, setIsJoining] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [listDetails, setListDetails] = useState<any>(null);
  const [isLoadingList, setIsLoadingList] = useState(false);
  const [showCharacterDropdown, setShowCharacterDropdown] = useState(false);
  const [inputFocused, setInputFocused] = useState(false);
  
  // Fetch list details if code is provided
  useEffect(() => {
    const fetchListDetails = async () => {
      if (!shareCode) return;
      
      setIsLoadingList(true);
      try {
        const list = await listApi.getListByShareCode(shareCode);
        setListDetails(list);
      } catch (err) {
        console.error('Failed to fetch list details:', err);
        setError('Invalid or expired share code. Please check and try again.');
      } finally {
        setIsLoadingList(false);
      }
    };
    
    fetchListDetails();
  }, [shareCode]);
  
  // Fetch characters when component mounts
  useEffect(() => {
    if (player) {
      fetchCharacters();
    }
  }, [player, fetchCharacters]);
  
  // Handle character selection
  const handleCharacterSelect = (character: any) => {
    if (character === 'new') {
      setCharacterName('');
      setSelectedCharacterId('new');
    } else {
      setCharacterName(character.name);
      setSelectedCharacterId(character.id);
    }
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
    
    const characters = [...player.characters];
    
    // Filter by name if there's input
    const filtered = characterName.trim() 
      ? characters.filter(character => 
          character.name.toLowerCase().includes(characterName.toLowerCase())
        )
      : characters;
    
    return filtered;
  };
  
  // Check if player is already a member of the list
  const isAlreadyMember = () => {
    if (!player || !listDetails || !listDetails.members) return false;
    
    // Check if any member has the same player ID
    return listDetails.members.some((member: any) => member.player_id === player.id);
  };
  
  // Check if character is already a member of the list
  const isCharacterAlreadyMember = (characterId: string) => {
    if (!listDetails || !listDetails.members) return false;
    
    // Check if any member has the same character ID
    return listDetails.members.some((member: any) => member.character_id === characterId);
  };
  
  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!characterName.trim()) {
      setError('Please enter a character name');
      return;
    }
    
    // Check if the list already has 5 members
    if (listDetails.members && listDetails.members.length >= 5) {
      setError('This list has reached the maximum number of members (5). You cannot join at this time.');
      return;
    }
    
    // Check if player is already a member
    if (player && isAlreadyMember()) {
      setError('You are already a member of this list.');
      return;
    }
    
    setIsJoining(true);
    setError(null);
    
    try {
      if (player) {
        // For logged-in users
        await handleJoinList();
      } else {
        // For anonymous users
        // Check if character exists
        const character = await fetchCharacterData(characterName);
        
        if (!character) {
          setError('Character not found. Please check the name and try again.');
          setIsJoining(false);
          return;
        }
        
        // Check if the character's world matches the list's world
        if (listDetails.world && character.world !== listDetails.world) {
          setError(`This list is for characters in ${listDetails.world} world. Your character is in ${character.world} world.`);
          setIsJoining(false);
          return;
        }
        
        // Character exists, create a temporary session ID
        const tempSessionId = localStorage.getItem('tempSessionId') || Math.random().toString(36).substring(2, 15);
        
        // Join the list as an anonymous user
        await listApi.joinList(shareCode, {
          token: shareCode,
          character_name: character.name,
          world: character.world,
          session_id: tempSessionId
        });
        
        // Store the session ID if it's new
        if (!localStorage.getItem('tempSessionId')) {
          localStorage.setItem('tempSessionId', tempSessionId);
        }
        
        // Fetch the anonymous player to update the player context
        await fetchAnonymousPlayer(tempSessionId);
        
        // Redirect to the list page
        router.push(`/lists/${listDetails.id}`);
      }
    } catch (err: any) {
      console.error('Failed to join list:', err);
      setError(err.message || 'Failed to join list. Please try again.');
    } finally {
      setIsJoining(false);
    }
  };
  
  // Handle joining a list for logged-in users
  const handleJoinList = async () => {
    if (!player || !listDetails) return;
    
    // Check if the list already has 5 members
    if (listDetails.members && listDetails.members.length >= 5) {
      setError('This list has reached the maximum number of members (5). You cannot join at this time.');
      return;
    }
    
    // Check if player is already a member
    if (isAlreadyMember()) {
      setError('You are already a member of this list.');
      return;
    }
    
    // If a character is selected, check if it's already a member
    if (selectedCharacterId !== 'new' && isCharacterAlreadyMember(selectedCharacterId)) {
      setError('This character is already a member of this list.');
      return;
    }
    
    // Get the selected character
    let selectedCharacter;
    if (selectedCharacterId !== 'new' && player.characters) {
      selectedCharacter = player.characters.find(char => char.id === selectedCharacterId);
    }
    
    if (selectedCharacterId === 'new' && !characterName) {
      setError('Please enter a character name.');
      return;
    }
    
    try {
      if (selectedCharacter) {
        // Check if the character's world matches the list's world
        if (listDetails.world && selectedCharacter.world !== listDetails.world) {
          setError(`This list is for characters in ${listDetails.world} world. Your character is in ${selectedCharacter.world} world.`);
          return;
        }
        
        // Join with existing character
        await listApi.joinList(shareCode, {
          token: shareCode,
          player_id: player.id,
          character_id: selectedCharacter.id
        });
      } else {
        // Check if character exists in Tibia
        const character = await fetchCharacterData(characterName);
        
        if (!character) {
          setError('Character not found. Please check the name and try again.');
          return;
        }
        
        // Check if the character's world matches the list's world
        if (listDetails.world && character.world !== listDetails.world) {
          setError(`This list is for characters in ${listDetails.world} world. Your character is in ${character.world} world.`);
          return;
        }
        
        // Join with new character
        await listApi.joinList(shareCode, {
          token: shareCode,
          player_id: player.id,
          character_name: character.name,
          world: character.world
        });
      }
      
      // Redirect to the list page
      router.push(`/lists/${listDetails.id}`);
    } catch (err: any) {
      console.error('Failed to join list:', err);
      throw err;
    }
  };

  if (playerLoading || isLoadingList) {
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
            href="/"
            className="mr-2 text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
          >
            <svg className="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
          </Link>
          <h1 className="text-2xl font-bold text-amber-800 dark:text-amber-500">
            Join List
          </h1>
        </div>
        
        {listDetails && (
          <div className="mb-6 p-4 bg-amber-50 dark:bg-amber-900/30 rounded-md border border-amber-200 dark:border-amber-800">
            <h2 className="text-lg font-medium text-amber-800 dark:text-amber-500 mb-2">
              {listDetails.name}
            </h2>
            {listDetails.description && (
              <p className="text-amber-800 dark:text-amber-400 mb-2">
                {listDetails.description}
              </p>
            )}
            <div className="flex flex-col sm:flex-row sm:justify-between">
              <p className="text-sm text-amber-700 dark:text-amber-300">
                Created by: {listDetails.members?.find((m: any) => m.is_creator)?.character_name || 'Unknown'}
              </p>
              <p className="text-sm font-medium text-amber-700 dark:text-amber-300">
                World: {listDetails.world || 'Any'}
              </p>
            </div>
          </div>
        )}
        
        {error && (
          <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/30 rounded-md border border-red-200 dark:border-red-800">
            <p className="text-red-800 dark:text-red-400">{error}</p>
          </div>
        )}
        
        {isAlreadyMember() ? (
          <div className="mb-6 p-4 bg-green-50 dark:bg-green-900/30 rounded-md border border-green-200 dark:border-green-800">
            <p className="text-green-800 dark:text-green-400">You are already a member of this list.</p>
            <div className="mt-4">
              <button
                onClick={() => router.push(`/lists/${listDetails.id}`)}
                className="w-full px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
              >
                Go to List
              </button>
            </div>
          </div>
        ) : (
          <form onSubmit={handleSubmit}>
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
                    setSelectedCharacterId('new');
                    if (player && player.characters && player.characters.length > 0) {
                      setShowCharacterDropdown(true);
                    }
                  }}
                  onFocus={handleInputFocus}
                  onBlur={handleInputBlur}
                  placeholder="Enter or select a character name"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
                  disabled={isJoining}
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
                  {/* New Character Option */}
                  <div 
                    className={`px-4 py-2 cursor-pointer hover:bg-amber-50 dark:hover:bg-amber-900/30 ${
                      selectedCharacterId === 'new' ? 'bg-amber-100 dark:bg-amber-800/50' : ''
                    }`}
                    onClick={() => handleCharacterSelect('new')}
                  >
                    <div className="font-medium text-amber-800 dark:text-amber-400">
                      + Add New Character
                    </div>
                    <div className="text-xs text-gray-500 dark:text-gray-400">
                      Enter a new character name
                    </div>
                  </div>
                  
                  {/* Divider */}
                  {getFilteredCharacters().length > 0 && (
                    <div className="border-t border-gray-200 dark:border-gray-700 my-1"></div>
                  )}
                  
                  {/* Existing Characters */}
                  {getFilteredCharacters().length > 0 ? (
                    getFilteredCharacters().map(character => (
                      <div 
                        key={character.id}
                        className={`px-4 py-2 cursor-pointer hover:bg-amber-50 dark:hover:bg-amber-900/30 ${
                          selectedCharacterId === character.id ? 'bg-amber-100 dark:bg-amber-800/50' : ''
                        } ${isCharacterAlreadyMember(character.id) ? 'opacity-50' : ''}`}
                        onClick={() => !isCharacterAlreadyMember(character.id) && handleCharacterSelect(character)}
                      >
                        <div className="font-medium text-amber-800 dark:text-amber-400">
                          {character.name}
                          {isCharacterAlreadyMember(character.id) && (
                            <span className="ml-2 text-xs text-red-500">(Already a member)</span>
                          )}
                        </div>
                        <div className="text-xs text-gray-500 dark:text-gray-400">
                          Level {character.level} {character.vocation}, {character.world}
                        </div>
                      </div>
                    ))
                  ) : characterName.trim() ? (
                    <div className="px-4 py-2 text-gray-500 dark:text-gray-400 text-sm">
                      No matching characters found
                    </div>
                  ) : null}
                </div>
              )}
              
              <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                {player && player.characters && player.characters.length > 0 
                  ? "Select an existing character or enter a new character name"
                  : "Enter your Tibia character name to join this list. No account required!"}
              </p>
            </div>
            
            <button
              type="submit"
              disabled={isJoining || !characterName.trim() || isAlreadyMember()}
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
        )}
      </div>
    </div>
  );
}

// Main component with Suspense boundary
export default function JoinListPage() {
  return (
    <Suspense fallback={<div className="container mx-auto p-4">Loading...</div>}>
      <JoinListContent />
    </Suspense>
  );
} 