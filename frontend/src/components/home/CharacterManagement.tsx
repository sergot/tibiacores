'use client';

import { useState, useEffect } from 'react';

interface Character {
  id: string;
  name: string;
  world: string;
  level: number;
  vocation: string;
}

export default function CharacterManagement() {
  const [characters, setCharacters] = useState<Character[]>([]);
  const [selectedCharacter, setSelectedCharacter] = useState<Character | null>(null);
  const [newCharacterName, setNewCharacterName] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load characters on component mount
  useEffect(() => {
    const fetchCharacters = async () => {
      try {
        // In a real implementation, we would fetch from the API
        // For now, we'll use mock data
        const storedCharacters = localStorage.getItem('characters');
        if (storedCharacters) {
          const parsedCharacters = JSON.parse(storedCharacters);
          setCharacters(parsedCharacters);
          if (parsedCharacters.length > 0) {
            setSelectedCharacter(parsedCharacters[0]);
          }
        } else {
          // Mock data for development
          const mockCharacters = [
            { id: '1', name: 'Rook Knight', world: 'Antica', level: 120, vocation: 'Knight' },
            { id: '2', name: 'Magic User', world: 'Secura', level: 85, vocation: 'Sorcerer' }
          ];
          
          setCharacters(mockCharacters);
          if (mockCharacters.length > 0) {
            setSelectedCharacter(mockCharacters[0]);
          }
          
          // Save to localStorage for persistence
          localStorage.setItem('characters', JSON.stringify(mockCharacters));
        }
      } catch (err) {
        console.error('Failed to load characters:', err);
      }
    };

    fetchCharacters();
  }, []);

  // Function to add a character
  const addCharacter = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!newCharacterName) return;
    
    setIsLoading(true);
    setError(null);
    
    try {
      // In a real implementation, we would call the API
      // For now, we'll simulate it
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const newChar = {
        id: Date.now().toString(),
        name: newCharacterName,
        world: 'Antica',
        level: 100,
        vocation: 'Knight'
      };
      
      const updatedCharacters = [...characters, newChar];
      setCharacters(updatedCharacters);
      setSelectedCharacter(newChar);
      setNewCharacterName('');
      
      // Save to localStorage for persistence
      localStorage.setItem('characters', JSON.stringify(updatedCharacters));
    } catch (err: any) {
      setError(err.message || 'Failed to add character');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-amber-200 dark:border-amber-800">
      <h2 className="text-xl font-bold text-amber-800 dark:text-amber-500 mb-4">Your Characters</h2>

      {characters.length > 0 ? (
        <div className="mb-4">
          <div className="flex flex-wrap gap-2 mb-4">
            {characters.map(character => (
              <button
                key={character.id}
                onClick={() => setSelectedCharacter(character)}
                className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${
                  selectedCharacter?.id === character.id
                    ? 'bg-amber-500 text-white'
                    : 'bg-gray-200 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
                }`}
              >
                {character.name}
                <span className="ml-2 text-xs opacity-75">{character.world}</span>
              </button>
            ))}
          </div>
          
          {selectedCharacter && (
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Selected: {selectedCharacter.name} ({selectedCharacter.world})
            </p>
          )}
        </div>
      ) : (
        <div className="text-center py-4 text-gray-500 dark:text-gray-400">
          <p>You haven't added any characters yet.</p>
        </div>
      )}
      
      <hr className="my-4 border-gray-200 dark:border-gray-700" />
      
      {/* Add Character Form */}
      <form onSubmit={addCharacter} className="mb-4">
        <h3 className="text-lg font-semibold mb-2 text-amber-800 dark:text-amber-500">Add a Character</h3>
        
        <div className="mb-4">
          <label htmlFor="characterName" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Character Name
          </label>
          <input
            id="characterName"
            type="text"
            value={newCharacterName}
            onChange={(e) => setNewCharacterName(e.target.value)}
            placeholder="Enter your Tibia character name"
            disabled={isLoading}
            required
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
          />
        </div>
        
        <button
          type="submit"
          disabled={!newCharacterName || isLoading}
          className={`w-full px-4 py-2 rounded-md text-white font-medium ${
            !newCharacterName || isLoading
              ? 'bg-gray-400 cursor-not-allowed'
              : 'bg-amber-600 hover:bg-amber-700'
          }`}
        >
          {isLoading ? 'Adding...' : 'Add Character'}
        </button>
        
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
      </form>
    </div>
  );
} 