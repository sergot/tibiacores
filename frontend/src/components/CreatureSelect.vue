<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'

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

interface UnlockStats {
  creature_id: string
  unlocked_count: number
  unlocked_by: Array<{
    character_name: string
  }>
}

const props = defineProps<{
  creatures: Creature[]
  modelValue: string
  existingSoulCores: SoulCore[]
  unlockStats?: Record<string, UnlockStats>
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const isOpen = ref(false)
const searchInput = ref(props.modelValue)
const isMobile = ref(false)

// Keep searchInput in sync with modelValue
watch(
  () => props.modelValue,
  (newValue) => {
    searchInput.value = newValue
  },
)

const selectedIndex = ref(-1)
const dropdownRef = ref<HTMLDivElement | null>(null)

const handleResize = () => {
  isMobile.value = window.innerWidth < 640
}

onMounted(() => {
  handleResize()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
})

const getUnlockInfo = (creature: Creature) => {
  return props.unlockStats?.[creature.id] || null
}

const filteredCreatures = computed(() => {
  const query = searchInput.value.toLowerCase()
  return props.creatures
    .filter((creature) => creature.name.toLowerCase().includes(query))
    .map((creature) => ({
      ...creature,
      isObtained: props.existingSoulCores.some(
        (sc) =>
          sc.creature_id === creature.id && (sc.status === 'obtained' || sc.status === 'unlocked'),
      ),
      unlockStats: getUnlockInfo(creature),
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

const handleKeydown = (event: KeyboardEvent) => {
  if (!isOpen.value || filteredCreatures.value.length === 0) return

  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      selectedIndex.value = Math.min(selectedIndex.value + 1, filteredCreatures.value.length - 1)
      break
    case 'ArrowUp':
      event.preventDefault()
      selectedIndex.value = Math.max(selectedIndex.value - 1, -1)
      break
    case 'Enter':
      event.preventDefault()
      if (selectedIndex.value >= 0 && selectedIndex.value < filteredCreatures.value.length) {
        const selectedCreature = filteredCreatures.value[selectedIndex.value]
        if (!selectedCreature.isObtained) {
          searchInput.value = selectedCreature.name
          emit('update:modelValue', selectedCreature.name)
          isOpen.value = false
          selectedIndex.value = -1
        }
      }
      break
    case 'Escape':
      isOpen.value = false
      selectedIndex.value = -1
      break
  }
}

const handleClickOutside = (event: MouseEvent) => {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    isOpen.value = false
    selectedIndex.value = -1
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
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
      class="w-full min-w-[20rem] max-w-2xl p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
    />

    <!-- Enhanced Dropdown -->
    <div
      v-if="isOpen && filteredCreatures.length > 0"
      class="fixed z-[100] left-0 top-0 w-full h-0 pointer-events-none"
    >
      <div
        class="absolute z-50 w-full mt-1 bg-white border border-gray-200 rounded-lg shadow-lg pointer-events-auto"
        :style="{
          left: `${$el?.getBoundingClientRect().left}px`,
          top: `${$el?.getBoundingClientRect().bottom + 4}px`,
          width: `${$el?.getBoundingClientRect().width}px`,
        }"
      >
        <div class="max-h-64 overflow-y-auto">
          <div
            v-for="(creature, index) in filteredCreatures"
            :key="creature.id"
            @click="
              !creature.isObtained &&
              ((searchInput = creature.name),
              emit('update:modelValue', creature.name),
              (isOpen = false),
              (selectedIndex = -1))
            "
            class="p-2 group relative"
            :class="{
              'cursor-pointer text-sm': true,
              'text-gray-700 hover:bg-gray-50': !creature.isObtained,
              'text-gray-400 cursor-not-allowed': creature.isObtained,
              'bg-indigo-50': index === selectedIndex && !creature.isObtained,
            }"
          >
            <div class="flex items-center justify-between gap-2">
              <span>{{ creature.name }}</span>
              <div class="flex items-center gap-2">
                <div v-if="creature.unlockStats?.unlocked_count" class="group/tooltip relative">
                  <span
                    class="px-2 py-0.5 text-xs font-medium rounded-full"
                    :class="{
                      'bg-purple-100 text-purple-800': !creature.isObtained,
                      'bg-gray-100 text-gray-600': creature.isObtained,
                    }"
                  >
                    {{ creature.unlockStats.unlocked_count }} unlocked
                  </span>
                  <div class="fixed z-[100] left-0 top-0 w-full h-0 pointer-events-none">
                    <div
                      class="invisible group-hover/tooltip:visible group-active/tooltip:visible opacity-0 group-hover/tooltip:opacity-100 group-active/tooltip:opacity-100 transition-opacity absolute z-50 w-48 p-2 bg-white rounded-lg shadow-lg border border-gray-200 pointer-events-auto"
                      :style="{
                        left: isMobile ? '1rem' : `${$el?.getBoundingClientRect().right + 8}px`,
                        top: isMobile
                          ? `${$el?.getBoundingClientRect().bottom + 8}px`
                          : `${$el?.getBoundingClientRect().top}px`,
                        maxWidth: 'calc(100vw - 2rem)',
                      }"
                    >
                      <div class="font-medium text-gray-700 border-b border-gray-100 pb-1 mb-1">
                        Characters who unlocked
                      </div>
                      <div
                        v-for="member in creature.unlockStats.unlocked_by"
                        :key="member.character_name"
                        class="text-sm py-1"
                      >
                        <span class="text-gray-600">{{ member.character_name }}</span>
                      </div>
                    </div>
                  </div>
                </div>
                <span v-if="creature.isObtained" class="text-xs text-gray-500"
                  >(already obtained)</span
                >
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
