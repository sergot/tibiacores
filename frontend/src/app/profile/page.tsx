"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { usePlayer } from "@/contexts/PlayerContext";
import { playerApi } from "@/services/api";
import Link from 'next/link';

export default function ProfilePage() {
  const router = useRouter();
  const { player, characters, fetchCharacters, loading, error, fetchAnonymousPlayer, isAnonymous } = usePlayer();
  const [username, setUsername] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [updateError, setUpdateError] = useState("");
  const [updateSuccess, setUpdateSuccess] = useState("");
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
          // Simply log the error and redirect
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
      setUpdateSuccess("Username updated successfully!");
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
        
        {updateSuccess && (
          <div className="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded relative mb-4" role="alert">
            <span className="block sm:inline">{updateSuccess}</span>
          </div>
        )}
        
        {updateError && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4" role="alert">
            <span className="block sm:inline">{updateError}</span>
          </div>
        )}
        
        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Username</label>
          {isEditing ? (
            <form onSubmit={handleUpdateUsername} className="mb-4">
              <div className="flex flex-col md:flex-row gap-2">
                <input
                  type="text"
                  value={username}
                  onChange={handleUsernameChange}
                  className="flex-grow px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-amber-500 dark:bg-gray-700 dark:text-white"
                  placeholder="Enter new username"
                  minLength={3}
                  maxLength={30}
                  required
                />
                <div className="flex gap-2">
                  <button
                    type="submit"
                    className="px-4 py-2 bg-amber-500 hover:bg-amber-600 text-white rounded-md transition-colors"
                  >
                    Save
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setIsEditing(false);
                      setUsername(player.username);
                      setUpdateError('');
                    }}
                    className="px-4 py-2 bg-gray-300 hover:bg-gray-400 dark:bg-gray-600 dark:hover:bg-gray-500 text-gray-800 dark:text-white rounded-md transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            </form>
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
        </div>
        
        <div className="mb-4">
          <p className="text-sm text-gray-500 dark:text-gray-400">Account Type</p>
          <p className="text-lg font-medium text-gray-800 dark:text-gray-200">
            {isAnonymous ? 'Anonymous (Not Registered)' : 'Registered'}
          </p>
        </div>
        
        {isAnonymous && (
          <div className="bg-amber-50 dark:bg-amber-900/30 border border-amber-200 dark:border-amber-800 rounded-lg p-4 mt-4">
            <h3 className="font-medium text-amber-800 dark:text-amber-400 mb-2">Register Your Account</h3>
            <p className="text-amber-700 dark:text-amber-300 mb-3 text-sm">
              Your progress is currently stored on this device only. Register to:
            </p>
            <ul className="list-disc list-inside text-amber-700 dark:text-amber-300 text-sm mb-4">
              <li>Access your characters from any device</li>
              <li>Never lose your progress</li>
              <li>Get updates about new features</li>
            </ul>
            <Link 
              href="/register" 
              className="inline-block px-4 py-2 bg-amber-500 hover:bg-amber-600 text-white rounded-md transition-colors"
            >
              Register Now
            </Link>
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