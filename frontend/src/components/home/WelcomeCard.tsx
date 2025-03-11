'use client';

import { useEffect, useState } from 'react';

interface WelcomeCardProps {
  hasCharacters?: boolean;
}

export default function WelcomeCard({ hasCharacters = false }: WelcomeCardProps) {
  // For client-side rendering, we need to track if characters exist
  const [hasChars, setHasChars] = useState(hasCharacters);

  // In a real implementation, we would fetch this from a store or context
  useEffect(() => {
    // Check localStorage or fetch from API
    const checkForCharacters = async () => {
      // Mock implementation
      const mockCharacters = localStorage.getItem('characters');
      setHasChars(!!mockCharacters);
    };

    checkForCharacters();
  }, []);

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-amber-200 dark:border-amber-800">
      <h1 className="text-2xl font-bold text-amber-800 dark:text-amber-500 mb-4">Welcome to SoulPit Manager</h1>
      <p className="mb-4">
        SoulPit Manager helps you track your progress in Tibia's Soul Pit. 
        Add your character, join a list, and start tracking your soul cores!
      </p>
      
      {!hasChars && (
        <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-4 mb-4">
          <div className="flex items-start">
            <svg className="h-5 w-5 text-amber-500 mr-2 mt-0.5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
            </svg>
            <div>
              <h3 className="text-sm font-medium text-amber-800 dark:text-amber-500">Get Started</h3>
              <p className="text-sm text-amber-700 dark:text-amber-400">Add your Tibia character to begin tracking your soul cores.</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
} 