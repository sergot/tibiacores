<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { tibiaDataService, type Character as TibiaCharacter } from '../services/tibiadata'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import axios from 'axios'
import ClaimSuggestion from './ClaimSuggestion.vue'

interface DBCharacter extends TibiaCharacter {
  id: string
  user_id: string
}

const router = useRouter()
const userStore = useUserStore()
const characterName = ref('')
const error = ref('')
const loading = ref(false)
const existingCharacters = ref<DBCharacter[]>([])
const filteredCharacters = ref<DBCharacter[]>([])
const isLoadingCharacters = ref(false)
const showDropdown = ref(false)
const selectedCharacter = ref<DBCharacter | null>(null)
const showNameConflict = ref(false)

onMounted(async () => {
  if (userStore.isAuthenticated) {
    try {
      isLoadingCharacters.value = true
      const response = await axios.get(`/api/users/${userStore.userId}/characters`)
      existingCharacters.value = response.data
    } catch (e) {
      console.error('Failed to fetch characters:', e)
    } finally {
      isLoadingCharacters.value = false
    }
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

const handleInput = (event: Event) => {
  const input = (event.target as HTMLInputElement).value
  characterName.value = input
  selectedCharacter.value = null
  filterCharacters(input)
  showDropdown.value = true
}

const handleFocus = () => {
  showDropdown.value = true
  filterCharacters(characterName.value)
}

const handleBlur = () => {
  setTimeout(() => {
    showDropdown.value = false
  }, 200)
}

const selectCharacter = (character: DBCharacter) => {
  characterName.value = character.name
  selectedCharacter.value = character
  showDropdown.value = false
}

const verifyCharacter = async () => {
  error.value = ''
  loading.value = true

  try {
    // If we selected an existing character, use that
    if (selectedCharacter.value) {
      router.push({
        name: 'create-list',
        query: {
          character: JSON.stringify(selectedCharacter.value),
          useExisting: 'true',
        },
      })
      return
    }

    // Otherwise verify and create a new character
    const character = await tibiaDataService.getCharacter(characterName.value)
    router.push({
      name: 'create-list',
      query: { character: JSON.stringify(character) },
    })
  } catch (e) {
    if (axios.isAxiosError(e) && e.response?.status === 409) {
      showNameConflict.value = true
    } else {
      error.value = e instanceof Error ? e.message : 'Failed to verify character'
    }
  } finally {
    loading.value = false
  }
}

const handleTryDifferent = () => {
  showNameConflict.value = false
  characterName.value = ''
}
</script>

<template>
  <div class="p-6 rounded-lg">
    <h2 class="mb-4 text-2xl">Create a list</h2>
    <ClaimSuggestion 
      v-if="showNameConflict"
      :character-name="characterName"
      @try-different="handleTryDifferent"
    />
    <form v-else @submit.prevent="verifyCharacter" class="flex flex-col gap-4">
      <div class="relative">
        <input
          v-model="characterName"
          type="text"
          :placeholder="
            userStore.isAuthenticated
              ? 'Enter new character or select existing'
              : 'Enter character name'
          "
          required
          :disabled="loading"
          @input="handleInput"
          @focus="handleFocus"
          @blur="handleBlur"
          class="w-full p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
        />
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
      <div v-if="isLoadingCharacters" class="text-sm text-gray-500">Loading your characters...</div>
      <div v-if="error" class="text-red-500 text-sm">{{ error }}</div>
      <button
        type="submit"
        :disabled="loading"
        class="px-4 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 disabled:bg-gray-400"
      >
        {{ loading ? 'Verifying...' : 'Create' }}
      </button>
    </form>
  </div>
</template>
