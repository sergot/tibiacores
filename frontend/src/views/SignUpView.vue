<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { useRouter, RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { PlusIcon } from '@heroicons/vue/24/solid'
import axios from 'axios'

const userStore = useUserStore()
const router = useRouter()
const { t } = useI18n()

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)

onMounted(() => {
  // Redirect to home if user already has email
  if (userStore.hasEmail) {
    router.push('/')
  }
})

const handleSubmit = async () => {
  if (loading.value) return
  loading.value = true

  try {
    if (password.value !== confirmPassword.value) {
      error.value = t('auth.errors.passwordMismatch')
      return
    }

    const response = await axios.post('/signup', {
      email: email.value,
      password: password.value,
    })

    const token = response.headers['x-auth-token']
    if (!token) {
      throw new Error('No token received')
    }

    userStore.setUser({
      session_token: token,
      id: response.data.id,
      has_email: response.data.has_email,
    })

    router.push('/')
  } catch (err) {
    if (axios.isAxiosError(err) && err.response) {
      error.value = err.response.data.message || t('auth.errors.accountNotFound')
    } else {
      error.value = t('auth.errors.accountNotFound')
    }
  } finally {
    loading.value = false
  }
}

const handleDiscordSignup = () => {
  router.push('/oauth/discord')
}

const handleGoogleSignup = () => {
  router.push('/oauth/google')
}
</script>

<template>
  <div
    class="min-h-[calc(100vh-8rem)] flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8 bg-gray-100"
  >
    <main class="max-w-md w-full space-y-8">
      <div>
        <div class="flex justify-center">
          <img class="h-20 w-20" src="/logo.png" alt="Logo" />
        </div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
          {{ t('auth.signUp.title') }}
        </h2>
        <p class="mt-2 text-center text-sm text-gray-600">
          {{ t('auth.signUp.hasAccount') }}
          <RouterLink to="/signin" class="font-medium text-indigo-600 hover:text-indigo-500">
            {{ t('auth.signUp.signIn') }}
          </RouterLink>
        </p>
      </div>

      <div v-if="error" class="rounded-md bg-red-50 p-4">
        <div class="flex">
          <div class="ml-3">
            <h3 class="text-sm font-medium text-red-800">{{ error }}</h3>
          </div>
        </div>
      </div>

      <div class="space-y-4">
        <!-- Social Login Buttons -->
        <div class="grid grid-cols-2 gap-4">
          <button
            @click="handleDiscordSignup"
            type="button"
            class="relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-[#5865F2] hover:bg-[#4752C4] focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[#5865F2]"
          >
            <span class="absolute left-0 inset-y-0 flex items-center pl-3">
              <svg class="h-5 w-5" viewBox="0 0 71 55" fill="currentColor">
                <path
                  d="M60.1045 4.8978C55.5792 2.8214 50.7265 1.2916 45.6527 0.41542C45.5603 0.39851 45.468 0.440769 45.4204 0.525289C44.7963 1.6353 44.105 3.0834 43.6209 4.2216C38.1637 3.4046 32.7345 3.4046 27.3892 4.2216C26.905 3.0581 26.1886 1.6353 25.5617 0.525289C25.5141 0.443589 25.4218 0.40133 25.3294 0.41542C20.2584 1.2888 15.4057 2.8186 10.8776 4.8978C10.8384 4.9147 10.8048 4.9429 10.7825 4.9795C1.57795 18.7309 -0.943561 32.1443 0.293408 45.3914C0.299005 45.4562 0.335386 45.5182 0.385761 45.5576C6.45866 50.0174 12.3413 52.7249 18.1147 54.5195C18.2071 54.5477 18.305 54.5139 18.3638 54.4378C19.7295 52.5728 20.9469 50.6063 21.9907 48.5383C22.0523 48.4172 21.9935 48.2735 21.8676 48.2256C19.9366 47.4931 18.0979 46.6 16.3292 45.5858C16.1893 45.5041 16.1781 45.304 16.3068 45.2082C16.679 44.9293 17.0513 44.6391 17.4067 44.3461C17.471 44.2926 17.5606 44.2813 17.6362 44.3151C29.2558 49.6202 41.8354 49.6202 53.3179 44.3151C53.3935 44.2785 53.4831 44.2898 53.5502 44.3433C53.9057 44.6363 54.2779 44.9293 54.6529 45.2082C54.7816 45.304 54.7732 45.5041 54.6333 45.5858C52.8646 46.6197 51.0259 47.4931 49.0921 48.2228C48.9662 48.2707 48.9102 48.4172 48.9718 48.5383C50.038 50.6034 51.2554 52.5699 52.5959 54.435C52.6519 54.5139 52.7526 54.5477 52.845 54.5195C58.6464 52.7249 64.529 50.0174 70.6019 45.5576C70.6551 45.5182 70.6887 45.459 70.6943 45.3942C72.1747 30.0791 68.2147 16.7757 60.1968 4.9823C60.1772 4.9429 60.1437 4.9147 60.1045 4.8978ZM23.7259 37.3253C20.2276 37.3253 17.3451 34.1136 17.3451 30.1693C17.3451 26.225 20.1717 23.0133 23.7259 23.0133C27.308 23.0133 30.1626 26.2532 30.1066 30.1693C30.1066 34.1136 27.28 37.3253 23.7259 37.3253ZM47.3178 37.3253C43.8196 37.3253 40.9371 34.1136 40.9371 30.1693C40.9371 26.225 43.7636 23.0133 47.3178 23.0133C50.9 23.0133 53.7545 26.2532 53.6986 30.1693C53.6986 34.1136 50.9 37.3253 47.3178 37.3253Z"
                />
              </svg>
            </span>
            {{ t('auth.signUp.withProvider.discord') }}
          </button>
          <button
            @click="handleGoogleSignup"
            type="button"
            class="relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
          >
            <span class="absolute left-0 inset-y-0 flex items-center pl-3">
              <svg class="h-5 w-5" viewBox="0 0 24 24">
                <path
                  fill="currentColor"
                  d="M12.545,10.239v3.821h5.445c-0.712,2.315-2.647,3.972-5.445,3.972c-3.332,0-6.033-2.701-6.033-6.032s2.701-6.032,6.033-6.032c1.498,0,2.866,0.549,3.921,1.453l2.814-2.814C17.503,2.988,15.139,2,12.545,2C7.021,2,2.543,6.477,2.543,12s4.478,10,10.002,10c8.396,0,10.249-7.85,9.426-11.748L12.545,10.239z"
                />
              </svg>
            </span>
            {{ t('auth.signUp.withProvider.google') }}
          </button>
        </div>

        <div class="relative">
          <div class="absolute inset-0 flex items-center">
            <div class="w-full border-t border-gray-300"></div>
          </div>
          <div class="relative flex justify-center text-sm">
            <span class="px-2 bg-gray-100 text-gray-500">{{
              t('auth.signUp.withProvider.title')
            }}</span>
          </div>
        </div>

        <form class="mt-8 space-y-6" @submit.prevent="handleSubmit">
          <div class="rounded-md shadow-sm -space-y-px">
            <div>
              <label for="email-address" class="sr-only">{{
                t('auth.signUp.withEmail.email')
              }}</label>
              <input
                v-model="email"
                id="email-address"
                name="email"
                type="email"
                autocomplete="email"
                required
                class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                :placeholder="t('auth.signUp.withEmail.emailPlaceholder')"
              />
            </div>
            <div>
              <label for="password" class="sr-only">{{
                t('auth.signUp.withEmail.password')
              }}</label>
              <input
                v-model="password"
                id="password"
                name="password"
                type="password"
                autocomplete="new-password"
                required
                class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                :placeholder="t('auth.signUp.withEmail.passwordPlaceholder')"
              />
            </div>
            <div>
              <label for="confirm-password" class="sr-only">{{
                t('auth.signUp.withEmail.confirmPassword')
              }}</label>
              <input
                v-model="confirmPassword"
                id="confirm-password"
                name="confirm-password"
                type="password"
                autocomplete="new-password"
                required
                class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                :placeholder="t('auth.signUp.withEmail.confirmPasswordPlaceholder')"
              />
            </div>
          </div>

          <div>
            <button
              type="submit"
              :disabled="loading"
              class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:bg-indigo-400"
            >
              <span class="absolute left-0 inset-y-0 flex items-center pl-3">
                <PlusIcon
                  class="h-5 w-5 text-indigo-500 group-hover:text-indigo-400"
                  aria-hidden="true"
                />
              </span>
              {{
                loading ? t('auth.signUp.withEmail.creating') : t('auth.signUp.withEmail.submit')
              }}
            </button>
          </div>
        </form>
      </div>
    </main>
  </div>
</template>
