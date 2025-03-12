'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { usePlayer } from '@/contexts/PlayerContext';
import listApi from '@/services/listApi';
import { use } from 'react';

// Define interfaces
interface SoulCore {
  id: string;
  creatureName: string;
  obtained: boolean;
  unlocked: boolean;
  obtainedBy?: string;
}

interface ListMember {
  id: string;
  username: string;
  characterName: string;
  world: string;
  isOwner: boolean;
}

interface ListDetails {
  id: string;
  name: string;
  description?: string;
  isOwner: boolean;
  createdAt: string;
  updatedAt: string;
  members: ListMember[];
  soulCores: SoulCore[];
  share_code: string;
}

interface Creature {
  endpoint: string;
  name: string;
  plural_name: string;
}

export default function ListDetailPage({ params }: { params: any }) {
  const router = useRouter();
  const { player, loading: playerLoading, fetchCharacters, fetchAnonymousPlayer: fetchAnonymousPlayerFromContext } = usePlayer();
  
  // Properly unwrap the params object using React.use()
  const unwrappedParams = use(params) as { id: string };
  const listId = unwrappedParams.id;
  
  // State
  const [listDetails, setListDetails] = useState<ListDetails | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'cores' | 'members'>('cores');
  const [searchTerm, setSearchTerm] = useState('');
  const [creatures, setCreatures] = useState<Creature[]>([]);
  const [selectedCreature, setSelectedCreature] = useState<Creature | null>(null);
  const [isAddingCore, setIsAddingCore] = useState(false);
  const [showCreatureDropdown, setShowCreatureDropdown] = useState(false);
  const [creatureSearchTerm, setCreatureSearchTerm] = useState('');
  const [hasFetchedList, setHasFetchedList] = useState(false);
  const [showShareModal, setShowShareModal] = useState(false);
  const [copySuccess, setCopySuccess] = useState(false);
  
  // Debug information
  const [showDebugInfo, setShowDebugInfo] = useState(false);
  const [localStorageData, setLocalStorageData] = useState<{[key: string]: string | null}>({});
  const [contextInfo, setContextInfo] = useState<any>(null);
  
  const [sortField, setSortField] = useState<'name' | 'status'>('name');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');
  
  useEffect(() => {
    // Get localStorage data
    const playerData = localStorage.getItem('player');
    const tempSessionId = localStorage.getItem('tempSessionId');
    
    setLocalStorageData({
      player: playerData,
      tempSessionId: tempSessionId
    });
    
    // Get context info
    setContextInfo({
      playerExists: !!player,
      playerLoading,
      playerIsAnonymous: player?.is_anonymous || false,
      playerHasCharacters: player?.characters && player.characters.length > 0 || false,
      playerHasId: !!player?.id,
      playerHasUsername: !!player?.username
    });
  }, [player, playerLoading]);
  
  const toggleDebugInfo = () => {
    setShowDebugInfo(!showDebugInfo);
  };
  
  // Function to manually load player data from localStorage
  const loadPlayerFromLocalStorage = () => {
    try {
      // Try to get player data from localStorage
      const storedPlayer = localStorage.getItem('player');
      if (storedPlayer) {
        const parsedPlayer = JSON.parse(storedPlayer);
        console.log('Manually loaded player from localStorage:', parsedPlayer);
        
        // Update localStorage data state
        setLocalStorageData(prev => ({
          ...prev,
          player: storedPlayer
        }));
        
        // Refresh the page to reload the player context
        window.location.reload();
      } else {
        console.log('No player data found in localStorage');
        alert('No player data found in localStorage');
      }
    } catch (err) {
      console.error('Error loading player from localStorage:', err);
      alert(`Error loading player from localStorage: ${err}`);
    }
  };
  
  // Function to manually create a test player in localStorage
  const createTestPlayerInLocalStorage = () => {
    try {
      const testPlayer = {
        id: "test_player_id",
        username: "Test User",
        characters: [],
        is_anonymous: false,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      };
      
      localStorage.setItem('player', JSON.stringify(testPlayer));
      console.log('Created test player in localStorage:', testPlayer);
      
      // Update localStorage data state
      setLocalStorageData(prev => ({
        ...prev,
        player: JSON.stringify(testPlayer)
      }));
      
      // Refresh the page to reload the player context
      window.location.reload();
    } catch (err) {
      console.error('Error creating test player in localStorage:', err);
      alert(`Error creating test player in localStorage: ${err}`);
    }
  };
  
  // Function to clear player data from localStorage
  const clearPlayerFromLocalStorage = () => {
    try {
      localStorage.removeItem('player');
      console.log('Cleared player data from localStorage');
      
      // Update localStorage data state
      setLocalStorageData(prev => ({
        ...prev,
        player: null
      }));
      
      // Refresh the page to reload the player context
      window.location.reload();
    } catch (err) {
      console.error('Error clearing player from localStorage:', err);
      alert(`Error clearing player from localStorage: ${err}`);
    }
  };
  
  // Function to manually fetch anonymous player data from the API
  const fetchAnonymousPlayerData = async () => {
    try {
      const tempSessionId = localStorage.getItem('tempSessionId');
      if (!tempSessionId) {
        console.log('No tempSessionId found in localStorage');
        alert('No tempSessionId found in localStorage. Cannot fetch anonymous player.');
        return;
      }
      
      console.log('Fetching anonymous player with session ID:', tempSessionId);
      
      // Call the fetchAnonymousPlayer function from the PlayerContext
      await fetchAnonymousPlayerFromContext(tempSessionId);
      
      // Update localStorage data
      setLocalStorageData(prev => ({
        ...prev,
        tempSessionId
      }));
      
      alert('Attempted to fetch anonymous player. Check console for details.');
    } catch (err) {
      console.error('Error fetching anonymous player:', err);
      // Simply log the error and show a message
      alert('Failed to fetch anonymous player. Session may be invalid.');
    }
  };
  
  // Function to create a temporary session ID
  const createTempSessionId = () => {
    try {
      // Generate a random session ID
      const tempSessionId = Math.random().toString(36).substring(2, 15);
      
      // Store it in localStorage
      localStorage.setItem('tempSessionId', tempSessionId);
      console.log('Created temporary session ID:', tempSessionId);
      
      // Update localStorage data
      setLocalStorageData(prev => ({
        ...prev,
        tempSessionId
      }));
      
      alert(`Created temporary session ID: ${tempSessionId}`);
    } catch (err) {
      console.error('Error creating temporary session ID:', err);
      alert(`Error creating temporary session ID: ${err}`);
    }
  };
  
  // Function to manually set the player ID
  const setPlayerIdManually = () => {
    try {
      // Get the player ID from the list members
      if (!listDetails || !listDetails.members || listDetails.members.length === 0) {
        alert('No list details or members available');
        return;
      }
      
      // Use the first member's ID as the player ID
      const playerId = listDetails.members[0].id;
      if (!playerId) {
        alert('No player ID found in list members');
        return;
      }
      
      // Create a temporary player object with the ID
      const tempPlayer = {
        id: playerId,
        username: listDetails.members[0].username || "User",
        characters: [],
        is_anonymous: false,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      };
      
      // Store it in localStorage for the PlayerContext to pick up
      localStorage.setItem('player', JSON.stringify(tempPlayer));
      console.log('Stored player in localStorage with ID:', playerId, tempPlayer);
      
      // Update localStorage data state
      setLocalStorageData(prev => ({
        ...prev,
        player: JSON.stringify(tempPlayer)
      }));
      
      // Add this ID to the URL for future reference
      const url = new URL(window.location.href);
      url.searchParams.set('player_id', playerId);
      window.history.replaceState({}, '', url.toString());
      
      alert(`Set player ID to: ${playerId}\nRefresh the page to apply changes.`);
    } catch (err) {
      console.error('Error setting player ID:', err);
      alert(`Error setting player ID: ${err}`);
    }
  };
  
  // Calculate summary statistics
  const calculateSummary = useCallback(() => {
    if (!listDetails) return null;
    
    const totalCores = listDetails.soulCores.length;
    const obtainedCores = listDetails.soulCores.filter(core => core.obtained).length;
    const unlockedCores = listDetails.soulCores.filter(core => core.unlocked).length;
    
    // Calculate which members obtained cores
    const memberContributions: Record<string, { obtained: number }> = {};
    
    listDetails.members.forEach(member => {
      memberContributions[member.characterName] = { obtained: 0 };
    });
    
    listDetails.soulCores.forEach(core => {
      if (core.obtained && core.obtainedBy) {
        // Find the member who obtained this core
        const memberName = core.obtainedBy;
        if (memberContributions[memberName]) {
          memberContributions[memberName].obtained += 1;
        } else {
          memberContributions[memberName] = { obtained: 1 };
        }
      }
    });
    
    // Constants for the game
    const maxXpBoostCores = 200; // Cores needed for max XP boost
    const totalCoresInGame = creatures.length; // Total cores available in the game
    
    return {
      totalCores,
      obtainedCores,
      unlockedCores,
      memberContributions,
      maxXpBoostCores,
      totalCoresInGame
    };
  }, [listDetails, creatures]);

  const summary = calculateSummary();
  
  // Update list details when player context changes
  useEffect(() => {
    if (listDetails && player) {
      // Update isOwner flag based on player context
      const isOwner = listDetails.members.some(member => 
        member.id === player.id && member.isOwner
      );
      
      if (isOwner !== listDetails.isOwner) {
        setListDetails({
          ...listDetails,
          isOwner
        });
      }
      
      console.log('Player context updated:', player.id, 'isOwner:', isOwner);
    }
  }, [player, listDetails]);
  
  // Handle direct URL access by checking for player ID in URL or query parameters
  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search);
    const playerIdFromUrl = searchParams.get('player_id');
    
    // If we have a player ID in the URL and no player in context, try to load it
    if (playerIdFromUrl && !player && !playerLoading) {
      console.log('Found player ID in URL:', playerIdFromUrl);
      
      // Check if we already have player data in localStorage
      const storedPlayer = localStorage.getItem('player');
      if (storedPlayer) {
        try {
          const parsedPlayer = JSON.parse(storedPlayer);
          console.log('Found player data in localStorage:', parsedPlayer);
          
          // If the player ID in localStorage doesn't match the one in the URL,
          // log a warning but don't override it
          if (parsedPlayer.id !== playerIdFromUrl) {
            console.warn('Player ID in localStorage does not match URL parameter:', 
              parsedPlayer.id, 'vs', playerIdFromUrl);
          }
          
          // No need to do anything else, the PlayerContext will load this data
        } catch (err) {
          console.error('Failed to parse stored player:', err);
        }
      } else {
        console.log('No player data found in localStorage, but player_id is in URL');
        // We don't want to automatically create player data from the URL parameter
        // as this should come from the authentication flow
      }
    }
  }, [player, playerLoading]);
  
  // Fetch list details
  useEffect(() => {
    const fetchListDetails = async () => {
      setIsLoading(true);
      setError(null);
      
      try {
        if (!listId) {
          setError('List ID is missing');
          setIsLoading(false);
          setHasFetchedList(true);
          return;
        }
        
        const list = await listApi.getList(listId) as any;
        
        if (!list || !list.id) {
          setError('Failed to load list details');
          setIsLoading(false);
          setHasFetchedList(true);
          return;
        }
        
        // Log player and list information for debugging
        if (player) {
          console.log('Player context is available:', player.id);
          if (Array.isArray(list.members)) {
            const isPlayerInList = list.members.some((member: any) => member.player_id === player.id);
            console.log('Player in list check:', isPlayerInList, 'Player ID:', player.id);
          }
        } else {
          console.log('No player context available');
        }
        
        // Transform the data to match our interface
        const transformedList: ListDetails = {
          id: list.id,
          name: list.name || 'Unnamed List',
          description: list.description || '',
          isOwner: !!list.is_creator,
          createdAt: list.created_at || new Date().toISOString(),
          updatedAt: list.updated_at || new Date().toISOString(),
          members: Array.isArray(list.members) ? list.members.map((member: any) => ({
            id: member.player_id || '',
            username: member.username || member.character_name || 'Unknown',
            characterName: member.character_name || 'Unknown',
            world: member.world || list.world || 'Unknown',
            isOwner: !!member.is_creator
          })) : [],
          soulCores: Array.isArray(list.soul_cores) ? list.soul_cores.map((core: any) => ({
            id: core.id || '',
            creatureName: core.creature_name || (core.creature && core.creature.name) || 'Unknown Creature',
            obtained: !!core.obtained,
            unlocked: !!core.unlocked,
            obtainedBy: core.obtained_by_name || 'Unknown',
          })) : [],
          share_code: list.share_code || ''
        };
        
        setListDetails(transformedList);
        setHasFetchedList(true);
      } catch (err: any) {
        console.error('Failed to fetch list details:', err);
        setError(err.message || 'Failed to fetch list details. Please try again.');
        setHasFetchedList(true);
      } finally {
        setIsLoading(false);
      }
    };
    
    fetchListDetails();
  }, [listId]);
  
  // Fetch creatures
  useEffect(() => {
    const fetchCreatures = async () => {
      try {
        const creaturesData = await listApi.getCreatures() as Creature[];
        setCreatures(creaturesData);
      } catch (err: any) {
        console.error('Failed to fetch creatures:', err);
      }
    };
    
    fetchCreatures();
  }, []);
  
  // Handle marking a soul core as obtained
  const handleMarkObtained = async (soulCoreId: string) => {
    if (!listDetails || !soulCoreId) {
      setError('Cannot update soul core: Missing required information');
      return;
    }
    
    if (!player) {
      setError('You must be logged in to mark a soul core as obtained');
      return;
    }
    
    try {
      // Update the soul core in the backend
      await listApi.updateSoulCore(listDetails.id, soulCoreId, {
        obtained: true
      });
      
      // Refresh the list details
      const updatedList = await listApi.getList(listId) as any;
      
      if (!updatedList || !updatedList.soul_cores) {
        setError('Failed to refresh list after updating soul core');
        return;
      }
      
      // Update the list details state
      setListDetails({
        ...listDetails,
        soulCores: updatedList.soul_cores.map((core: any) => ({
          id: core.id,
          creatureName: core.creature?.name || 'Unknown',
          obtained: core.obtained || false,
          unlocked: core.unlocked || false,
          obtainedBy: core.obtained_by_name || 'Unknown',
        }))
      });
    } catch (err: any) {
      console.error('Failed to mark soul core as obtained:', err);
      setError(err.message || 'Failed to mark soul core as obtained');
    }
  };
  
  // Handle marking a soul core as unlocked
  const handleMarkUnlocked = async (soulCoreId: string) => {
    if (!listDetails || !soulCoreId) {
      setError('Cannot update soul core: Missing required information');
      return;
    }
    
    if (!player) {
      setError('You must be logged in to mark a soul core as unlocked');
      return;
    }
    
    try {
      // Update the soul core in the backend
      await listApi.updateSoulCore(listDetails.id, soulCoreId, {
        unlocked: true
      });
      
      // Refresh the list details
      const updatedList = await listApi.getList(listId) as any;
      
      if (!updatedList || !updatedList.soul_cores) {
        setError('Failed to refresh list after updating soul core');
        return;
      }
      
      // Update the list details state
      setListDetails({
        ...listDetails,
        soulCores: updatedList.soul_cores.map((core: any) => ({
          id: core.id,
          creatureName: core.creature?.name || 'Unknown',
          obtained: core.obtained || false,
          unlocked: core.unlocked || false,
          obtainedBy: core.obtained_by_name || 'Unknown',
        }))
      });
    } catch (err: any) {
      console.error('Failed to mark soul core as unlocked:', err);
      setError(err.message || 'Failed to mark soul core as unlocked');
    }
  };
  
  // Handle adding a soul core
  const handleAddSoulCore = async () => {
    if (!listDetails || !selectedCreature) {
      setError('Cannot add soul core: Missing required information');
      return;
    }
    
    if (!player) {
      setError('You must be logged in to add a soul core');
      return;
    }
    
    setIsAddingCore(true);
    setError(null);
    
    try {
      // Find the current user's member entry
      const currentMember = listDetails.members.find(member => 
        member.id === player.id
      );
      
      if (!currentMember) {
        setError('You are not a member of this list');
        setIsAddingCore(false);
        return;
      }
      
      // Add the soul core
      await listApi.addSoulCore(listDetails.id, {
        creature_id: selectedCreature.endpoint,
        player_id: player.id
      });
      
      // Refresh the list details
      const updatedList = await listApi.getList(listId) as any;
      
      if (!updatedList || !updatedList.soul_cores) {
        setError('Failed to refresh list after adding soul core');
        setIsAddingCore(false);
        return;
      }
      
      // Transform and update the local state
      setListDetails(prev => {
        if (!prev) return prev;
        
        return {
          ...prev,
          soulCores: Array.isArray(updatedList.soul_cores) ? updatedList.soul_cores.map((core: any) => ({
            id: core.id || '',
            creatureName: core.creature && core.creature.name ? core.creature.name : (core.creature_name || 'Unknown Creature'),
            obtained: !!core.obtained,
            unlocked: !!core.unlocked,
            obtainedBy: core.obtained_by_name || 'Unknown',
          })) : prev.soulCores
        };
      });
      
      // Reset the selected creature
      setSelectedCreature(null);
      setCreatureSearchTerm('');
    } catch (err: any) {
      console.error('Failed to add soul core:', err);
      setError(err.message || 'Failed to add soul core. Please try again.');
    } finally {
      setIsAddingCore(false);
    }
  };
  
  // Filter creatures based on search term and exclude those already in the list
  const filteredCreatures = creatures.filter(creature => {
    // Check if the creature matches the search term
    const matchesSearch = (creature.name?.toLowerCase() || '').includes((creatureSearchTerm?.toLowerCase() || '')) ||
      (creature.plural_name?.toLowerCase() || '').includes((creatureSearchTerm?.toLowerCase() || ''));
    
    // Check if the creature is already in the list
    const isAlreadyInList = listDetails?.soulCores.some(core => 
      core.creatureName.toLowerCase() === creature.name.toLowerCase()
    );
    
    // Only include creatures that match the search and are not already in the list
    return matchesSearch && !isAlreadyInList;
  });
  
  // Function to format dates
  const formatDate = (dateString: string) => {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleDateString();
  };
  
  // Sort soul cores based on sort field and direction
  const sortSoulCores = (cores: SoulCore[]) => {
    return [...cores].sort((a, b) => {
      if (sortField === 'name') {
        // Sort by creature name
        return sortDirection === 'asc' 
          ? a.creatureName.localeCompare(b.creatureName)
          : b.creatureName.localeCompare(a.creatureName);
      } else {
        // Sort by status (obtained/unlocked/missing)
        const getStatusPriority = (core: SoulCore) => {
          if (core.unlocked) return 1;
          if (core.obtained) return 2;
          return 3; // missing
        };
        
        const priorityA = getStatusPriority(a);
        const priorityB = getStatusPriority(b);
        
        if (priorityA === priorityB) {
          // If status is the same, sort by name
          return a.creatureName.localeCompare(b.creatureName);
        }
        
        return sortDirection === 'asc'
          ? priorityA - priorityB
          : priorityB - priorityA;
      }
    });
  };
  
  // Handle sort change
  const handleSort = (field: 'name' | 'status') => {
    if (sortField === field) {
      // Toggle direction if clicking the same field
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      // Set new field and reset direction to ascending
      setSortField(field);
      setSortDirection('asc');
    }
  };
  
  // Filter soul cores based on search term
  const filteredSoulCores = sortSoulCores(
    listDetails?.soulCores?.filter(core => 
      (core.creatureName?.toLowerCase() || '').includes((searchTerm?.toLowerCase() || ''))
    ) || []
  );
  
  // Filter members based on search term
  const filteredMembers = listDetails?.members?.filter(member => 
    (member.username?.toLowerCase() || '').includes((searchTerm?.toLowerCase() || '')) ||
    (member.characterName?.toLowerCase() || '').includes((searchTerm?.toLowerCase() || ''))
  ) || [];
  
  // Handle share button click
  const handleShare = () => {
    setShowShareModal(true);
  };

  // Handle copy link
  const handleCopyLink = () => {
    const shareUrl = `${window.location.origin}/lists/join?code=${listDetails?.share_code}`;
    navigator.clipboard.writeText(shareUrl)
      .then(() => {
        setCopySuccess(true);
        setTimeout(() => setCopySuccess(false), 2000);
      })
      .catch(err => {
        console.error('Failed to copy link:', err);
      });
  };

  // Handle copy code
  const handleCopyCode = () => {
    navigator.clipboard.writeText(listDetails?.share_code || '')
      .then(() => {
        setCopySuccess(true);
        setTimeout(() => setCopySuccess(false), 2000);
      })
      .catch(err => {
        console.error('Failed to copy code:', err);
      });
  };

  // Handle close share modal
  const handleCloseShareModal = () => {
    setShowShareModal(false);
    setCopySuccess(false);
  };
  
  // Loading state
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
      <div className="max-w-4xl mx-auto">
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
              href="/"
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
            >
              Back to Lists
            </Link>
          </div>
        </div>
      </div>
    );
  }

  if (!listDetails) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-amber-200 dark:border-amber-800">
          <div className="text-center py-8">
            <div className="text-amber-600 dark:text-amber-400 mb-4">
              <svg className="mx-auto h-12 w-12" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-amber-800 dark:text-amber-500 mb-2">List Not Found</h3>
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              The list you&apos;re looking for doesn&apos;t exist or you don&apos;t have access to it.
            </p>
            <Link
              href="/"
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
            >
              Back to Lists
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto">
      {/* Debug Info Box */}
      <div className="mb-4">
        <button 
          onClick={toggleDebugInfo}
          className="px-3 py-1 text-xs bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-md hover:bg-gray-300 dark:hover:bg-gray-600"
        >
          {showDebugInfo ? 'Hide Debug Info' : 'Show Debug Info'}
        </button>
        
        {showDebugInfo && (
          <div className="mt-2 p-4 bg-gray-100 dark:bg-gray-800 rounded-md border border-gray-300 dark:border-gray-700 text-xs font-mono overflow-auto">
            <h3 className="font-bold mb-2">Debug Information</h3>
            
            <div className="mb-2">
              <h4 className="font-bold">Player Context:</h4>
              <div>Loading: {playerLoading ? 'true' : 'false'}</div>
              <div>Player: {player ? 'exists' : 'null'}</div>
              {player && (
                <>
                  <div>Player ID: {player.id}</div>
                  <div>Username: {player.username}</div>
                  <div>Is Anonymous: {player.is_anonymous ? 'true' : 'false'}</div>
                  <div>Characters: {player.characters?.length || 0}</div>
                </>
              )}
              
              <div className="mt-2">
                <h5 className="font-bold">Context Details:</h5>
                {contextInfo && (
                  <pre className="mt-1 p-2 bg-gray-200 dark:bg-gray-900 rounded overflow-auto max-h-40">
                    {JSON.stringify(contextInfo, null, 2)}
                  </pre>
                )}
              </div>
              
              {!player && !playerLoading && (
                <div className="mt-2 p-2 bg-red-50 dark:bg-red-900/20 rounded-md border border-red-200 dark:border-red-800">
                  <p className="text-sm text-red-700 dark:text-red-400 mb-2">
                    <strong>Authentication Issue:</strong> No player data found in context.
                  </p>
                  <p className="text-xs text-red-600 dark:text-red-300 mb-2">
                    You need to log in or create an account to access all features.
                    The player data should be stored in localStorage during the authentication flow.
                  </p>
                  <div className="flex space-x-2">
                    <Link
                      href="/lists/join"
                      className="px-2 py-1 bg-blue-500 text-white rounded-md hover:bg-blue-600 text-xs"
                    >
                      Go to Login Page
                    </Link>
                    <Link
                      href="/"
                      className="px-2 py-1 bg-green-500 text-white rounded-md hover:bg-green-600 text-xs"
                    >
                      Go to Home Page
                    </Link>
                  </div>
                </div>
              )}
            </div>
            
            <div className="mb-2">
              <h4 className="font-bold">localStorage:</h4>
              <div>player: {localStorageData.player ? 'exists' : 'null'}</div>
              {localStorageData.player && (
                <pre className="mt-1 p-2 bg-gray-200 dark:bg-gray-900 rounded overflow-auto max-h-40">
                  {JSON.stringify(JSON.parse(localStorageData.player), null, 2)}
                </pre>
              )}
              <div>tempSessionId: {localStorageData.tempSessionId || 'null'}</div>
              
              <div className="mt-2">
                <h5 className="font-bold">Authentication Flow:</h5>
                <ol className="list-decimal list-inside text-xs text-gray-600 dark:text-gray-400 mt-1 space-y-1 pl-2">
                  <li>User logs in or creates an account</li>
                  <li>Player data is stored in localStorage</li>
                  <li>PlayerContext loads player data from localStorage</li>
                  <li>Player context is available throughout the app</li>
                </ol>
              </div>
              
              <div className="mt-2 flex space-x-2">
                <button 
                  onClick={loadPlayerFromLocalStorage}
                  className="px-2 py-1 bg-blue-500 text-white rounded-md hover:bg-blue-600 text-xs"
                >
                  Reload Player
                </button>
                <button 
                  onClick={createTestPlayerInLocalStorage}
                  className="px-2 py-1 bg-green-500 text-white rounded-md hover:bg-green-600 text-xs"
                >
                  Create Test Player
                </button>
                <button 
                  onClick={clearPlayerFromLocalStorage}
                  className="px-2 py-1 bg-red-500 text-white rounded-md hover:bg-red-600 text-xs"
                >
                  Clear Player
                </button>
              </div>
            </div>
            
            <div>
              <h4 className="font-bold">List Details:</h4>
              <div>List ID: {listId}</div>
              <div>Is Owner: {listDetails?.isOwner ? 'true' : 'false'}</div>
              <div>Members: {listDetails?.members.length || 0}</div>
              <div>Soul Cores: {listDetails?.soulCores.length || 0}</div>
            </div>
          </div>
        )}
      </div>
      
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-amber-200 dark:border-amber-800">
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center">
            <Link
              href="/"
              className="mr-2 text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
            >
              <svg className="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
            </Link>
            <h1 className="text-2xl font-bold text-amber-800 dark:text-amber-500">{listDetails.name}</h1>
            {listDetails.isOwner && (
              <span className="ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300">
                Owner
              </span>
            )}
          </div>
          <button
            onClick={handleShare}
            className="inline-flex items-center px-3 py-1.5 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
          >
            <svg className="h-4 w-4 mr-1.5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
            </svg>
            Share
          </button>
        </div>
        
        {listDetails.description && (
          <p className="text-gray-600 dark:text-gray-400 mb-4">{listDetails.description}</p>
        )}
        
        <div className="flex flex-wrap text-sm text-gray-500 dark:text-gray-500 mb-6">
          <div className="mr-4">Created: {formatDate(listDetails.createdAt)}</div>
          <div>Updated: {formatDate(listDetails.updatedAt)}</div>
        </div>
        
        {/* Summary Box */}
        {summary && (
          <div className="mb-6 p-4 bg-amber-50 dark:bg-amber-900/30 rounded-md border border-amber-200 dark:border-amber-800">
            <h3 className="text-lg font-medium text-amber-800 dark:text-amber-500 mb-2">Summary</h3>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <div className="bg-white dark:bg-gray-700 p-3 rounded-md shadow-sm">
                <div className="text-2xl font-bold text-yellow-600 dark:text-yellow-400">
                  {summary.obtainedCores} / {summary.maxXpBoostCores}
                  <span className="text-sm ml-1 text-gray-500 dark:text-gray-400">
                    ({Math.round((summary.obtainedCores / summary.maxXpBoostCores) * 100)}%)
                  </span>
                </div>
                <div className="text-sm text-gray-500 dark:text-gray-400">Obtained (Max XP Boost)</div>
              </div>
              
              <div className="bg-white dark:bg-gray-700 p-3 rounded-md shadow-sm">
                <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                  {summary.obtainedCores} / {summary.totalCoresInGame}
                  <span className="text-sm ml-1 text-gray-500 dark:text-gray-400">
                    ({Math.round((summary.obtainedCores / summary.totalCoresInGame) * 100)}%)
                  </span>
                </div>
                <div className="text-sm text-gray-500 dark:text-gray-400">Obtained (All Creatures)</div>
              </div>
            </div>
            
            {Object.keys(summary.memberContributions).length > 0 && (
              <div>
                <h4 className="text-sm font-medium text-amber-700 dark:text-amber-400 mb-2">Member Contributions</h4>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                  {Object.entries(summary.memberContributions)
                    .filter(([_, stats]) => stats.obtained > 0)
                    .sort((a, b) => b[1].obtained - a[1].obtained)
                    .map(([memberName, stats]) => (
                      <div key={memberName} className="flex justify-between items-center bg-white dark:bg-gray-700 p-2 rounded-md text-sm">
                        <div className="font-medium text-gray-700 dark:text-gray-300">{memberName}</div>
                        <div className="flex space-x-2">
                          {stats.obtained > 0 && (
                            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300">
                              {stats.obtained} obtained
                            </span>
                          )}
                        </div>
                      </div>
                    ))}
                </div>
              </div>
            )}
          </div>
        )}
        
        <div className="mb-6">
          <div className="flex border-b border-amber-200 dark:border-amber-800">
            <button
              className={`px-4 py-2 font-medium text-sm ${
                activeTab === 'cores'
                  ? 'border-b-2 border-amber-500 text-amber-600 dark:text-amber-400'
                  : 'text-gray-500 hover:text-amber-500 dark:text-gray-400 dark:hover:text-amber-400'
              }`}
              onClick={() => setActiveTab('cores')}
            >
              Soul Cores
            </button>
            <button
              className={`px-4 py-2 font-medium text-sm ${
                activeTab === 'members'
                  ? 'border-b-2 border-amber-500 text-amber-600 dark:text-amber-400'
                  : 'text-gray-500 hover:text-amber-500 dark:text-gray-400 dark:hover:text-amber-400'
              }`}
              onClick={() => setActiveTab('members')}
            >
              Members
            </button>
          </div>
        </div>
        
        <div className="mb-4">
          {activeTab === 'cores' ? (
            <div className="flex flex-col md:flex-row gap-2">
              <div className="relative flex-1">
                <input
                  type="text"
                  placeholder="Search existing soul cores..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full px-3 py-2 pl-10 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
                />
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <svg className="h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" />
                  </svg>
                </div>
              </div>
              
              {player && (
                <div className="relative md:w-1/3">
                  <div className="flex space-x-2">
                    <div className="flex-1 relative">
                      <input
                        type="text"
                        placeholder="Add new soul core..."
                        value={creatureSearchTerm}
                        onChange={(e) => {
                          setCreatureSearchTerm(e.target.value);
                          setShowCreatureDropdown(true);
                        }}
                        onFocus={() => setShowCreatureDropdown(true)}
                        onBlur={() => setTimeout(() => setShowCreatureDropdown(false), 200)}
                        className="w-full px-3 py-2 pl-10 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
                      />
                      <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                        <svg className="h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                          <path fillRule="evenodd" d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z" clipRule="evenodd" />
                        </svg>
                      </div>
                      {showCreatureDropdown && filteredCreatures.length > 0 && (
                        <div className="absolute z-10 mt-1 w-full bg-white dark:bg-gray-800 shadow-lg rounded-md border border-gray-200 dark:border-gray-700 max-h-60 overflow-auto">
                          {filteredCreatures.map(creature => (
                            <div 
                              key={creature.endpoint}
                              className="px-4 py-2 cursor-pointer hover:bg-amber-50 dark:hover:bg-amber-900/30"
                              onClick={() => {
                                setSelectedCreature(creature);
                                setCreatureSearchTerm(creature.name);
                                setShowCreatureDropdown(false);
                              }}
                            >
                              <div className="font-medium text-gray-800 dark:text-gray-200">{creature.name}</div>
                              <div className="text-xs text-gray-500 dark:text-gray-400">{creature.plural_name}</div>
                            </div>
                          ))}
                        </div>
                      )}
                      {showCreatureDropdown && filteredCreatures.length === 0 && creatureSearchTerm && (
                        <div className="absolute z-10 mt-1 w-full bg-white dark:bg-gray-800 shadow-lg rounded-md border border-gray-200 dark:border-gray-700 p-4">
                          <p className="text-sm text-gray-500 dark:text-gray-400">
                            No matching creatures found or all matching creatures are already in the list.
                          </p>
                        </div>
                      )}
                    </div>
                    <button
                      onClick={handleAddSoulCore}
                      disabled={!selectedCreature || isAddingCore}
                      className={`px-4 py-2 rounded-md text-white font-medium ${
                        !selectedCreature || isAddingCore
                          ? 'bg-gray-400 cursor-not-allowed'
                          : 'bg-amber-600 hover:bg-amber-700'
                      }`}
                    >
                      {isAddingCore ? 'Adding...' : 'Add'}
                    </button>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="relative">
              <input
                type="text"
                placeholder="Search members..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full px-3 py-2 pl-10 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
              />
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg className="h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" />
                </svg>
              </div>
            </div>
          )}
        </div>
        
        {activeTab === 'cores' && (
          <div className="mb-6">
            {player ? (
              error && (
                <div className="mt-2 text-sm text-red-600 dark:text-red-400">
                  {error}
                </div>
              )
            ) : playerLoading ? (
              <div className="p-4 bg-amber-50 dark:bg-amber-900/30 rounded-md border border-amber-200 dark:border-amber-800 text-sm">
                <div className="flex items-center">
                  <svg className="animate-spin h-4 w-4 text-amber-600 mr-2" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  <p className="text-amber-800 dark:text-amber-400">
                    Loading player data...
                  </p>
                </div>
              </div>
            ) : (
              <div className="p-4 bg-amber-50 dark:bg-amber-900/30 rounded-md border border-amber-200 dark:border-amber-800 text-sm">
                <p className="text-amber-800 dark:text-amber-400">
                  You need to be logged in to add soul cores to this list.
                </p>
              </div>
            )}
          </div>
        )}
        
        {activeTab === 'cores' ? (
          <div>
            {filteredSoulCores.length > 0 ? (
              <div className="overflow-hidden border border-amber-200 dark:border-amber-800 rounded-lg">
                <table className="min-w-full divide-y divide-amber-200 dark:divide-amber-800">
                  <thead className="bg-amber-50 dark:bg-amber-900/30">
                    <tr>
                      <th 
                        scope="col" 
                        className="px-6 py-3 text-left text-xs font-medium text-amber-800 dark:text-amber-400 uppercase tracking-wider cursor-pointer"
                        onClick={() => handleSort('name')}
                      >
                        <div className="flex items-center">
                          Creature
                          {sortField === 'name' && (
                            <span className="ml-1">
                              {sortDirection === 'asc' ? '↑' : '↓'}
                            </span>
                          )}
                        </div>
                      </th>
                      <th 
                        scope="col" 
                        className="px-6 py-3 text-left text-xs font-medium text-amber-800 dark:text-amber-400 uppercase tracking-wider cursor-pointer"
                        onClick={() => handleSort('status')}
                      >
                        <div className="flex items-center">
                          Status
                          {sortField === 'status' && (
                            <span className="ml-1">
                              {sortDirection === 'asc' ? '↑' : '↓'}
                            </span>
                          )}
                        </div>
                      </th>
                      <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-amber-800 dark:text-amber-400 uppercase tracking-wider">
                        Obtained By
                      </th>
                      <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-amber-800 dark:text-amber-400 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white dark:bg-gray-800 divide-y divide-amber-100 dark:divide-amber-900/30">
                    {filteredSoulCores.map((core) => (
                      <tr key={core.id} className="hover:bg-amber-50 dark:hover:bg-amber-900/10">
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="font-medium text-gray-900 dark:text-white">{core.creatureName}</div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          {core.unlocked ? (
                            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300">
                              Unlocked
                            </span>
                          ) : core.obtained ? (
                            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300">
                              Obtained
                            </span>
                          ) : (
                            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300">
                              Missing
                            </span>
                          )}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          {core.unlocked ? (
                            <div className="text-sm text-gray-500 dark:text-gray-400">
                              {core.obtainedBy}
                            </div>
                          ) : core.obtained ? (
                            <div className="text-sm text-gray-500 dark:text-gray-400">
                              {core.obtainedBy}
                            </div>
                          ) : (
                            <div className="text-sm text-gray-500 dark:text-gray-400">-</div>
                          )}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                          <div className="flex flex-col space-y-2 items-end">
                            {!core.obtained && player && (
                              <button
                                onClick={() => handleMarkObtained(core.id)}
                                className="text-yellow-600 hover:text-yellow-900 dark:text-yellow-400 dark:hover:text-yellow-300"
                              >
                                Mark as Obtained
                              </button>
                            )}
                            {core.obtained && !core.unlocked && player && (
                              <button
                                onClick={() => handleMarkUnlocked(core.id)}
                                className="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
                              >
                                Mark as Unlocked
                              </button>
                            )}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center py-8">
                <p className="text-gray-500 dark:text-gray-400">No soul cores found matching your search.</p>
              </div>
            )}
          </div>
        ) : (
          <div>
            {filteredMembers.length > 0 ? (
              <div className="overflow-hidden border border-amber-200 dark:border-amber-800 rounded-lg">
                <table className="min-w-full divide-y divide-amber-200 dark:divide-amber-800">
                  <thead className="bg-amber-50 dark:bg-amber-900/30">
                    <tr>
                      <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-amber-800 dark:text-amber-400 uppercase tracking-wider">
                        Username
                      </th>
                      <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-amber-800 dark:text-amber-400 uppercase tracking-wider">
                        Character
                      </th>
                      <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-amber-800 dark:text-amber-400 uppercase tracking-wider">
                        World
                      </th>
                      <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-amber-800 dark:text-amber-400 uppercase tracking-wider">
                        Role
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white dark:bg-gray-800 divide-y divide-amber-100 dark:divide-amber-900/30">
                    {filteredMembers.map((member) => (
                      <tr key={member.id} className="hover:bg-amber-50 dark:hover:bg-amber-900/10">
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="font-medium text-gray-900 dark:text-white">{member.username}</div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="text-sm text-gray-500 dark:text-gray-400">{member.characterName}</div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="text-sm text-gray-500 dark:text-gray-400">{member.world}</div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          {member.isOwner ? (
                            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300">
                              Owner
                            </span>
                          ) : (
                            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300">
                              Member
                            </span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center py-8">
                <p className="text-gray-500 dark:text-gray-400">No members found matching your search.</p>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Share Modal */}
      {showShareModal && (
        <div 
          className="fixed inset-0 bg-black/20 flex items-center justify-center z-50"
          onClick={handleCloseShareModal}
        >
          <div 
            className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 max-w-md w-full"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium text-amber-800 dark:text-amber-500">Share List</h3>
              <button
                onClick={handleCloseShareModal}
                className="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300"
              >
                <svg className="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              Share this link or code with others to invite them to join your list.
            </p>
            
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Share Link
              </label>
              <div className="flex">
                <input
                  type="text"
                  readOnly
                  value={`${window.location.origin}/lists/join?code=${listDetails?.share_code}`}
                  className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-l-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
                />
                <button
                  onClick={handleCopyLink}
                  className="px-3 py-2 border border-l-0 border-gray-300 dark:border-gray-600 rounded-r-md bg-gray-100 dark:bg-gray-600 hover:bg-gray-200 dark:hover:bg-gray-500"
                >
                  <svg className="h-5 w-5 text-gray-500 dark:text-gray-300" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                </button>
              </div>
            </div>
            
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Share Code
              </label>
              <div className="flex">
                <input
                  type="text"
                  readOnly
                  value={listDetails?.share_code}
                  className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-l-md shadow-sm focus:outline-none focus:ring-amber-500 focus:border-amber-500 dark:bg-gray-700 dark:text-white"
                />
                <button
                  onClick={handleCopyCode}
                  className="px-3 py-2 border border-l-0 border-gray-300 dark:border-gray-600 rounded-r-md bg-gray-100 dark:bg-gray-600 hover:bg-gray-200 dark:hover:bg-gray-500"
                >
                  <svg className="h-5 w-5 text-gray-500 dark:text-gray-300" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                </button>
              </div>
            </div>
            
            {copySuccess && (
              <div className="p-2 bg-green-50 dark:bg-green-900/30 rounded-md border border-green-200 dark:border-green-800 text-sm text-green-700 dark:text-green-300 mb-4">
                Copied to clipboard!
              </div>
            )}
            
            <div className="flex justify-end">
              <button
                onClick={handleCloseShareModal}
                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}