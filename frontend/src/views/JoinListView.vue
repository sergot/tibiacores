<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useListsStore } from '@/stores/lists'
import type { Character as TibiaCharacter } from '../services/tibiadata'
import { tibiaDataService } from '../services/tibiadata'
import axios from 'axios'

interface DBCharacter {
  id: string
  user_id: string
  name: string
  world: string
}

interface ListPreview {
  id: string
  name: string
  world: string
  member_count: number
}

const props = defineProps<{
  share_code: string
}>()

const router = useRouter()
const userStore = useUserStore()
const listsStore = useListsStore()

const characterName = ref('')
const error = ref('')
const loading = ref(true)
const existingCharacters = ref<DBCharacter[]>([])
const filteredCharacters = ref<DBCharacter[]>([])
const showDropdown = ref(false)
const selectedCharacter = ref<DBCharacter | null>(null)
const isExistingCharacter = ref(false)
const listPreview = ref<ListPreview | null>(null)

onMounted(async () => {
  try {
    // Try to get list preview
    const response = await axios.get<ListPreview>(`/api/lists/preview/${props.share_code}`)
    listPreview.value = response.data

    // If user is authenticated, fetch their characters
    if (userStore.isAuthenticated) {
      const charResponse = await axios.get<DBCharacter[]>(`/api/users/${userStore.userId}/characters`)
      existingCharacters.value = charResponse.data.filter(char => char.world === listPreview.value?.world)
    }
  } catch (e) {
    if (axios.isAxiosError(e)) {
      error.value = e.response?.data?.message || 'Failed to load list details'
    } else {
      error.value = 'Failed to load list details'
    }
    setTimeout(() => router.push('/'), 2000)
  } finally {
    loading.value = false
  }
})

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

const handleJoin = async () => {
  error.value = ''
  loading.value = true

  try {
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
      
      // Verify world matches
      if (character.world !== listPreview.value?.world) {
        error.value = `Character must be from ${listPreview.value?.world}`
        loading.value = false
        return
      }

      requestData = {
        character_name: character.name,
        world: character.world,
      }
    }

    const response = await axios.post(`/api/lists/join/${props.share_code}`, requestData)
    
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
  <div class="min-h-[calc(100vh-8rem)] flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8 bg-gray-100">
    <main class="max-w-xl w-full space-y-8">
      <div v-if="loading" class="text-center py-12">
        <div
          class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full mx-auto mb-4"
        ></div>
        <p class="text-gray-600 font-medium">Loading list details...</p>
      </div>

      <div v-else-if="error" class="text-center py-12">
        <div class="bg-red-50 rounded-lg p-6">
          <p class="text-red-600 font-medium mb-4">{{ error }}</p>
          <p class="text-gray-600">Redirecting you to home...</p>
        </div>
      </div>

      <div v-else-if="listPreview" class="bg-white rounded-lg shadow p-8">
        <div class="text-center mb-8">
          <h1 class="text-3xl font-bold text-gray-900">Join List</h1>
          <p class="mt-2 text-gray-600">
            You've been invited to join <span class="font-medium text-gray-800">{{ listPreview.name }}</span>
          </p>
        </div>

        <div class="mb-8 p-4 bg-gray-50 rounded-lg">
          <h2 class="text-lg font-medium mb-3">List Details</h2>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-gray-600">Name</p>
              <p class="font-medium">{{ listPreview.name }}</p>
            </div>
            <div>
              <p class="text-gray-600">World</p>
              <p class="font-medium">{{ listPreview.world }}</p>
            </div>
            <div>
              <p class="text-gray-600">Members</p>
              <p class="font-medium">{{ listPreview.member_count }}</p>
            </div>
          </div>
        </div>

        <div class="space-y-6">
          <div>
            <div class="flex justify-between items-center mb-2">
              <h3 class="text-lg font-medium">Choose Character</h3>
              <div class="space-x-4">
                <button
                  v-if="userStore.isAuthenticated && existingCharacters.length > 0 && !isExistingCharacter"
                  type="button"
                  class="text-sm text-indigo-600 hover:text-indigo-800"
                  @click="isExistingCharacter = true"
                >
                  Use existing character
                </button>
                <button
                  v-if="isExistingCharacter"
                  type="button"
                  class="text-sm text-indigo-600 hover:text-indigo-800"
                  @click="isExistingCharacter = false; selectedCharacter = null; characterName = ''"
                >
                  Add new character
                </button>
              </div>
            </div>

            <div class="relative">
              <input
                v-model="characterName"
                type="text"
                :placeholder="isExistingCharacter ? 'Select existing character' : `Enter character name from ${listPreview.world}`"
                required
                :disabled="loading"
                @input="handleCharacterInput"
                @focus="handleCharacterFocus"
                @blur="handleCharacterBlur"
                class="w-full p-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />

              <!-- Dropdown for existing characters -->
              <div
                v-if="showDropdown && isExistingCharacter && filteredCharacters.length > 0"
                class="absolute top-full left-0 right-0 mt-1 bg-white border border-gray-300 rounded-lg shadow-lg z-10 max-h-60 overflow-y-auto"
              >
                <div
                  v-for="character in filteredCharacters"
                  :key="character.name"
                  @click="selectCharacter(character)"
                  @mousedown.prevent
                  class="p-3 hover:bg-gray-50 cursor-pointer"
                >
                  <div class="font-medium">{{ character.name }}</div>
                </div>
              </div>
            </div>
          </div>

          <div class="flex gap-4">
            <button
              @click="handleJoin"
              :disabled="loading || !characterName"
              class="flex-1 px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:bg-gray-400 transition-colors duration-200 flex items-center justify-center"
            >
              <svg
                v-if="loading"
                class="animate-spin -ml-1 mr-2 h-5 w-5 text-white"
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
              @click="router.push('/')"
              :disabled="loading"
              class="px-6 py-3 text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 disabled:bg-gray-100 transition-colors duration-200"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>