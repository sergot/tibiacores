<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useListsStore } from '@/stores/lists'
import { useI18n } from 'vue-i18n'
import type { Character as TibiaCharacter } from '../services/tibiadata'
import axios from 'axios'
import ClaimSuggestion from '../components/ClaimSuggestion.vue'

interface DBCharacter extends TibiaCharacter {
  id: string
  user_id: string
}

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const listsStore = useListsStore()
const character = ref<TibiaCharacter | DBCharacter | null>(null)
const listName = ref('')
const error = ref('')
const loading = ref(false)
const useExisting = ref(false)
const showNameConflict = ref(false)

onMounted(() => {
  try {
    const characterData = route.query.character as string
    if (!characterData) {
      throw new Error('No character data provided')
    }
    character.value = JSON.parse(characterData)
    useExisting.value = route.query.useExisting === 'true'
  } catch (e) {
    error.value = 'Invalid character data'
    console.error('Failed to parse character data:', e)
    setTimeout(() => router.push('/'), 2000)
  }
})

const handleSubmit = async () => {
  if (!character.value || !listName.value) return

  loading.value = true
  error.value = ''
  showNameConflict.value = false

  try {
    const requestData = {
      name: listName.value,
      ...(useExisting.value && 'id' in character.value && {
        character_id: character.value.id,
      }),
      ...(!useExisting.value && {
        character_name: character.value.name,
        world: character.value.world,
      }),
    }

    const response = await axios.post('/lists', requestData)

    // For anonymous users, get the token from response header
    const authToken = response.headers['x-auth-token']
    if (authToken && !userStore.isAuthenticated) {
      userStore.setUser({
        session_token: authToken,
        id: response.data.author_id,
        has_email: false,
      })
    }

    // Fetch updated lists and redirect to home
    await listsStore.fetchUserLists()
    router.push('/')
  } catch (err) {
    if (axios.isAxiosError(err) && err.response?.status === 409) {
      showNameConflict.value = true
    } else if (axios.isAxiosError(err) && err.response) {
      error.value = err.response.data.message
    } else {
      error.value = 'Network error. Please try again.'
    }
  } finally {
    loading.value = false
  }
}

const handleTryDifferent = () => {
  router.push('/')
}
</script>

<template>
  <div class="max-w-2xl mx-auto px-4 py-8">
    <div v-if="error" class="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <ClaimSuggestion
      v-if="showNameConflict && character"
      :character-name="character.name"
      @try-different="handleTryDifferent"
    />

    <div v-else-if="character" class="bg-white rounded-lg shadow p-6">
      <h1 class="text-2xl font-semibold mb-6">{{ t('createList.title') }}</h1>

      <div class="mb-6 p-4 bg-gray-50 rounded-lg">
        <h2 class="text-lg font-medium mb-3">{{ t('characterDetails.title') }}</h2>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-gray-600">{{ t('characterDetails.information.name') }}</p>
            <p class="font-medium">{{ character.name }}</p>
          </div>
          <div>
            <p class="text-gray-600">{{ t('characterDetails.information.world') }}</p>
            <p class="font-medium">{{ character.world }}</p>
          </div>
          <div>
            <p class="text-gray-600">{{ t('characterDetails.information.level') }}</p>
            <p class="font-medium">{{ character.level }}</p>
          </div>
          <div>
            <p class="text-gray-600">{{ t('characterDetails.information.vocation') }}</p>
            <p class="font-medium">{{ character.vocation }}</p>
          </div>
        </div>
      </div>

      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <label for="listName" class="block text-sm font-medium text-gray-700 mb-1">
            {{ t('lists.title') }}
          </label>
          <input
            id="listName"
            v-model="listName"
            type="text"
            required
            :disabled="loading"
            :placeholder="t('createList.form.characterNamePlaceholder')"
            class="w-full p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
          />
        </div>

        <div class="flex gap-4">
          <button
            type="submit"
            :disabled="loading"
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
            {{ loading ? t('createList.form.verifying') : t('createList.form.submit') }}
          </button>
          <button
            type="button"
            :disabled="loading"
            @click="router.push('/')"
            class="px-4 py-2 text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 disabled:bg-gray-100"
          >
            {{ t('characterDetails.confirmDelete.cancel') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
