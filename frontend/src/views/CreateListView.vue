<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import type { Character as TibiaCharacter } from '../services/tibiadata'
import { v4 as uuidv4 } from 'uuid'
import axios from 'axios'

interface DBCharacter extends TibiaCharacter {
  id: string
  user_id: string
}

interface BaseListRequest {
  name: string
}

interface AnonymousListRequest extends BaseListRequest {
  session_token: string
  character_name: string
  world: string
}

interface ExistingCharacterListRequest extends BaseListRequest {
  character_id: string
  user_id: string
}

interface NewCharacterListRequest extends BaseListRequest {
  character_name: string
  world: string
  user_id: string
}

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const character = ref<TibiaCharacter | DBCharacter | null>(null)
const listName = ref('')
const error = ref('')
const loading = ref(false)
const useExisting = ref(false)

onMounted(() => {
  try {
    const characterData = route.query.character as string
    if (!characterData) {
      throw new Error('No character data provided')
    }
    character.value = JSON.parse(characterData)
    useExisting.value = route.query.useExisting === 'true'
  } catch (e) {
    console.error('Failed to parse character data:', e)
    router.push('/')
  }
})

const isAnonymousListRequest = (
  data:
    | BaseListRequest
    | AnonymousListRequest
    | ExistingCharacterListRequest
    | NewCharacterListRequest,
): data is AnonymousListRequest => {
  return 'session_token' in data
}

const createList = async () => {
  if (!character.value) return

  loading.value = true
  error.value = ''

  try {
    let requestData:
      | BaseListRequest
      | AnonymousListRequest
      | ExistingCharacterListRequest
      | NewCharacterListRequest = {
      name: listName.value,
    }

    // Case 1: First-time user with session token
    if (!userStore.isAuthenticated) {
      const sessionToken = uuidv4()
      requestData = {
        ...requestData,
        session_token: sessionToken,
        character_name: character.value.name,
        world: character.value.world,
      }
    }
    // Case 2a: Existing character with ID
    else if (useExisting.value && 'id' in character.value) {
      requestData = {
        ...requestData,
        character_id: character.value.id,
        user_id: userStore.userId,
      }
    }
    // Case 2b: New character for existing user
    else {
      requestData = {
        ...requestData,
        character_name: character.value.name,
        world: character.value.world,
        user_id: userStore.userId,
      }
    }

    const response = await axios.post('/api/lists', requestData)

    // Set user state with anonymous user data if this is a first-time user
    if (!userStore.isAuthenticated && isAnonymousListRequest(requestData)) {
      userStore.setUser({
        session_token: requestData.session_token,
        id: response.data.author_id,
        is_anonymous: true,
      })
    }

    router.push('/')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to create list'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-2xl mx-auto px-4 py-8">
    <div v-if="character" class="bg-white rounded-lg shadow p-6">
      <h1 class="text-2xl font-semibold mb-6">Create New List</h1>

      <div class="mb-6 p-4 bg-gray-50 rounded-lg">
        <h2 class="text-lg font-medium mb-3">Character Details</h2>
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

      <form @submit.prevent="createList" class="space-y-4">
        <div>
          <label for="listName" class="block text-sm font-medium text-gray-700 mb-1">
            List Name
          </label>
          <input
            id="listName"
            v-model="listName"
            type="text"
            required
            :disabled="loading"
            placeholder="Enter list name"
            class="w-full p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
          />
        </div>

        <div v-if="error" class="text-red-500 text-sm">{{ error }}</div>

        <div class="flex gap-4">
          <button
            type="submit"
            :disabled="loading"
            class="flex-1 px-4 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 disabled:bg-gray-400"
          >
            {{ loading ? 'Creating...' : 'Create List' }}
          </button>
          <button
            type="button"
            :disabled="loading"
            @click="router.push('/')"
            class="px-4 py-2 text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 disabled:bg-gray-100"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
