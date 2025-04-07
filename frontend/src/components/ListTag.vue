<template>
  <div class="relative group" @mouseenter="fetchListMembers">
    <!-- Hover area that includes both the tag and tooltip -->
    <div class="absolute -inset-2 z-40"></div>

    <div
      class="flex items-center space-x-1 text-gray-600 hover:text-gray-900 cursor-pointer relative z-50"
    >
      <ListBulletIcon class="h-4 w-4" />
      <span>{{ list?.name || t('soulcoreSuggestions.unknownList') }}</span>
    </div>

    <!-- Tooltip -->
    <div
      v-if="list"
      class="absolute left-0 -top-2 -translate-y-full hidden group-hover:block z-50 w-72 tooltip-shadow"
    >
      <div class="bg-white rounded-lg shadow-lg border border-gray-200 overflow-hidden">
        <!-- Header -->
        <div class="bg-gradient-to-r from-indigo-50 to-blue-50 p-3 border-b border-gray-200">
          <h4 class="font-medium text-gray-900">{{ list.name }}</h4>
          <p class="text-sm text-gray-600 mt-1">
            {{ t('listTag.members', { count: members.length || 0 }) }}
          </p>
        </div>

        <!-- Members list -->
        <div class="p-3">
          <div
            v-if="loadingMembers"
            class="py-2 text-center text-sm text-gray-500 flex items-center justify-center space-x-2"
          >
            <div
              class="w-4 h-4 border-2 border-gray-300 border-t-indigo-500 rounded-full animate-spin"
            ></div>
            <span>{{ t('listTag.loadingMembers') }}</span>
          </div>
          <div v-else-if="members.length > 0" class="max-h-32 overflow-y-auto mb-2">
            <div class="text-sm font-medium text-gray-700 mb-2 flex items-center">
              <UsersIcon class="h-4 w-4 mr-1 text-gray-500" />
              {{ t('listTag.membersList') }}
            </div>
            <ul class="text-sm text-gray-600 space-y-1 pl-2 divide-y divide-gray-100">
              <li v-for="member in members" :key="member.creature_id" class="truncate py-1">
                {{ member.character_name }}
              </li>
            </ul>
          </div>

          <RouterLink
            :to="`/lists/${list.id}`"
            class="mt-2 inline-flex items-center text-sm text-blue-600 hover:text-blue-800 hover:underline"
          >
            <ArrowTopRightOnSquareIcon class="h-4 w-4 mr-1" />
            {{ t('listTag.goToList') }}
          </RouterLink>
        </div>
      </div>
      <!-- Tooltip arrow -->
      <div
        class="absolute left-5 bottom-0 w-3 h-3 bg-white transform rotate-45 border-r border-b border-gray-200 tooltip-arrow"
      ></div>
    </div>
  </div>
</template>

<style scoped>
.group:hover .group-hover\:block {
  display: block !important;
}

.group .group-hover\:block:hover {
  display: block !important;
}

.tooltip-shadow {
  box-shadow:
    0 10px 15px -3px rgba(0, 0, 0, 0.1),
    0 4px 6px -2px rgba(0, 0, 0, 0.05);
}

.tooltip-arrow {
  box-shadow: 2px 2px 2px rgba(0, 0, 0, 0.02);
}
</style>

<script setup lang="ts">
import { ref } from 'vue'
import { ListBulletIcon, UsersIcon, ArrowTopRightOnSquareIcon } from '@heroicons/vue/24/outline'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
import axios from 'axios'

const { t } = useI18n()

const props = defineProps<{
  list?: {
    id: string
    name: string
    member_count?: number
  }
}>()

interface ListMember {
  creature_id: string
  character_name: string
}

const members = ref<ListMember[]>([])
const loadingMembers = ref(false)
const membersLoaded = ref(false)

const fetchListMembers = async () => {
  // Skip if no list or already loaded
  if (!props.list || membersLoaded.value) return

  loadingMembers.value = true
  try {
    const response = await axios.get(`/lists/${props.list.id}`)
    // Extract members from the response
    if (response.data.members && Array.isArray(response.data.members)) {
      members.value = response.data.members
    }
    membersLoaded.value = true
  } catch (error) {
    console.error('Failed to load list members:', error)
  } finally {
    loadingMembers.value = false
  }
}
</script>
