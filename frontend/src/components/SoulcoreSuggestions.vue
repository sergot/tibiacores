<!-- A component to display and manage soulcore suggestions for a character -->
<template>
  <div v-if="suggestions.length > 0" class="bg-white shadow rounded-lg p-4 mt-4">
    <h3 class="text-lg font-semibold mb-3">Soul Core Suggestions</h3>
    <div class="space-y-3">
      <div v-for="suggestion in suggestions" :key="`${suggestion.character_id}-${suggestion.creature_id}`" class="flex items-center justify-between p-2 bg-gray-50 rounded">
        <div>
          <span class="font-medium">{{ suggestion.creature_name }}</span>
          <span class="text-sm text-gray-500 ml-2">from list "{{ getListName(suggestion.list_id) }}"</span>
        </div>
        <div class="flex space-x-2">
          <button 
            @click="acceptSuggestion(suggestion)"
            class="px-3 py-1 bg-green-500 text-white text-sm rounded hover:bg-green-600"
          >
            Accept
          </button>
          <button 
            @click="dismissSuggestion(suggestion)"
            class="px-3 py-1 bg-gray-300 text-gray-700 text-sm rounded hover:bg-gray-400"
          >
            Dismiss
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'

const props = defineProps<{
  characterId: string
}>()

const emit = defineEmits<{
  (e: 'suggestion-accepted'): void
}>()

interface Suggestion {
  character_id: string
  creature_id: string
  list_id: string
  suggested_at: string
  creature_name: string
}

const suggestions = ref<Suggestion[]>([])
const lists = ref<Record<string, { name: string }>>({})

const loadSuggestions = async () => {
  try {
    const response = await axios.get(`/api/characters/${props.characterId}/suggestions`)
    suggestions.value = response.data
    
    // Load list names for all unique list IDs
    const listIds = [...new Set(suggestions.value.map(s => s.list_id))]
    await Promise.all(listIds.map(async (listId) => {
      try {
        const listResponse = await axios.get(`/api/lists/${listId}`)
        lists.value[listId] = { name: listResponse.data.name }
      } catch (error) {
        console.error('Failed to load list details:', error)
      }
    }))
  } catch (error) {
    console.error('Failed to load suggestions:', error)
  }
}

const getListName = (listId: string): string => {
  return lists.value[listId]?.name || 'Unknown List'
}

const acceptSuggestion = async (suggestion: Suggestion) => {
  try {
    await axios.post(`/api/characters/${props.characterId}/suggestions/accept`, {
      creature_id: suggestion.creature_id
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
    await axios.post(`/api/characters/${props.characterId}/suggestions/dismiss`, {
      creature_id: suggestion.creature_id
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