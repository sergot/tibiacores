<script setup lang="ts">
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="px-4 py-5 sm:px-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900">Profile</h3>
        <p class="mt-1 max-w-2xl text-sm text-gray-500">Your personal information and characters</p>
      </div>
      <div class="border-t border-gray-200 px-4 py-5 sm:p-0">
        <dl class="sm:divide-y sm:divide-gray-200">
          <div class="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt class="text-sm font-medium text-gray-500">User ID</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{{ userStore.userId }}</dd>
          </div>
          <!-- Characters section - to be implemented -->
          <div class="py-4 sm:py-5 sm:px-6">
            <dt class="text-sm font-medium text-gray-500 mb-4">Characters</dt>
            <dd class="mt-1 text-sm text-gray-900">
              <div class="text-gray-500 italic">No characters added yet</div>
            </dd>
          </div>
        </dl>
      </div>
    </div>
  </div>
</template>