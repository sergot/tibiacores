<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { useListsStore } from '../stores/lists'
import type { Character as TibiaCharacter } from '../services/tibiadata'
import { tibiaDataService } from '../services/tibiadata'
import axios from 'axios'

interface DBCharacter {
  id: string
  user_id: string
  name: string
  world: string
}

const router = useRouter()
const userStore = useUserStore()
const listsStore = useListsStore()

const shareCode = ref('')
const characterName = ref('')
const error = ref('')
const loading = ref(false)
const existingCharacters = ref<DBCharacter[]>([])
const filteredCharacters = ref<DBCharacter[]>([])
const showDropdown = ref(false)
const selectedCharacter = ref<DBCharacter | null>(null)
const isExistingCharacter = ref(false)
const isSelectingCharacter = ref(false)

onMounted(async () => {
  if (userStore.isAuthenticated) {
    try {
      const response = await axios.get<DBCharacter[]>(`/api/users/${userStore.userId}/characters`)
      existingCharacters.value = response.data
    } catch (e) {
      console.error('Failed to fetch characters:', e)
    }
  }
})

const extractShareCode = (input: string): string => {
  // Try to extract share code from URL if it's a URL
  try {
    const url = new URL(input)
    const parts = url.pathname.split('/')
    return parts[parts.length - 1]
  } catch {
    // If not a URL, return as is
    return input
  }
}

const filterCharacters = (input: string) => {
  if (!input) {
    filteredCharacters.value = existingCharacters.value
    return
  }

  filteredCharacters.value = existingCharacters.value.filter((char) =>
    char.name.toLowerCase().includes(input.toLowerCase()),
  )
}

const handleCharacterInput = (event: Event) => {
  const input = (event.target as HTMLInputElement).value
  characterName.value = input
  selectedCharacter.value = null
  filterCharacters(input)
  showDropdown.value = true
}

const handleCharacterFocus = () => {
  showDropdown.value = true
  filterCharacters(characterName.value)
}

const handleCharacterBlur = () => {
  setTimeout(() => {
    showDropdown.value = false
  }, 200)
}

const selectCharacter = (character: DBCharacter) => {
  characterName.value = character.name
  selectedCharacter.value = character
  showDropdown.value = false
}

const joinList = async () => {
  error.value = ''
  loading.value = true

  try {
    const code = extractShareCode(shareCode.value)

    if (!code) {
      error.value = 'Please enter a valid share code or URL'
      return
    }

    let requestData = {}

    // If user selected existing character
    if (selectedCharacter.value) {
      requestData = {
        character_id: selectedCharacter.value.id,
      }
    }
    // If user is entering a new character
    else if (characterName.value) {
      // For new character, verify it exists in Tibia
      const character = await tibiaDataService.getCharacter(characterName.value)
      requestData = {
        character_name: character.name,
        world: character.world,
      }
    }

    const response = await axios.post(`/api/lists/join/${code}`, requestData)
    
    // For anonymous users, get the token from response header
    const authToken = response.headers['x-auth-token']
    if (authToken && !userStore.isAuthenticated) {
      userStore.setUser({
        session_token: authToken,
        id: response.data.author_id,
        has_email: false,
      })
    }

    // Fetch updated lists and redirect to the joined list
    await listsStore.fetchUserLists()
    router.push(`/lists/${response.data.id}`)
  } catch (err) {
    if (axios.isAxiosError(err)) {
      error.value = err.response?.data?.message || 'Failed to join list'
    } else {
      error.value = 'Failed to join list'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="p-6 rounded-lg">
    <h2 class="mb-4 text-2xl">Join a list</h2>

    <div v-if="error" class="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <form @submit.prevent="joinList" class="space-y-4">
      <div>
        <label for="shareCode" class="block text-sm font-medium text-gray-700 mb-1">
          Share Code or URL
        </label>
        <input
          id="shareCode"
          v-model="shareCode"
          type="text"
          placeholder="Enter share code or paste URL"
          required
          :disabled="loading"
          class="w-full p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
        />
      </div>

      <div v-if="!isSelectingCharacter && !isExistingCharacter" class="space-y-2">
        <button
          v-if="userStore.isAuthenticated && existingCharacters.length > 0"
          type="button"
          class="text-sm text-indigo-600 hover:text-indigo-800"
          @click="isExistingCharacter = true"
        >
          Use existing character
        </button>
        <button
          type="button"
          class="text-sm text-indigo-600 hover:text-indigo-800"
          @click="isSelectingCharacter = true"
        >
          Add new character
        </button>
      </div>

      <div v-if="isSelectingCharacter || isExistingCharacter" class="relative">
        <label for="characterName" class="block text-sm font-medium text-gray-700 mb-1">
          Character Name
        </label>
        <input
          id="characterName"
          v-model="characterName"
          type="text"
          :placeholder="
            userStore.isAuthenticated
              ? 'Enter new character or select existing'
              : 'Enter character name'
          "
          required
          :disabled="loading"
          @input="handleCharacterInput"
          @focus="handleCharacterFocus"
          @blur="handleCharacterBlur"
          class="w-full p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
        />

        <!-- Dropdown for existing characters -->
        <div
          v-if="showDropdown && userStore.isAuthenticated && filteredCharacters.length > 0"
          class="absolute top-full left-0 right-0 mt-1 bg-white border border-gray-300 rounded-md shadow-lg z-10 max-h-60 overflow-y-auto"
        >
          <div
            v-for="character in filteredCharacters"
            :key="character.name"
            @click="selectCharacter(character)"
            @mousedown.prevent
            class="p-2 hover:bg-gray-100 cursor-pointer flex justify-between items-center"
          >
            <span>{{ character.name }}</span>
            <span class="text-sm text-gray-500">{{ character.world }}</span>
          </div>
        </div>
      </div>

      <div class="flex gap-4">
        <button
          type="submit"
          :disabled="loading || (!characterName.value && (isSelectingCharacter || isExistingCharacter))"
          class="flex-1 px-4 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 disabled:bg-gray-400 flex items-center justify-center"
        >
          <svg
            v-if="loading"
            class="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          {{ loading ? 'Joining...' : 'Join List' }}
        </button>

        <button
          v-if="isSelectingCharacter || isExistingCharacter"
          type="button"
          :disabled="loading"
          @click="isSelectingCharacter = false; isExistingCharacter = false; characterName.value = ''"
          class="px-4 py-2 text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
        >
          Cancel
        </button>
      </div>
    </form>
  </div>
</template>
