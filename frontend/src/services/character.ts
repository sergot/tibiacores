// Character service for fetching data from TibiaData API

// TibiaData API base URL
const TIBIA_DATA_API_URL = 'https://api.tibiadata.com/v4';

// Define the TibiaCharacter interface
export interface TibiaCharacter {
  name: string;
  world: string;
  level: number;
  vocation: string;
}

/**
 * Fetch character data from TibiaData API
 * @param characterName The name of the character to fetch
 * @returns The character data or null if not found
 */
export const fetchCharacterData = async (characterName: string): Promise<TibiaCharacter | null> => {
  try {
    // Encode the character name for URL
    const encodedName = encodeURIComponent(characterName);
    
    // Fetch character data from TibiaData API
    const response = await fetch(`${TIBIA_DATA_API_URL}/character/${encodedName}`);
    
    if (!response.ok) {
      if (response.status === 404) {
        return null; // Character not found
      }
      throw new Error(`API error: ${response.status}`);
    }
    
    const data = await response.json();
    
    // Check if character exists
    if (!data.character || !data.character.character) {
      return null;
    }
    
    const characterData = data.character.character;
    
    // Return formatted character data
    return {
      name: characterData.name,
      world: characterData.world,
      level: characterData.level,
      vocation: characterData.vocation
    };
  } catch (error) {
    console.error('Error fetching character data:', error);
    throw new Error('Failed to fetch character data from TibiaData API');
  }
};

/**
 * Validate if a character exists
 * @param characterName The name of the character to validate
 * @returns The character data
 * @throws Error if character not found
 */
export const validateCharacter = async (characterName: string): Promise<TibiaCharacter> => {
  const character = await fetchCharacterData(characterName);
  
  if (!character) {
    throw new Error(`Character "${characterName}" not found`);
  }
  
  return character;
};

export default {
  fetchCharacterData,
  validateCharacter
}; 