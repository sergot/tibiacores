import axios from 'axios'

interface Character {
  name: string
  level: number
  vocation: string
  world: string
  residence: string
  lastLogin: string
  accountStatus: string
  comment?: string
}

class TibiaDataService {
  private readonly baseUrl = 'https://api.tibiadata.com/v4'

  async getCharacter(name: string): Promise<Character> {
    try {
      const response = await axios.get(`${this.baseUrl}/character/${name}`)
      const characterData = response.data.character.character

      return {
        name: characterData.name,
        level: characterData.level,
        vocation: characterData.vocation,
        world: characterData.world,
        residence: characterData.residence,
        lastLogin: characterData.last_login,
        accountStatus: characterData.account_status,
        comment: characterData.comment,
      }
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 404) {
        throw new Error('Character not found')
      }
      throw new Error('Failed to fetch character data')
    }
  }

  async verifyCharacterClaim(name: string, verificationCode: string): Promise<boolean> {
    try {
      const character = await this.getCharacter(name)
      return character.comment?.includes(verificationCode) || false
    } catch {
      return false
    }
  }
}

export const tibiaDataService = new TibiaDataService()
export type { Character }
