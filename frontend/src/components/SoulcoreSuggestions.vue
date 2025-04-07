<!-- A component to display and manage soulcore suggestions for a character -->
<template>
  <div class="bg-white shadow rounded-lg p-4 mt-4 overflow-visible">
    <h3 class="text-lg font-semibold mb-3">{{ t('soulcoreSuggestions.title') }}</h3>
    <div v-if="suggestions.length === 0" class="text-gray-500 text-center py-4">
      {{ t('soulcoreSuggestions.noSuggestions') }}
    </div>
    <div v-else class="space-y-3 overflow-visible">
      <div
        v-for="suggestion in suggestions"
        :key="`${suggestion.character_id}-${suggestion.creature_id}`"
        class="flex items-center justify-between p-2 bg-gray-50 rounded overflow-visible"
      >
        <div class="overflow-visible">
          <div class="mb-1">
            <span class="font-medium">{{ suggestion.creature_name }}</span>
          </div>
          <div class="flex items-center overflow-visible">
            <span class="text-sm text-gray-500">{{ t('soulcoreSuggestions.fromList') }}:</span>
            <span class="ml-1"><ListTag :list="lists[suggestion.list_id]" /></span>
          </div>
        </div>
        <div class="flex space-x-2">
          <button
            @click="acceptSuggestion(suggestion)"
            class="px-3 py-1 bg-green-500 text-white text-sm rounded hover:bg-green-600"
          >
            {{ t('soulcoreSuggestions.accept') }}
          </button>
          <button
            @click="dismissSuggestion(suggestion)"
            class="px-3 py-1 bg-gray-300 text-gray-700 text-sm rounded hover:bg-gray-400"
          >
            {{ t('soulcoreSuggestions.dismiss') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import ListTag from './ListTag.vue'

const props = defineProps<{
  characterId: string
}>()

const emit = defineEmits<{
  (e: 'suggestion-accepted'): void
}>()

const { t } = useI18n()

interface Suggestion {
  character_id: string
  creature_id: string
  list_id: string
  suggested_at: string
  creature_name: string
}

const suggestions = ref<Suggestion[]>([])
const lists = ref<Record<string, { id: string; name: string; member_count?: number }>>({})

const loadSuggestions = async () => {
  try {
    const response = await axios.get(`/characters/${props.characterId}/suggestions`)
    suggestions.value = response.data

    // Load list names for all unique list IDs
    const listIds = [...new Set(suggestions.value.map((s) => s.list_id))]
    await Promise.all(
      listIds.map(async (listId) => {
        try {
          const listResponse = await axios.get(`/lists/${listId}`)
          lists.value[listId] = {
            id: listId,
            name: listResponse.data.name,
            member_count: listResponse.data.member_count,
          }
        } catch (error) {
          console.error('Failed to load list details:', error)
        }
      }),
    )
  } catch (error) {
    console.error('Failed to load suggestions:', error)
  }
}

const acceptSuggestion = async (suggestion: Suggestion) => {
  try {
    await axios.post(`/characters/${props.characterId}/suggestions/accept`, {
      creature_id: suggestion.creature_id,
    })
    emit('suggestion-accepted')
    // Refresh suggestions after accepting one
    await loadSuggestions()
  } catch (error) {
    console.error('Failed to accept suggestion:', error)
  }
}

const dismissSuggestion = async (suggestion: Suggestion) => {
  try {
    await axios.post(`/characters/${props.characterId}/suggestions/dismiss`, {
      creature_id: suggestion.creature_id,
    })
    // Refresh suggestions after dismissing one
    await loadSuggestions()
  } catch (error) {
    console.error('Failed to dismiss suggestion:', error)
  }
}

onMounted(() => {
  loadSuggestions()
})
</script>
