'use client';

import Link from 'next/link';
import { usePlayer } from '@/contexts/PlayerContext';
import { useState } from 'react';
import { useRouter } from 'next/navigation';

export default function Navbar() {
  const { player, isAnonymous, logout, refreshPlayer } = usePlayer();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const router = useRouter();

  const handleLogout = async () => {
    logout();
    // Refresh player context to ensure UI updates
    await refreshPlayer();
    router.push('/');
    // Close mobile menu if open
    setIsMenuOpen(false);
  };

  return (
    <header className="bg-white dark:bg-gray-800 shadow">
      <div className="container mx-auto px-4 py-4 flex justify-between items-center">
        <Link href="/" className="text-xl font-bold text-amber-600 dark:text-amber-400 hover:text-amber-700 dark:hover:text-amber-300 transition-colors">
          SoulPit Manager
        </Link>
        
        <nav className="flex items-center">
          <ul className="hidden md:flex space-x-4 items-center">
            {player ? (
              <>
                <li>
                  <Link 
                    href="/profile" 
                    className="flex items-center gap-2 px-4 py-2 rounded-full bg-amber-100 hover:bg-amber-200 dark:bg-amber-800 dark:hover:bg-amber-700 text-amber-800 dark:text-amber-200 transition-colors font-medium"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clipRule="evenodd" />
                    </svg>
                    <span>{player.username}</span>
                    {isAnonymous && (
                      <span className="ml-1 text-xs bg-amber-200 dark:bg-amber-700 px-1.5 py-0.5 rounded-full">
                        Anonymous
                      </span>
                    )}
                  </Link>
                </li>
                {!isAnonymous && (
                  <li>
                    <button 
                      onClick={handleLogout}
                      className="flex items-center gap-2 px-4 py-2 rounded-full bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 transition-colors font-medium"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M3 3a1 1 0 00-1 1v12a1 1 0 102 0V4a1 1 0 00-1-1zm10.293 9.293a1 1 0 001.414 1.414l3-3a1 1 0 000-1.414l-3-3a1 1 0 10-1.414 1.414L14.586 9H7a1 1 0 100 2h7.586l-1.293 1.293z" clipRule="evenodd" />
                      </svg>
                      <span>Logout</span>
                    </button>
                  </li>
                )}
              </>
            ) : (
              <>
                <li>
                  <Link 
                    href="/login" 
                    className="flex items-center gap-2 px-4 py-2 rounded-full bg-white hover:bg-gray-100 dark:bg-gray-700 dark:hover:bg-gray-600 text-amber-600 dark:text-amber-300 border border-amber-300 dark:border-amber-700 transition-colors font-medium"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M3 3a1 1 0 011 1v12a1 1 0 11-2 0V4a1 1 0 011-1zm7.707 3.293a1 1 0 010 1.414L9.414 9H17a1 1 0 110 2H9.414l1.293 1.293a1 1 0 01-1.414 1.414l-3-3a1 1 0 010-1.414l3-3a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                    <span>Log In</span>
                  </Link>
                </li>
                <li>
                  <Link 
                    href="/register" 
                    className="flex items-center gap-2 px-4 py-2 rounded-full bg-amber-100 hover:bg-amber-200 dark:bg-amber-800 dark:hover:bg-amber-700 text-amber-800 dark:text-amber-200 transition-colors font-medium"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                      <path d="M8 9a3 3 0 100-6 3 3 0 000 6zM8 11a6 6 0 016 6H2a6 6 0 016-6zM16 7a1 1 0 10-2 0v1h-1a1 1 0 100 2h1v1a1 1 0 102 0v-1h1a1 1 0 100-2h-1V7z" />
                    </svg>
                    <span>Register</span>
                  </Link>
                </li>
              </>
            )}
          </ul>
          
          {/* Mobile menu button */}
          <button 
            className="md:hidden text-gray-600 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white"
            onClick={() => setIsMenuOpen(!isMenuOpen)}
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
        </nav>
      </div>
      
      {/* Mobile menu */}
      {isMenuOpen && (
        <div className="md:hidden bg-white dark:bg-gray-800 shadow-md">
          <ul className="px-4 py-2">
            {player ? (
              <>
                <li className="py-2">
                  <Link 
                    href="/profile" 
                    className="flex items-center gap-2 px-4 py-2 rounded-md bg-amber-100 hover:bg-amber-200 dark:bg-amber-800 dark:hover:bg-amber-700 text-amber-800 dark:text-amber-200 transition-colors font-medium"
                    onClick={() => setIsMenuOpen(false)}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clipRule="evenodd" />
                    </svg>
                    <span>{player.username}</span>
                    {isAnonymous && (
                      <span className="ml-1 text-xs bg-amber-200 dark:bg-amber-700 px-1.5 py-0.5 rounded-full">
                        Anonymous
                      </span>
                    )}
                  </Link>
                </li>
                {!isAnonymous && (
                  <li className="py-2">
                    <button 
                      onClick={handleLogout}
                      className="flex w-full items-center gap-2 px-4 py-2 rounded-md bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 transition-colors font-medium"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M3 3a1 1 0 00-1 1v12a1 1 0 102 0V4a1 1 0 00-1-1zm10.293 9.293a1 1 0 001.414 1.414l3-3a1 1 0 000-1.414l-3-3a1 1 0 10-1.414 1.414L14.586 9H7a1 1 0 100 2h7.586l-1.293 1.293z" clipRule="evenodd" />
                      </svg>
                      <span>Logout</span>
                    </button>
                  </li>
                )}
              </>
            ) : (
              <>
                <li className="py-2">
                  <Link 
                    href="/login" 
                    className="flex items-center gap-2 px-4 py-2 rounded-md bg-white hover:bg-gray-100 dark:bg-gray-700 dark:hover:bg-gray-600 text-amber-600 dark:text-amber-300 border border-amber-300 dark:border-amber-700 transition-colors font-medium"
                    onClick={() => setIsMenuOpen(false)}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M3 3a1 1 0 011 1v12a1 1 0 11-2 0V4a1 1 0 011-1zm7.707 3.293a1 1 0 010 1.414L9.414 9H17a1 1 0 110 2H9.414l1.293 1.293a1 1 0 01-1.414 1.414l-3-3a1 1 0 010-1.414l3-3a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                    <span>Log In</span>
                  </Link>
                </li>
                <li className="py-2">
                  <Link 
                    href="/register" 
                    className="flex items-center gap-2 px-4 py-2 rounded-md bg-amber-100 hover:bg-amber-200 dark:bg-amber-800 dark:hover:bg-amber-700 text-amber-800 dark:text-amber-200 transition-colors font-medium"
                    onClick={() => setIsMenuOpen(false)}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                      <path d="M8 9a3 3 0 100-6 3 3 0 000 6zM8 11a6 6 0 016 6H2a6 6 0 016-6zM16 7a1 1 0 10-2 0v1h-1a1 1 0 100 2h1v1a1 1 0 102 0v-1h1a1 1 0 100-2h-1V7z" />
                    </svg>
                    <span>Register</span>
                  </Link>
                </li>
              </>
            )}
          </ul>
        </div>
      )}
    </header>
  );
} 