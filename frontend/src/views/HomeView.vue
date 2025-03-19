<script setup lang="ts">
import { onMounted } from 'vue'
import CreateListForm from '../components/CreateListForm.vue'
import JoinListForm from '../components/JoinListForm.vue'
import { useUserStore } from '../stores/user'
import { useListsStore } from '../stores/lists'

const userStore = useUserStore()
const listsStore = useListsStore()

onMounted(() => {
  if (userStore.isAuthenticated) {
    listsStore.fetchUserLists()
  }
})
</script>

<template>
  <div class="w-full">
    <main class="max-w-6xl mx-auto px-8 py-8">
      <div class="text-center mb-8">
        <h1 class="text-4xl mb-2">Welcome to Fiendlist</h1>
        <p class="text-xl text-gray-600">
          Your companion app for managing and sharing gaming lists
        </p>
      </div>

      <div class="grid md:grid-cols-2 gap-8 mb-8">
        <div class="p-6 rounded-lg shadow-sm bg-white">
          <CreateListForm />
        </div>
        <div class="p-6 rounded-lg shadow-sm bg-white">
          <JoinListForm />
        </div>
      </div>

      <div v-if="userStore.isAuthenticated" class="mb-8 p-8 rounded-xl bg-white shadow-md">
        <div class="flex items-center justify-between mb-8">
          <h2 class="text-3xl font-semibold text-gray-800">Your Lists</h2>
          <button
            @click="listsStore.fetchUserLists()"
            class="px-4 py-2 text-sm font-medium text-gray-600 hover:text-gray-800 rounded-lg hover:bg-gray-100 transition-all duration-200 flex items-center gap-2"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            Refresh
          </button>
        </div>

        <div v-if="listsStore.isLoading" class="text-center py-12">
          <div
            class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full mx-auto mb-4"
          ></div>
          <p class="text-gray-600 font-medium">Loading your lists...</p>
        </div>

        <div v-else-if="listsStore.error" class="text-center py-12">
          <div class="bg-red-50 rounded-lg p-6 max-w-md mx-auto">
            <p class="text-red-600 font-medium mb-4">{{ listsStore.error }}</p>
            <button
              @click="listsStore.fetchUserLists()"
              class="px-6 py-2.5 text-sm font-medium bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors"
            >
              Try again
            </button>
          </div>
        </div>

        <div v-else-if="!listsStore.hasLists" class="text-center py-12">
          <div class="bg-gray-50 rounded-lg p-8 max-w-md mx-auto">
            <p class="text-gray-700 font-medium mb-2">
              You haven't created or joined any lists yet.
            </p>
            <p class="text-gray-500">Use the forms above to get started!</p>
          </div>
        </div>

        <div v-else class="grid gap-6">
          <div
            v-for="list in listsStore.lists"
            :key="list.id"
            class="p-6 border border-gray-200 rounded-xl hover:bg-gray-50 transition-all duration-200 hover:shadow-lg group"
          >
            <div class="flex justify-between items-start">
              <div class="space-y-2">
                <h3 class="font-semibold text-xl text-gray-800 group-hover:text-gray-900">
                  {{ list.name }}
                </h3>
                <div class="space-y-1">
                  <p class="text-sm text-gray-600 flex items-center gap-2">
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      class="h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"
                      />
                    </svg>
                    World: {{ list.world }}
                  </p>
                  <p
                    v-if="list.character_name"
                    class="text-sm text-gray-600 flex items-center gap-2"
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      class="h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                      />
                    </svg>
                    {{ list.is_author ? 'Owner Character' : 'Member Character' }}:
                    {{ list.character_name }}
                  </p>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <span
                  v-if="list.is_author"
                  class="px-3 py-1.5 text-xs font-semibold bg-blue-100 text-blue-800 rounded-lg border border-blue-200"
                >
                  Owner
                </span>
                <span
                  v-else
                  class="px-3 py-1.5 text-xs font-semibold bg-gray-100 text-gray-700 rounded-lg border border-gray-200"
                >
                  Member
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="mt-8 p-6 rounded-lg bg-white text-center">
        <h2 class="mb-4 text-2xl">About Fiendlist</h2>
        <p class="text-gray-600 leading-relaxed">
          Fiendlist is your companion app for managing and sharing gaming lists. Create characters,
          build your collections, and share them with friends. Join existing lists using share codes
          or direct links.
        </p>
      </div>
    </main>
  </div>
</template>

<style scoped></style>
