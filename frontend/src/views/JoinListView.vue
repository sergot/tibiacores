<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
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

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const listsStore = useListsStore()

const characterName = ref('')
const error = ref('')
const loading = ref(true)
const existingCharacters = ref<DBCharacter[]>([])
const selectedCharacter = ref<DBCharacter | null>(null)
const character = ref<TibiaCharacter | null>(null)
const listPreview = ref<ListPreview | null>(null)
const isExistingCharacter = ref(false)

onMounted(async () => {
  try {
    // Try to get list preview
    const response = await axios.get<ListPreview>(`/api/lists/preview/${props.share_code}`)
    listPreview.value = response.data

    // If user is authenticated, fetch their characters from the same world
    if (userStore.isAuthenticated) {
      const charResponse = await axios.get<DBCharacter[]>(`/api/users/${userStore.userId}/characters`)
      existingCharacters.value = charResponse.data.filter(char => char.world === listPreview.value?.world)
    }

    // If character name was provided in query, verify it
    const queryCharacter = route.query.character as string
    if (queryCharacter) {
      characterName.value = queryCharacter
      await verifyCharacter()
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

const verifyCharacter = async () => {
  error.value = ''
  loading.value = true
  character.value = null

  try {
    const tibiaCharacter = await tibiaDataService.getCharacter(characterName.value)
    
    // Verify world matches
    if (tibiaCharacter.world !== listPreview.value?.world) {
      throw new Error(`Character must be from ${listPreview.value?.world}`)
    }

    character.value = tibiaCharacter
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to verify character'
    character.value = null
  } finally {
    loading.value = false
  }
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
    // If user verified a new character
    else if (character.value) {
      requestData = {
        character_name: character.value.name,
        world: character.value.world,
      }
    } else {
      throw new Error('Please verify your character first')
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
      error.value = err instanceof Error ? err.message : 'Failed to join list'
    }
  } finally {
    loading.value = false
  }
}

const selectExistingCharacter = (char: DBCharacter) => {
  selectedCharacter.value = char
  character.value = null
  characterName.value = char.name
}
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)] flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8 bg-gray-100">
    <main class="max-w-xl w-full space-y-8">
      <div v-if="loading && !error" class="text-center py-12">
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
              <h3 class="text-lg font-medium">Character Information</h3>
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
                  @click="isExistingCharacter = false; selectedCharacter = null"
                >
                  Add new character
                </button>
              </div>
            </div>

            <!-- Existing characters selection -->
            <div v-if="isExistingCharacter" class="mb-4">
              <div class="grid gap-2">
                <button
                  v-for="char in existingCharacters"
                  :key="char.id"
                  @click="selectExistingCharacter(char)"
                  class="p-3 text-left border rounded-lg hover:bg-gray-50"
                  :class="selectedCharacter?.id === char.id ? 'border-indigo-500 bg-indigo-50' : 'border-gray-200'"
                >
                  <div class="font-medium">{{ char.name }}</div>
                  <div class="text-sm text-gray-500">{{ char.world }}</div>
                </button>
              </div>
            </div>

            <!-- New character verification -->
            <div v-else>
              <div class="flex gap-2">
                <input
                  v-model="characterName"
                  type="text"
                  :placeholder="`Enter character name from ${listPreview.world}`"
                  class="flex-1 p-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
                <button
                  @click="verifyCharacter"
                  :disabled="!characterName || loading"
                  class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:bg-gray-400"
                >
                  Verify
                </button>
              </div>

              <!-- Character details after verification -->
              <div v-if="character" class="mt-4 p-4 bg-green-50 border border-green-200 rounded-lg">
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <p class="text-gray-600">Name</p>
                    <p class="font-medium">{{ character.name }}</p>
                  </div>
                  <div>
                    <p class="text-gray-600">World</p>
                    <p class="font-medium">{{ character.world }}</p>
                  </div>
                  <div>
                    <p class="text-gray-600">Level</p>
                    <p class="font-medium">{{ character.level }}</p>
                  </div>
                  <div>
                    <p class="text-gray-600">Vocation</p>
                    <p class="font-medium">{{ character.vocation }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="flex gap-4">
            <button
              @click="handleJoin"
              :disabled="loading || (!selectedCharacter && !character)"
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