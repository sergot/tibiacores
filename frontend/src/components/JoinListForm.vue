<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import { tibiaDataService } from '../services/tibiadata'
import ClaimSuggestion from './ClaimSuggestion.vue'
import type { Character } from '../services/tibiadata'

interface DBCharacter extends Character {
  id: string
  user_id: string
}

const router = useRouter()
const { t } = useI18n()
const shareCode = ref('')
const characterName = ref('')
const error = ref('')
const loading = ref(false)
const showNameConflict = ref(false)
const selectedCharacter = ref<DBCharacter | null>(null)

const handleSubmit = async () => {
  loading.value = true
  error.value = ''

  try {
    // First verify character exists in Tibia and get their world
    const tibiaChar = await tibiaDataService.getCharacter(characterName.value)

    // Now join the list with the verified character info
    const response = await axios.post(`/lists/join/${shareCode.value}`, {
      character_id: selectedCharacter.value?.id,
      character_name: characterName.value,
      world: tibiaChar.world,
    })
    router.push({
      name: 'list-detail',
      params: { id: response.data.id }
    })
  } catch (err: unknown) {
    if (axios.isAxiosError(err) && err.response?.status === 409) {
      showNameConflict.value = true
    } else if (err instanceof Error && err.message === 'Character not found') {
      error.value = t('joinList.form.errors.characterNotFound')
    } else {
      error.value = axios.isAxiosError(err) ? err.response?.data?.message || t('joinList.form.errors.joinFailed') : t('joinList.form.errors.joinFailed')
    }
  } finally {
    loading.value = false
  }
}

const handleTryDifferent = () => {
  showNameConflict.value = false
  characterName.value = ''
  selectedCharacter.value = null
}
</script>

<template>
  <div class="p-6 rounded-lg">
    <h2 class="mb-4 text-2xl">{{ t('joinList.title') }}</h2>
    <div v-if="error" class="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <ClaimSuggestion
      v-if="showNameConflict"
      :character-name="characterName"
      @try-different="handleTryDifferent"
    />

    <form v-else @submit.prevent="handleSubmit" class="space-y-4">
      <div>
        <label for="shareCode" class="block text-sm font-medium text-gray-700 mb-1">
          {{ t('joinList.form.shareCode') }}
        </label>
        <input
          id="shareCode"
          v-model="shareCode"
          type="text"
          :placeholder="t('joinList.form.shareCodePlaceholder')"
          required
          :disabled="loading"
          class="w-full p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
        />
      </div>

      <div>
        <label for="characterName" class="block text-sm font-medium text-gray-700 mb-1">
          {{ t('joinList.form.characterName') }}
        </label>
        <input
          id="characterName"
          v-model="characterName"
          type="text"
          :placeholder="t('joinList.form.characterNamePlaceholder')"
          required
          :disabled="loading"
          class="w-full p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
        />
      </div>

      <button
        type="submit"
        :disabled="loading || !shareCode"
        class="w-full px-4 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 disabled:bg-gray-400"
      >
        {{ loading ? t('joinList.form.joining') : t('joinList.form.submit') }}
      </button>
    </form>
  </div>
</template>
