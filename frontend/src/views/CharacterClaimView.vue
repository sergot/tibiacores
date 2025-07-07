<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useI18n } from 'vue-i18n'
import axios from 'axios'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()

interface ClaimResponse {
  claim_id: string
  verification_code: string
  status: string
  token?: string
  claimer_id?: string
}

interface ApiError {
  response?: {
    data?: {
      message?: string
    }
    status?: number
  }
}

const characterName = ref('')
const claim = ref<ClaimResponse | null>(null)
const error = ref('')
const loading = ref(false)
const lastCheckTime = ref(0)
const claimId = ref<string | null>(null)

onMounted(() => {
  // Get claim ID from the URL if it exists
  claimId.value = route.query.claim_id as string

  // If we have a claim ID, fetch the claim
  if (claimId.value) {
    checkClaim(claimId.value)
  } else {
    // Otherwise check if we should start a new claim
    const queryCharacter = route.query.character as string
    if (queryCharacter) {
      characterName.value = queryCharacter
      startClaim()
    }
  }
})

const startClaim = async () => {
  if (!characterName.value) return

  loading.value = true
  error.value = ''

  try {
    const response = await axios.post<ClaimResponse>('/claims', {
      character_name: characterName.value,
    })

    // handle anonymous users
    const authToken = response.headers['x-auth-token']
    if (authToken && response.data.claimer_id && !userStore.isAuthenticated) {
      userStore.setUser({
        session_token: authToken,
        id: response.data.claimer_id, // Backend provides claimer_id
        has_email: false,
      })

      // Set the token for future requests
      axios.defaults.headers.common['Authorization'] = `Bearer ${authToken}`
    }

    claim.value = response.data
  } catch (err: unknown) {
    error.value =
      err instanceof Error
        ? err.message
        : (err as ApiError).response?.data?.message || 'Failed to start claim'
  } finally {
    loading.value = false
  }
}

const checkClaim = async (id?: string) => {
  if (!id && !claim.value?.claim_id) return

  loading.value = true
  error.value = ''

  try {
    const response = await axios.get(`/claims/${id || claim.value?.claim_id}`)
    claim.value = response.data
    lastCheckTime.value = Date.now()

    // Update URL with claim ID if not already there
    if (!route.query.claim_id) {
      router.replace({
        query: {
          ...route.query,
          claim_id: response.data.claim_id,
        },
      })
    }

    if (response.data.status === 'approved') {
      setTimeout(() => {
        router.push('/')
      }, 2000)
    }
  } catch (err: unknown) {
    const apiError = err as ApiError
    if (apiError.response?.status === 403) {
      // If forbidden (claim doesn't belong to user), start a new claim
      claim.value = null
      claimId.value = null
      router.replace({ query: { character: characterName.value } })
      startClaim()
    } else {
      error.value =
        err instanceof Error
          ? err.message
          : apiError.response?.data?.message || 'Failed to check claim'
    }
  } finally {
    loading.value = false
  }
}

const resetClaim = () => {
  claim.value = null
  error.value = ''
  characterName.value = ''
  lastCheckTime.value = 0
}
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)] flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <main class="max-w-2xl w-full space-y-8">
      <!-- Error message -->
      <div v-if="error" class="rounded-md bg-red-50 p-4">
        <div class="flex">
          <div class="ml-3">
            <h3 class="text-sm font-medium text-red-800">{{ error }}</h3>
          </div>
        </div>
      </div>

      <!-- Start claim form -->
      <div v-if="!claim" class="bg-white rounded-lg shadow p-8">
        <div class="text-center mb-8">
          <h1 class="text-3xl font-bold text-gray-900">{{ t('characterClaim.title') }}</h1>
          <p class="mt-2 text-gray-600">
            {{ t('characterClaim.subtitle') }}
          </p>
        </div>

        <form @submit.prevent="startClaim" class="space-y-6">
          <div>
            <label for="characterName" class="block text-sm font-medium text-gray-700 mb-1">
              {{ t('characterClaim.characterName') }}
            </label>
            <input
              id="characterName"
              v-model="characterName"
              type="text"
              required
              :disabled="loading"
              :placeholder="t('characterClaim.characterNamePlaceholder')"
              class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-green-500 focus:border-green-500 sm:text-sm"
            />
          </div>

          <button
            type="submit"
            :disabled="loading || !characterName"
            class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            {{ loading ? t('characterClaim.starting') : t('characterClaim.startClaim') }}
          </button>
        </form>
      </div>

      <!-- Claim verification instructions -->
      <div v-else-if="claim.status === 'pending'" class="bg-white rounded-lg shadow p-8">
        <div class="text-center mb-8">
          <h1 class="text-3xl font-bold text-gray-900">{{ t('characterClaim.verifyTitle') }}</h1>
          <p class="mt-2 text-gray-600">
            {{ t('characterClaim.verifySubtitle') }}
          </p>
        </div>

        <div class="space-y-6">
          <ol class="list-decimal pl-5 space-y-4 text-gray-600">
            <li>{{ t('characterClaim.steps.1') }}</li>
            <li>{{ t('characterClaim.steps.2') }}</li>
            <li>{{ t('characterClaim.steps.3') }}</li>
            <li>{{ t('characterClaim.steps.4') }}</li>
          </ol>

          <div class="bg-gray-50 p-4 rounded-md">
            <p class="font-mono text-sm break-all text-center select-all">
              {{ claim.verification_code }}
            </p>
          </div>

          <div class="bg-blue-50 border border-blue-200 rounded-md p-4">
            <p class="text-sm text-blue-700">
              {{ t('characterClaim.waitNote') }}
            </p>
          </div>

          <button
            @click="() => checkClaim()"
            :disabled="loading || lastCheckTime > Date.now() - 60000"
            class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            {{ loading ? t('characterClaim.checking') : t('characterClaim.checkStatus') }}
          </button>

          <p v-if="lastCheckTime > Date.now() - 60000" class="text-sm text-gray-500 text-center">
            {{
              t('characterClaim.waitTime', {
                seconds: Math.ceil((60000 - (Date.now() - lastCheckTime)) / 1000),
              })
            }}
          </p>
        </div>
      </div>

      <!-- Claim success message -->
      <div v-else-if="claim.status === 'approved'" class="bg-white rounded-lg shadow p-8">
        <div class="text-center">
          <div
            class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 mb-4"
          >
            <svg
              class="h-6 w-6 text-green-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M5 13l4 4L19 7"
              />
            </svg>
          </div>
          <h2 class="text-2xl font-bold text-gray-900 mb-2">
            {{ t('characterClaim.success.title') }}
          </h2>
          <p class="text-gray-600 mb-6">
            {{ t('characterClaim.success.message') }}
          </p>
        </div>
      </div>

      <!-- Claim rejected message -->
      <div v-else-if="claim.status === 'rejected'" class="bg-white rounded-lg shadow p-8">
        <div class="text-center">
          <div
            class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 mb-4"
          >
            <svg class="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </div>
          <h2 class="text-2xl font-bold text-gray-900 mb-2">
            {{ t('characterClaim.rejected.title') }}
          </h2>
          <p class="text-gray-600 mb-6">
            {{ t('characterClaim.rejected.message') }}
          </p>
          <ul class="text-left text-gray-600 mb-6 list-disc pl-5">
            <li>{{ t('characterClaim.rejected.reasons.1') }}</li>
            <li>{{ t('characterClaim.rejected.reasons.2') }}</li>
          </ul>
          <button
            @click="resetClaim"
            class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700"
          >
            {{ t('characterClaim.rejected.tryAgain') }}
          </button>
        </div>
      </div>
    </main>
  </div>
</template>
