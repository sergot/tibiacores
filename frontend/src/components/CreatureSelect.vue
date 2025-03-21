<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'

interface Creature {
  id: string
  name: string
}

interface SoulCore {
  creature_id: string
  creature_name: string
  status: 'obtained' | 'unlocked'
  added_by: string | null
  added_by_user_id: string
}

const props = defineProps<{
  creatures: Creature[]
  modelValue: string
  existingSoulCores: SoulCore[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const isOpen = ref(false)
const searchInput = ref(props.modelValue)

// Keep searchInput in sync with modelValue
watch(() => props.modelValue, (newValue) => {
  searchInput.value = newValue
})

const selectedIndex = ref(-1)
const dropdownRef = ref<HTMLDivElement | null>(null)

const filteredCreatures = computed(() => {
  const query = searchInput.value.toLowerCase()
  return props.creatures
    .filter(creature => creature.name.toLowerCase().includes(query))
    .map(creature => ({
      ...creature,
      isObtained: props.existingSoulCores.some(sc => 
        sc.creature_id === creature.id && 
        (sc.status === 'obtained' || sc.status === 'unlocked')
      )
    }))
    .sort((a, b) => {
      // Sort obtained creatures to the bottom
      if (a.isObtained === b.isObtained) {
        return a.name.localeCompare(b.name)
      }
      return a.isObtained ? 1 : -1
    })
    .slice(0, 10) // Limit to 10 results
})

const selectCreature = (creature: Creature) => {
  // Don't select if already obtained
  if (props.existingSoulCores.some(sc => 
    sc.creature_id === creature.id && 
    (sc.status === 'obtained' || sc.status === 'unlocked')
  )) {
    return
  }

  searchInput.value = creature.name
  emit('update:modelValue', creature.name)
  isOpen.value = false
  selectedIndex.value = -1
}

const handleKeydown = (e: KeyboardEvent) => {
  if (!isOpen.value) return

  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault()
      selectedIndex.value = Math.min(selectedIndex.value + 1, filteredCreatures.value.length - 1)
      break
    case 'ArrowUp':
      e.preventDefault()
      selectedIndex.value = Math.max(selectedIndex.value - 1, -1)
      break
    case 'Enter':
      e.preventDefault()
      if (selectedIndex.value >= 0) {
        const creature = filteredCreatures.value[selectedIndex.value]
        if (!creature.isObtained) {
          selectCreature(creature)
        }
      }
      break
    case 'Escape':
      isOpen.value = false
      selectedIndex.value = -1
      break
  }
}

const handleClickOutside = (e: MouseEvent) => {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    isOpen.value = false
    selectedIndex.value = -1
  }
}

onMounted(() => {
  document.addEventListener('mousedown', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', handleClickOutside)
})
</script>

<template>
  <div class="relative" ref="dropdownRef">
    <input
      v-model="searchInput"
      type="text"
      placeholder="Select creature..."
      @focus="isOpen = true"
      @keydown="handleKeydown"
      class="w-64 p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
    />
    <div
      v-if="isOpen && filteredCreatures.length > 0"
      class="absolute z-10 w-full mt-1 bg-white border border-gray-200 rounded-lg shadow-lg overflow-y-auto"
      style="max-height: 12rem;"
    >
      <div
        v-for="(creature, index) in filteredCreatures"
        :key="creature.id"
        @click="selectCreature(creature)"
        :class="{
          'p-2 cursor-pointer text-sm': true,
          'text-gray-700 hover:bg-gray-50': !creature.isObtained,
          'text-gray-400 cursor-not-allowed': creature.isObtained,
          'bg-indigo-50': index === selectedIndex && !creature.isObtained
        }"
      >
        {{ creature.name }}
        <span v-if="creature.isObtained" class="ml-2 text-xs">(already obtained)</span>
      </div>
    </div>
  </div>
</template>