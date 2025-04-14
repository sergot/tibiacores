<template>
  <div class="w-full bg-gradient-to-b from-gray-50 to-gray-100 min-h-[calc(100vh-8rem)]">
    <main class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <div class="text-center mb-8">
        <h1
          class="text-5xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-indigo-600 to-blue-600 mb-2"
        >
          {{ t('highscores.title') }}
        </h1>
        <p class="text-xl text-gray-600">{{ t('highscores.subtitle') }}</p>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center items-center py-12">
        <svg
          class="animate-spin h-10 w-10 text-indigo-600"
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
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="text-center py-12">
        <div class="rounded-lg bg-red-50 p-6 border border-red-200 max-w-2xl mx-auto">
          <svg
            class="h-12 w-12 text-red-500 mx-auto mb-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
          <h3 class="text-lg font-medium text-red-800 mb-2">{{ t('highscores.error.title') }}</h3>
          <p class="text-red-700">{{ t('highscores.error.description') }}</p>
          <button
            @click="fetchHighscores"
            class="mt-4 bg-red-500 text-white px-4 py-2 rounded-md hover:bg-red-600 transition-colors"
          >
            {{ t('highscores.error.retry') }}
          </button>
        </div>
      </div>

      <!-- Data Display -->
      <div v-else>
        <!-- Highscores Table -->
        <div class="bg-white shadow-sm rounded-lg overflow-hidden border border-gray-200 mb-6">
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th
                    scope="col"
                    class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  >
                    {{ t('highscores.columns.rank') }}
                  </th>
                  <th
                    scope="col"
                    class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  >
                    {{ t('highscores.columns.name') }}
                  </th>
                  <th
                    scope="col"
                    class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  >
                    {{ t('highscores.columns.world') }}
                  </th>
                  <th
                    scope="col"
                    class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  >
                    {{ t('highscores.columns.cores') }}
                  </th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
                <tr
                  v-for="(character, index) in characters"
                  :key="character.id"
                  class="hover:bg-gray-50 transition-colors duration-200"
                >
                  <td class="px-6 py-4 whitespace-nowrap">
                    <div class="text-sm font-semibold text-gray-900">
                      {{ getCharacterRank(index) }}
                    </div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <div class="text-sm font-medium text-blue-600 hover:text-blue-800">
                      <RouterLink :to="`/characters/public/${character.name}`">{{
                        character.name
                      }}</RouterLink>
                    </div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <div class="text-sm text-gray-900">{{ character.world }}</div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <div class="text-sm text-gray-900">{{ character.core_count }}</div>
                  </td>
                </tr>
                <tr v-if="characters.length === 0">
                  <td colspan="4" class="px-6 py-8 text-center text-gray-500">
                    {{ t('highscores.empty') }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Pagination -->
        <div class="flex justify-between items-center">
          <div class="text-sm text-gray-700">
            {{
              characters.length > 0
                ? t('highscores.pagination.showing', {
                    from: (pagination.currentPage - 1) * pagination.pageSize + 1,
                    to: (pagination.currentPage - 1) * pagination.pageSize + characters.length,
                    total: pagination.totalRecords || characters.length,
                  })
                : t('highscores.empty')
            }}
          </div>
          <div class="flex space-x-1" v-if="pagination.totalPages > 0">
            <button
              @click="goToPage(1)"
              :disabled="pagination.currentPage === 1"
              class="px-3 py-1 border rounded-md text-sm font-medium"
              :class="
                pagination.currentPage === 1
                  ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                  : 'bg-white text-indigo-600 hover:bg-indigo-50'
              "
            >
              &laquo;
            </button>
            <button
              @click="goToPage(pagination.currentPage - 1)"
              :disabled="pagination.currentPage === 1"
              class="px-3 py-1 border rounded-md text-sm font-medium"
              :class="
                pagination.currentPage === 1
                  ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                  : 'bg-white text-indigo-600 hover:bg-indigo-50'
              "
            >
              &lsaquo;
            </button>

            <template v-for="pageNum in displayedPageNumbers" :key="pageNum">
              <div v-if="pageNum === '...'" class="px-3 py-1 text-sm font-medium text-gray-700">
                ...
              </div>
              <button
                v-else
                @click="goToPage(pageNum)"
                class="px-3 py-1 border rounded-md text-sm font-medium"
                :class="
                  pagination.currentPage === pageNum
                    ? 'bg-indigo-600 text-white'
                    : 'bg-white text-indigo-600 hover:bg-indigo-50'
                "
              >
                {{ pageNum }}
              </button>
            </template>

            <button
              @click="goToPage(pagination.currentPage + 1)"
              :disabled="pagination.currentPage === pagination.totalPages"
              class="px-3 py-1 border rounded-md text-sm font-medium"
              :class="
                pagination.currentPage === pagination.totalPages
                  ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                  : 'bg-white text-indigo-600 hover:bg-indigo-50'
              "
            >
              &rsaquo;
            </button>
            <button
              @click="goToPage(pagination.totalPages)"
              :disabled="pagination.currentPage === pagination.totalPages"
              class="px-3 py-1 border rounded-md text-sm font-medium"
              :class="
                pagination.currentPage === pagination.totalPages
                  ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                  : 'bg-white text-indigo-600 hover:bg-indigo-50'
              "
            >
              &raquo;
            </button>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

interface Character {
  id: string
  name: string
  world: string
  core_count: number
}

interface Pagination {
  totalPages: number
  currentPage: number
  totalRecords: number
  pageSize: number
}

const characters = ref<Character[]>([])
const pagination = ref<Pagination>({
  totalPages: 0,
  currentPage: 1,
  totalRecords: 0,
  pageSize: 20,
})
const loading = ref(true)
const error = ref('')

// Calculate displayed page numbers for pagination
const displayedPageNumbers = computed(() => {
  const current = pagination.value.currentPage || 1
  const total = pagination.value.totalPages || 1

  // If no pages or only one page, return empty array or just page 1
  if (total <= 0) {
    return []
  }

  // If fewer than 8 pages, show all
  if (total <= 8) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }

  // Otherwise, show a window around current page
  const pages: (number | string)[] = [1]

  if (current > 3) pages.push('...')

  // Calculate start and end of the window
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)

  for (let i = start; i <= end; i++) {
    pages.push(i)
  }

  if (current < total - 2) pages.push('...')

  if (total > 1) pages.push(total)

  return pages
})

// Calculate the rank of a character based on their position in the page
const getCharacterRank = (index: number) => {
  if (!pagination.value.currentPage || !pagination.value.pageSize) {
    return index + 1 // Fallback to index if pagination data is not available
  }
  return (pagination.value.currentPage - 1) * pagination.value.pageSize + index + 1
}

// Navigate to a specific page
const goToPage = (pageNum: number | string) => {
  if (typeof pageNum === 'string') return // Skip ellipsis
  if (pageNum === pagination.value.currentPage) return

  router.push({
    path: '/highscores',
    query: { page: pageNum.toString() },
  })
}

// Fetch highscores from the API
const fetchHighscores = async () => {
  const pageParam = (route.query.page as string) || '1'
  const page = parseInt(pageParam, 10) || 1

  loading.value = true
  error.value = ''

  try {
    const response = await axios.get('/highscores', {
      params: { page },
    })

    characters.value = response.data.characters || []

    // Set default pagination values if not provided by the backend
    if (response.data.pagination) {
      pagination.value = {
        totalPages: response.data.pagination.totalPages || 1,
        currentPage: response.data.pagination.currentPage || page,
        totalRecords: response.data.pagination.totalRecords || 0,
        pageSize: response.data.pagination.pageSize || 20,
      }
    }
  } catch (err) {
    console.error('Failed to load highscores:', err)
    error.value = t('highscores.error.description')
  } finally {
    loading.value = false
  }
}

// Watch for route changes and refetch data if needed
onMounted(() => {
  fetchHighscores()
})

// Watch for page changes in the route and refetch data
router.beforeResolve((to, from) => {
  if (to.path === '/highscores' && to.query.page !== from.query.page) {
    fetchHighscores()
  }
})
</script>
