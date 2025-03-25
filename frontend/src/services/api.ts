import axios from 'axios'
import type { AxiosInstance, AxiosResponse, AxiosError } from 'axios'
import { useUserStore } from '@/stores/user'

interface AuthResponse {
  session_token: string
  id: string
  has_email: boolean
}

interface ListDetails {
  id: string
  author_id: string
  name: string
  share_code: string
  world: string
  created_at: string
  updated_at: string
  members: {
    user_id: string
    character_name: string
    obtained_count: number
    unlocked_count: number
  }[]
  soul_cores: {
    creature_id: string
    creature_name: string
    status: 'obtained' | 'unlocked'
    added_by: string | null
    added_by_user_id: string
  }[]
  total_cores: number
}

interface Creature {
  id: string
  name: string
}

interface Character {
  id: string
  name: string
  world: string
}

interface ClaimResponse {
  claim_id: string
  verification_code: string
  status: string
  token?: string
  claimer_id?: string
}

interface ListPreview {
  id: string
  name: string
  world: string
  member_count: number
}

interface SoulCore {
  creature_id: string
  creature_name: string
}

interface PendingSuggestion {
  character_id: string
  character_name: string
  suggestion_count: number
}

interface APIErrorResponse {
  message: string
  error?: string
}

class ApiService {
  private api: AxiosInstance

  constructor() {
    this.api = axios.create({
      baseURL: import.meta.env.VITE_API_URL || '/api',
    })

    // Request interceptor
    this.api.interceptors.request.use((config) => {
      const userStore = useUserStore()
      if (userStore.token) {
        config.headers.Authorization = `Bearer ${userStore.token}`
      }
      return config
    })

    // Response interceptor
    this.api.interceptors.response.use(
      (response) => response,
      (error: AxiosError<APIErrorResponse>) => {
        if (error.response?.status === 401) {
          const userStore = useUserStore()
          userStore.clearUser()
          if (!window.location.pathname.match(/^\/(signin|signup)/)) {
            window.location.href = '/signin'
          }
        }
        return Promise.reject(error)
      }
    )
  }

  // Generic API methods with proper typing
  async get<T>(url: string): Promise<T> {
    const response: AxiosResponse<T> = await this.api.get(url)
    return response.data
  }

  async post<T>(url: string, data?: unknown): Promise<T> {
    const response: AxiosResponse<T> = await this.api.post(url, data)
    return response.data
  }

  async put<T>(url: string, data?: unknown): Promise<T> {
    const response: AxiosResponse<T> = await this.api.put(url, data)
    return response.data
  }

  async delete<T>(url: string): Promise<T> {
    const response: AxiosResponse<T> = await this.api.delete(url)
    return response.data
  }

  // Specific API endpoints with typed responses
  auth = {
    login: (email: string, password: string) => 
      this.post<AuthResponse>('/login', { email, password }),
    signup: (data: { email: string; password: string; user_id?: string }) => 
      this.post<AuthResponse>('/signup', data),
  }

  lists = {
    getAll: (userId: string) => 
      this.get<ListDetails[]>(`/users/${userId}/lists`),
    get: (id: string) => 
      this.get<ListDetails>(`/lists/${id}`),
    create: (data: unknown) => 
      this.post<ListDetails>('/lists', data),
    addSoulcore: (listId: string, data: { creature_id: string; status: 'obtained' | 'unlocked' }) => 
      this.post<void>(`/lists/${listId}/soulcores`, data),
    updateSoulcore: (listId: string, data: { creature_id: string; status: 'obtained' | 'unlocked' }) => 
      this.put<void>(`/lists/${listId}/soulcores`, data),
    removeSoulcore: (listId: string, creatureId: string) => 
      this.delete<void>(`/lists/${listId}/soulcores/${creatureId}`),
    getPreview: (shareCode: string) => 
      this.get<ListPreview>(`/lists/preview/${shareCode}`),
    join: (shareCode: string, data: unknown) => 
      this.post<ListDetails>(`/lists/join/${shareCode}`, data),
  }

  characters = {
    get: (characterId: string) => 
      this.get<Character>(`/characters/${characterId}`),
    getAll: (userId: string) => 
      this.get<Character[]>(`/users/${userId}/characters`),
    getSoulcores: (characterId: string) => 
      this.get<SoulCore[]>(`/characters/${characterId}/soulcores`),
    removeSoulcore: (characterId: string, creatureId: string) => 
      this.delete<void>(`/characters/${characterId}/soulcores/${creatureId}`),
    getSuggestions: (characterId: string) => 
      this.get<SoulCore[]>(`/characters/${characterId}/suggestions`),
    acceptSuggestion: (characterId: string, data: { creature_id: string }) => 
      this.post<void>(`/characters/${characterId}/suggestions/accept`, data),
    dismissSuggestion: (characterId: string, data: { creature_id: string }) => 
      this.post<void>(`/characters/${characterId}/suggestions/dismiss`, data),
  }

  creatures = {
    getAll: () => this.get<Creature[]>('/creatures'),
  }

  claims = {
    create: (data: { character_name: string }) => 
      this.post<ClaimResponse>('/claims', data),
    get: (claimId: string) => 
      this.get<ClaimResponse>(`/claims/${claimId}`),
  }

  suggestions = {
    getPending: () => this.get<PendingSuggestion[]>('/pending-suggestions'),
  }
}

export const api = new ApiService()
export type {
  AuthResponse,
  ListDetails,
  Creature,
  Character,
  ClaimResponse,
  ListPreview,
  SoulCore,
  PendingSuggestion,
  APIErrorResponse,
}