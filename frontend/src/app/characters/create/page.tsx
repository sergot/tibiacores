'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { usePlayer } from '@/contexts/PlayerContext';
import { fetchCharacterData } from '@/services/character';

// List of Tibia worlds
const tibiaWorlds = [
  'Antica', 'Secura', 'Monza', 'Premia', 'Harmonia',
  'Peloria', 'Marcia', 'Refugia', 'Vunira', 'Zuna',
  'Zunera', 'Kalibra', 'Menera', 'Celebra', 'Firmera',
  'Helera', 'Serdebra', 'Solidera', 'Venebra', 'Wintera'
].sort();

// Tibia vocations
const vocations = ['Knight', 'Paladin', 'Sorcerer', 'Druid'];

export default function CreateCharacterPage() {
  const router = useRouter();
  const { player, addCharacter, loading: playerLoading, error: playerError } = usePlayer();
  
  const [name, setName] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isCheckingCharacter, setIsCheckingCharacter] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [nameError, setNameError] = useState<string | null>(null);
  const [characterData, setCharacterData] = useState<{ world: string; level: number; vocation: string } | null>(null);

  // Validate character name
  const validateName = (value: string) => {
    if (!value.trim()) {
      setNameError('Character name is required');
      return false;
    }
    
    if (value.length < 3 || value.length > 20) {
      setNameError('Character name must be between 3 and 20 characters');
      return false;
    }
    
    // Check if name follows Tibia naming rules (simplified)
    const nameRegex = /^[A-Z][a-z]+(?: [A-Z][a-z]+)*$/;
    if (!nameRegex.test(value)) {
      setNameError('Character name must start with a capital letter and can contain spaces between words');
      return false;
    }
    
    setNameError(null);
    return true;
  };

  // Handle character name input change
  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setName(e.target.value);
    setCharacterData(null); // Reset character data when name changes
    if (nameError) validateName(e.target.value);
  };

  // Check character data from TibiaData API
  const checkCharacter = async () => {
    if (!validateName(name)) {
      return;
    }

    setIsCheckingCharacter(true);
    setError(null);
    setCharacterData(null);

    try {
      const character = await fetchCharacterData(name);
      
      if (!character) {
        setError('Character not found. Please check the name and try again.');
        return;
      }
      
      setCharacterData({
        world: character.world,
        level: character.level,
        vocation: character.vocation
      });
    } catch (err: any) {
      console.error('Failed to fetch character data:', err);
      setError('Failed to fetch character data. Please try again.');
    } finally {
      setIsCheckingCharacter(false);
    }
  };

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateName(name)) {
      return;
    }

    if (!characterData) {
      // Check character first if not already checked
      await checkCharacter();
      if (!characterData) return; // Exit if character check failed
    }
    
    setIsLoading(true);
    setError(null);
    
    try {
      // Check if player exists
      if (!player) {
        throw new Error('You need to create an account first');
      }
      
      // Add character using the API
      await addCharacter({
        name,
        world: characterData.world,
        level: characterData.level,
        vocation: characterData.vocation
      });
      
      // Navigate back to characters list
      router.push('/characters');
    } catch (err: any) {
      console.error('Failed to create character:', err);
      setError(err.message || 'Failed to create character. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-amber-200 dark:border-amber-800">
        <div className="flex items-center mb-6">
          <Link
            href="/characters"
            className="mr-2 text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
          >
            <svg className="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
          </Link>
          <h1 className="text-2xl font-bold text-amber-800 dark:text-amber-500">Add New Character</h1>
        </div>
        
        {!player && (
          <div className="mb-6 p-4 bg-amber-50 dark:bg-amber-900/30 rounded-md border border-amber-200 dark:border-amber-800">
            <p className="text-amber-800 dark:text-amber-400">
              You need to create an account before adding characters. Please go to the home page to set up your account.
            </p>
            <Link
              href="/"
              className="mt-2 inline-flex items-center px-4 py-2 rounded-md text-white font-medium bg-amber-600 hover:bg-amber-700"
            >
              Create Account
            </Link>
          </div>
        )}
        
        <form onSubmit={handleSubmit}>
          <div className="mb-6">
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Character Name*
            </label>
            <div className="flex">
              <input
                type="text"
                id="name"
                value={name}
                onChange={handleNameChange}
                onBlur={(e) => validateName(e.target.value)}
                placeholder="Enter character name"
                className={`flex-1 px-3 py-2 border ${nameError ? 'border-red-300 dark:border-red-600' : 'border-gray-300 dark:border-gray-600'} rounded-l-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white`}
                disabled={!player || isLoading}
              />
              <button
                type="button"
                onClick={checkCharacter}
                disabled={isCheckingCharacter || !name.trim() || !!nameError || !player || isLoading}
                className="px-4 py-2 border border-transparent rounded-r-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isCheckingCharacter ? 'Checking...' : 'Check'}
              </button>
            </div>
            {nameError && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">{nameError}</p>
            )}
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              Enter your Tibia character name to automatically fetch world, level, and vocation.
            </p>
          </div>
          
          {characterData && (
            <div className="mb-6 p-4 bg-green-50 dark:bg-green-900/30 rounded-md border border-green-200 dark:border-green-800">
              <p className="text-sm text-green-700 dark:text-green-300 font-medium">Character found!</p>
              <div className="mt-1 text-sm text-green-600 dark:text-green-400">
                <p>World: {characterData.world}</p>
                <p>Level: {characterData.level}</p>
                <p>Vocation: {characterData.vocation}</p>
              </div>
            </div>
          )}
          
          {(error || playerError) && (
            <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/30 rounded-md border border-red-200 dark:border-red-800 text-sm text-red-700 dark:text-red-300">
              {error || playerError}
            </div>
          )}
          
          <div className="flex justify-end space-x-2">
            <Link
              href="/characters"
              className="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
            >
              Cancel
            </Link>
            <button
              type="submit"
              disabled={isLoading || isCheckingCharacter || !name.trim() || !!nameError || !player || !characterData}
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? (
                <>
                  <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Creating...
                </>
              ) : (
                'Create Character'
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
} 