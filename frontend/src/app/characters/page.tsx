'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { usePlayer } from '@/contexts/PlayerContext';

export default function CharactersPage() {
  const router = useRouter();
  const { player, characters, loading, error, fetchCharacters, deleteCharacter } = usePlayer();
  const [isDeleting, setIsDeleting] = useState<string | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  useEffect(() => {
    if (player) {
      fetchCharacters();
    }
  }, [player, fetchCharacters]);

  // Handle character deletion
  const handleDelete = async (characterId: string, characterName: string) => {
    if (!confirm(`Are you sure you want to delete ${characterName}?`)) {
      return;
    }

    setIsDeleting(characterId);
    setDeleteError(null);

    try {
      await deleteCharacter(characterId);
    } catch (err: any) {
      console.error('Failed to delete character:', err);
      setDeleteError(`Failed to delete ${characterName}. Please try again.`);
    } finally {
      setIsDeleting(null);
    }
  };

  // Get vocation icon class
  const getVocationIcon = (vocation: string) => {
    const vocationLower = vocation.toLowerCase();
    if (vocationLower.includes('knight')) return 'fas fa-shield-alt text-red-600';
    if (vocationLower.includes('paladin')) return 'fas fa-bullseye text-green-600';
    if (vocationLower.includes('druid')) return 'fas fa-leaf text-teal-600';
    if (vocationLower.includes('sorcerer')) return 'fas fa-hat-wizard text-purple-600';
    return 'fas fa-user text-gray-600';
  };

  // Get level color class based on level range
  const getLevelColorClass = (level: number) => {
    if (level >= 400) return 'text-purple-600 font-bold';
    if (level >= 300) return 'text-amber-600 font-bold';
    if (level >= 200) return 'text-green-600 font-bold';
    if (level >= 100) return 'text-blue-600 font-bold';
    return 'text-gray-600';
  };

  return (
    <div className="max-w-4xl mx-auto">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-amber-200 dark:border-amber-800">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold text-amber-800 dark:text-amber-500">My Characters</h1>
          <Link
            href="/characters/create"
            className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
          >
            <svg className="-ml-1 mr-2 h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z" clipRule="evenodd" />
            </svg>
            Add Character
          </Link>
        </div>

        {!player && (
          <div className="mb-6 p-4 bg-amber-50 dark:bg-amber-900/30 rounded-md border border-amber-200 dark:border-amber-800">
            <p className="text-amber-800 dark:text-amber-400">
              You need to create an account before managing characters. Please go to the home page to set up your account.
            </p>
            <Link
              href="/"
              className="mt-2 inline-flex items-center px-4 py-2 rounded-md text-white font-medium bg-amber-600 hover:bg-amber-700"
            >
              Create Account
            </Link>
          </div>
        )}

        {(error || deleteError) && (
          <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/30 rounded-md border border-red-200 dark:border-red-800 text-sm text-red-700 dark:text-red-300">
            {error || deleteError}
          </div>
        )}

        {loading ? (
          <div className="flex justify-center items-center py-12">
            <svg className="animate-spin h-8 w-8 text-amber-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <span className="ml-2 text-amber-800 dark:text-amber-400">Loading characters...</span>
          </div>
        ) : characters && characters.length > 0 ? (
          <div className="overflow-hidden border border-amber-200 dark:border-amber-800 rounded-lg">
            <table className="min-w-full divide-y divide-amber-200 dark:divide-amber-800">
              <thead className="bg-amber-50 dark:bg-amber-900/30">
                <tr>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-amber-800 dark:text-amber-400 uppercase tracking-wider">
                    Character
                  </th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-amber-800 dark:text-amber-400 uppercase tracking-wider">
                    World
                  </th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-amber-800 dark:text-amber-400 uppercase tracking-wider">
                    Level
                  </th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-amber-800 dark:text-amber-400 uppercase tracking-wider">
                    Vocation
                  </th>
                  <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-amber-800 dark:text-amber-400 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-gray-800 divide-y divide-amber-100 dark:divide-amber-900/30">
                {characters.map((character) => (
                  <tr key={character.id} className="hover:bg-amber-50 dark:hover:bg-amber-900/10">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="font-medium text-amber-800 dark:text-amber-400">{character.name}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-700 dark:text-gray-300">{character.world}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className={`text-sm ${getLevelColorClass(character.level)}`}>{character.level}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-700 dark:text-gray-300 flex items-center">
                        <i className={`${getVocationIcon(character.vocation)} mr-2`}></i>
                        {character.vocation}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => handleDelete(character.id, character.name)}
                        disabled={isDeleting === character.id}
                        className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300 ml-2 focus:outline-none disabled:opacity-50"
                      >
                        {isDeleting === character.id ? (
                          <span className="flex items-center">
                            <svg className="animate-spin -ml-1 mr-2 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                            </svg>
                            Deleting...
                          </span>
                        ) : (
                          <span>Delete</span>
                        )}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : player ? (
          <div className="text-center py-12">
            <div className="text-amber-600 dark:text-amber-400 mb-4">
              <svg className="mx-auto h-12 w-12" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-amber-800 dark:text-amber-500 mb-2">No characters yet</h3>
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              Add your first Tibia character to start tracking your Soul Pit progress.
            </p>
            <Link
              href="/characters/create"
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
            >
              <svg className="-ml-1 mr-2 h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z" clipRule="evenodd" />
              </svg>
              Add Character
            </Link>
          </div>
        ) : null}
      </div>
    </div>
  );
} 