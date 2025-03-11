'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { usePlayer } from '@/contexts/PlayerContext';
import { soulpitListApi } from '@/services/api';

export default function ShareListPage({ params }: { params: { id: string } }) {
  const router = useRouter();
  const { player, loading: playerLoading } = usePlayer();
  
  // State
  const [listDetails, setListDetails] = useState<{ name: string; isOwner: boolean } | null>(null);
  const [shareCode, setShareCode] = useState<string>('');
  const [shareUrl, setShareUrl] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState<'code' | 'url' | null>(null);
  
  // Fetch list details and generate share code
  useEffect(() => {
    const fetchListDetails = async () => {
      if (!player) return;
      
      setIsLoading(true);
      setError(null);
      
      try {
        // In a real implementation, we would call the API
        // For now, we'll use mock data
        await new Promise(resolve => setTimeout(resolve, 500));
        
        // Mock data for list details
        const mockListDetails = {
          name: 'Soul Pit Hunters',
          isOwner: true
        };
        
        // Generate a mock share code
        const mockShareCode = `SP-${params.id}-${Math.random().toString(36).substring(2, 8).toUpperCase()}`;
        
        setListDetails(mockListDetails);
        setShareCode(mockShareCode);
        setShareUrl(`${window.location.origin}/lists/join?code=${mockShareCode}`);
      } catch (err: any) {
        console.error('Failed to fetch list details:', err);
        setError(err.message || 'Failed to fetch list details. Please try again.');
      } finally {
        setIsLoading(false);
      }
    };
    
    fetchListDetails();
  }, [params.id, player]);
  
  // Copy to clipboard
  const copyToClipboard = (text: string, type: 'code' | 'url') => {
    navigator.clipboard.writeText(text);
    setCopied(type);
    
    // Reset copied state after 2 seconds
    setTimeout(() => {
      setCopied(null);
    }, 2000);
  };
  
  // Generate new share code
  const generateNewCode = async () => {
    if (!player || !listDetails) return;
    
    setIsLoading(true);
    setError(null);
    
    try {
      // In a real implementation, we would call the API
      // For now, we'll just generate a new mock code
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // Generate a new mock share code
      const newShareCode = `SP-${params.id}-${Math.random().toString(36).substring(2, 8).toUpperCase()}`;
      
      setShareCode(newShareCode);
      setShareUrl(`${window.location.origin}/lists/join?code=${newShareCode}`);
    } catch (err: any) {
      console.error('Failed to generate new share code:', err);
      setError(err.message || 'Failed to generate new share code. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };
  
  // Redirect to login if not logged in
  if (!playerLoading && !player) {
    router.push('/');
    return null;
  }
  
  if (playerLoading || isLoading) {
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

  if (error) {
    return (
      <div className="max-w-2xl mx-auto">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-amber-200 dark:border-amber-800">
          <div className="text-center py-8">
            <div className="text-red-600 dark:text-red-400 mb-4">
              <svg className="mx-auto h-12 w-12" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-red-800 dark:text-red-500 mb-2">Error Loading List</h3>
            <p className="text-gray-600 dark:text-gray-400 mb-4">{error}</p>
            <Link
              href={`/lists/${params.id}`}
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
            >
              Go Back to List
            </Link>
          </div>
        </div>
      </div>
    );
  }

  if (!listDetails) {
    return (
      <div className="max-w-2xl mx-auto">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-amber-200 dark:border-amber-800">
          <div className="text-center py-8">
            <div className="text-amber-600 dark:text-amber-400 mb-4">
              <svg className="mx-auto h-12 w-12" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-amber-800 dark:text-amber-500 mb-2">List Not Found</h3>
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              The list you're looking for doesn't exist or you don't have access to it.
            </p>
            <Link
              href="/"
              className="text-amber-600 hover:text-amber-800 dark:text-amber-400 dark:hover:text-amber-300"
            >
              Back to Lists
            </Link>
          </div>
        </div>
      </div>
    );
  }

  if (!listDetails.isOwner) {
    return (
      <div className="max-w-2xl mx-auto">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-amber-200 dark:border-amber-800">
          <div className="text-center py-8">
            <div className="text-amber-600 dark:text-amber-400 mb-4">
              <svg className="mx-auto h-12 w-12" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-amber-800 dark:text-amber-500 mb-2">Access Denied</h3>
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              Only the list owner can access the sharing options.
            </p>
            <Link
              href={`/lists/${params.id}`}
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
            >
              Go Back to List
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-amber-200 dark:border-amber-800">
        <div className="flex items-center mb-6">
          <Link
            href={`/lists/${params.id}`}
            className="mr-2 text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
          >
            <svg className="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
          </Link>
          <h1 className="text-2xl font-bold text-amber-800 dark:text-amber-500">Share {listDetails.name}</h1>
        </div>
        
        <div className="mb-8">
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            Share your Soul Pit list with friends and guildmates. They can use either the share code or URL to join your list.
          </p>
        </div>
        
        <div className="space-y-6">
          {/* Share Code */}
          <div className="bg-amber-50 dark:bg-amber-900/30 rounded-lg p-4 border border-amber-200 dark:border-amber-800">
            <h2 className="text-lg font-semibold text-amber-800 dark:text-amber-500 mb-2">Share Code</h2>
            <p className="text-gray-600 dark:text-gray-400 text-sm mb-3">
              Share this code with others. They can enter it on the "Join List" page.
            </p>
            
            <div className="flex">
              <input
                type="text"
                value={shareCode}
                readOnly
                className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-l-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
              />
              <button
                onClick={() => copyToClipboard(shareCode, 'code')}
                className="px-4 py-2 border border-transparent rounded-r-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
              >
                {copied === 'code' ? 'Copied!' : 'Copy'}
              </button>
            </div>
          </div>
          
          {/* Share URL */}
          <div className="bg-amber-50 dark:bg-amber-900/30 rounded-lg p-4 border border-amber-200 dark:border-amber-800">
            <h2 className="text-lg font-semibold text-amber-800 dark:text-amber-500 mb-2">Share URL</h2>
            <p className="text-gray-600 dark:text-gray-400 text-sm mb-3">
              Share this URL with others. They can click it to join your list directly.
            </p>
            
            <div className="flex">
              <input
                type="text"
                value={shareUrl}
                readOnly
                className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-l-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
              />
              <button
                onClick={() => copyToClipboard(shareUrl, 'url')}
                className="px-4 py-2 border border-transparent rounded-r-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
              >
                {copied === 'url' ? 'Copied!' : 'Copy'}
              </button>
            </div>
          </div>
          
          {/* Generate New Code */}
          <div className="bg-amber-50 dark:bg-amber-900/30 rounded-lg p-4 border border-amber-200 dark:border-amber-800">
            <h2 className="text-lg font-semibold text-amber-800 dark:text-amber-500 mb-2">Generate New Code</h2>
            <p className="text-gray-600 dark:text-gray-400 text-sm mb-3">
              If you want to revoke access to the current share code, you can generate a new one. The old code will no longer work.
            </p>
            
            <button
              onClick={generateNewCode}
              className="w-full px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
            >
              Generate New Code
            </button>
          </div>
        </div>
        
        <div className="mt-8 pt-6 border-t border-gray-200 dark:border-gray-700">
          <div className="flex justify-end">
            <Link
              href={`/lists/${params.id}`}
              className="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
            >
              Back to List
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
} 