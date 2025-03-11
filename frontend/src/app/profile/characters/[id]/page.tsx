"use client";

import React, { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { usePlayer } from "@/contexts/PlayerContext";
import { playerApi } from "@/services/api";
import listApi from "@/services/listApi";
import Link from "next/link";

interface SoulCore {
  id: string;
  creature_name: string;
  list_id: string;
  list_name: string;
}

interface CharacterResponse {
  character: {
    id: string;
    name: string;
    world: string;
  };
  soul_cores: SoulCore[];
}

interface RouteParams {
  id: string;
}

export default function CharacterDetailsPage({ params }: { params: RouteParams | Promise<RouteParams> }) {
  // Properly unwrap params using React.use()
  const resolvedParams = React.use(params as any) as RouteParams;
  const characterId = resolvedParams.id;
  
  const router = useRouter();
  const { player, characters, fetchCharacters, loading, error, fetchAnonymousPlayer } = usePlayer();
  const [character, setCharacter] = useState<any>(null);
  const [soulCores, setSoulCores] = useState<SoulCore[]>([]);
  const [loadingCores, setLoadingCores] = useState(true);
  const [coreError, setCoreError] = useState("");
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
      // Only fetch characters once
      if (!hasFetchedCharactersRef.current && characters.length === 0) {
        hasFetchedCharactersRef.current = true;
        fetchCharacters();
      }
    }
  }, [player, characters, fetchCharacters, fetchAnonymousPlayer, router]);

  // Find the character in the characters array
  useEffect(() => {
    if (characters && characters.length > 0) {
      const foundCharacter = characters.find(c => c.id === characterId);
      if (foundCharacter) {
        setCharacter(foundCharacter);
      } else {
        setCoreError("Character not found");
        router.push("/profile");
      }
    }
  }, [characters, characterId, router]);

  // Fetch all lists where this character is a member using the new endpoint
  useEffect(() => {
    if (character && player) {
      const fetchSoulCores = async () => {
        setLoadingCores(true);
        setCoreError("");
        try {
          // Get all lists and soul cores for this character
          const response = await listApi.getListsByCharacterID(characterId) as CharacterResponse;
          
          if (response && response.soul_cores) {
            setSoulCores(response.soul_cores);
          } else {
            setSoulCores([]);
          }
        } catch (error: any) {
          console.error("Error fetching soul cores:", error);
          setCoreError(error.message || "Failed to fetch soul cores");
        } finally {
          setLoadingCores(false);
        }
      };
      
      fetchSoulCores();
    }
  }, [character, player, characterId]);

  if (loading || loadingCores) {
    return (
      <div className="container mx-auto p-4">
        <h1 className="text-2xl font-bold mb-4">Character Details</h1>
        <p>Loading...</p>
      </div>
    );
  }

  if (error || coreError || !player || !character) {
    return (
      <div className="container mx-auto p-4">
        <h1 className="text-2xl font-bold mb-4">Character Details</h1>
        <p className="text-red-500">{error || coreError || "Error loading character details"}</p>
        <div className="mt-4">
          <Link href="/profile" className="text-blue-500 hover:underline">
            Back to Profile
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-4">
      <div className="mb-4">
        <Link href="/profile" className="text-blue-500 hover:underline">
          &larr; Back to Profile
        </Link>
      </div>
      
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
        <h1 className="text-2xl font-bold mb-2">{character.name}</h1>
        <p className="text-gray-600 dark:text-gray-400 mb-4">World: {character.world}</p>
        
        {player && player.characters && player.characters.length > 0 && 
          player.characters[0]?.id === character.id && (
          <div className="mb-4">
            <span className="bg-green-100 text-green-800 px-2 py-1 rounded-full dark:bg-green-900 dark:text-green-100">
              Main Character
            </span>
          </div>
        )}
      </div>
      
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4">Unlocked Soul Cores</h2>
        
        {soulCores.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {soulCores.map((core) => (
              <div key={`${core.id}`} className="border rounded-md p-4 dark:border-gray-700">
                <div className="font-medium">{core.creature_name}</div>
                <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  From list: <Link href={`/lists/${core.list_id}`} className="text-blue-500 hover:underline">
                    {core.list_name}
                  </Link>
                </p>
              </div>
            ))}
          </div>
        ) : (
          <p>No soul cores unlocked yet.</p>
        )}
      </div>
    </div>
  );
} 