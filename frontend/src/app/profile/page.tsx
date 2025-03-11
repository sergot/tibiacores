"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { usePlayer } from "@/contexts/PlayerContext";
import { playerApi } from "@/services/api";

export default function ProfilePage() {
  const router = useRouter();
  const { player, characters, fetchCharacters, loading, error, fetchAnonymousPlayer } = usePlayer();
  const [username, setUsername] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [updateError, setUpdateError] = useState("");
  const [showRegistrationTooltip, setShowRegistrationTooltip] = useState(false);
  const hasFetchedRef = useRef(false);
  const hasFetchedCharactersRef = useRef(false);

  useEffect(() => {
    // Try to fetch anonymous player data if player is null
    if (!player && !hasFetchedRef.current) {
      const fetchAnonymousPlayerData = async () => {
        hasFetchedRef.current = true;
        try {
          const tempSessionId = localStorage.getItem('tempSessionId');
          if (tempSessionId) {
            console.log('Fetching anonymous player with session ID:', tempSessionId);
            await fetchAnonymousPlayer(tempSessionId);
          } else {
            console.log('No tempSessionId found in localStorage');
            router.push("/");
          }
        } catch (error) {
          console.error('Error fetching anonymous player:', error);
          router.push("/");
        }
      };
      
      fetchAnonymousPlayerData();
    } else if (player) {
      setUsername(player.username);
      
      // Only fetch characters once
      if (!hasFetchedCharactersRef.current && characters.length === 0) {
        hasFetchedCharactersRef.current = true;
        fetchCharacters();
      }
    }
  }, [player, characters, fetchCharacters, fetchAnonymousPlayer, router]);

  const handleUsernameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setUsername(e.target.value);
  };

  const handleUpdateUsername = async () => {
    if (!username.trim()) {
      setUpdateError("Username cannot be empty");
      return;
    }

    if (!player) {
      setUpdateError("No player found");
      return;
    }

    try {
      await playerApi.updateUsername(player.id, username);
      setIsEditing(false);
      setUpdateError("");
      // Refresh player data
      window.location.reload();
    } catch (error: any) {
      console.error("Error updating username:", error);
      setUpdateError(error.message || "Failed to update username");
    }
  };

  const handleRegisterClick = () => {
    // This will be implemented in the future
    alert("Registration functionality will be available soon!");
  };

  if (loading) {
    return (
      <div className="container mx-auto p-4">
        <h1 className="text-2xl font-bold mb-4">Profile</h1>
        <p>Loading...</p>
      </div>
    );
  }

  if (error || !player) {
    return (
      <div className="container mx-auto p-4">
        <h1 className="text-2xl font-bold mb-4">Profile</h1>
        <p className="text-red-500">Error loading profile. Please try again later.</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-2xl font-bold mb-6">Your Profile</h1>
      
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4">Account Information</h2>
        
        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Username</label>
          {isEditing ? (
            <div className="flex items-center">
              <input
                type="text"
                value={username}
                onChange={handleUsernameChange}
                className="border rounded-md px-3 py-2 mr-2 flex-grow dark:bg-gray-700 dark:border-gray-600"
              />
              <button
                onClick={handleUpdateUsername}
                className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-md"
              >
                Save
              </button>
              <button
                onClick={() => {
                  setIsEditing(false);
                  setUsername(player.username);
                }}
                className="bg-gray-300 hover:bg-gray-400 text-gray-800 px-4 py-2 rounded-md ml-2"
              >
                Cancel
              </button>
            </div>
          ) : (
            <div className="flex items-center">
              <span className="text-lg">{player.username}</span>
              <button
                onClick={() => setIsEditing(true)}
                className="ml-3 text-blue-500 hover:text-blue-600"
              >
                Edit
              </button>
            </div>
          )}
          {updateError && <p className="text-red-500 mt-1">{updateError}</p>}
        </div>
        
        {player.is_anonymous && (
          <div 
            className="bg-yellow-100 border-l-4 border-yellow-500 text-yellow-700 p-4 relative cursor-pointer"
            onMouseEnter={() => setShowRegistrationTooltip(true)}
            onMouseLeave={() => setShowRegistrationTooltip(false)}
            onClick={handleRegisterClick}
          >
            <p className="font-medium">You're using an anonymous account</p>
            <p>Register to save your progress and access from any device.</p>
            
            {showRegistrationTooltip && (
              <div className="absolute z-10 w-64 p-3 bg-white dark:bg-gray-700 rounded-md shadow-lg border border-gray-200 dark:border-gray-600 -top-2 left-full ml-2">
                <p className="text-sm">
                  Create an account to:
                </p>
                <ul className="list-disc ml-5 mt-1 text-sm">
                  <li>Save your progress</li>
                  <li>Access from any device</li>
                  <li>Never lose your characters</li>
                </ul>
              </div>
            )}
          </div>
        )}
      </div>
      
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4">Your Characters</h2>
        
        {characters && characters.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {characters.map((character) => (
              <div 
                key={character.id} 
                className="border rounded-md p-4 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer transition-colors"
                onClick={() => router.push(`/profile/characters/${character.id}`)}
              >
                <div className="flex items-center mb-2">
                  <span className="font-medium text-lg">{character.name}</span>
                  {player && player.characters && player.characters.length > 0 && 
                    player.characters[0]?.id === character.id && (
                    <span className="ml-2 bg-green-100 text-green-800 text-xs px-2 py-1 rounded-full dark:bg-green-900 dark:text-green-100">
                      Main
                    </span>
                  )}
                </div>
                <p className="text-gray-600 dark:text-gray-400">World: {character.world}</p>
              </div>
            ))}
          </div>
        ) : (
          <p>No characters found.</p>
        )}
      </div>
    </div>
  );
} 