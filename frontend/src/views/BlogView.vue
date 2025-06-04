<script setup lang="ts">
import { ref, onMounted } from 'vue'
import BlogCard from '@/components/BlogCard.vue'

interface BlogPost {
  id: string
  title: string
  excerpt: string
  date: string
  author: string
  image?: string
  tags: string[]
}

const posts = ref<BlogPost[]>([])
const loading = ref(true)
const error = ref('')

const loadBlogPosts = async () => {
  try {
    const response = await fetch('/assets/blog/index.json')
    if (!response.ok) {
      throw new Error('Failed to load blog posts')
    }
    const data = await response.json()

    // Sort posts by date (newest first)
    posts.value = data.sort((a: BlogPost, b: BlogPost) =>
      new Date(b.date).getTime() - new Date(a.date).getTime()
    )

    loading.value = false
  } catch (err) {
    console.error('Error loading blog posts:', err)
    error.value = 'Failed to load blog posts'
    loading.value = false
  }
}

onMounted(() => {
  loadBlogPosts()
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 py-8">
    <div class="grid lg:grid-cols-3 gap-8">
      <!-- Main content column -->
      <div class="lg:col-span-2">
        <!-- Header -->
        <div class="mb-12">
          <h1 class="text-4xl font-bold text-gray-900 mb-4">TibiaCores News & Updates</h1>
          <p class="text-xl text-gray-600">
            Stay up to date with the latest features, improvements, and announcements
          </p>
        </div>

        <!-- Loading state -->
        <div v-if="loading" class="flex justify-center items-center py-12">
          <div class="animate-spin h-8 w-8 border-2 border-indigo-500 border-t-transparent rounded-full"></div>
          <span class="ml-3 text-gray-600">Loading posts...</span>
        </div>

        <!-- Error state -->
        <div v-else-if="error" class="text-center py-12">
          <div class="bg-red-50 border border-red-200 rounded-lg p-6">
            <p class="text-red-700">{{ error }}</p>
          </div>
        </div>

        <!-- Blog posts -->
        <div v-else-if="posts.length > 0">
          <h2 class="text-2xl font-semibold text-gray-900 mb-6">All Posts</h2>
          <div class="space-y-6">
            <BlogCard
              v-for="post in posts"
              :key="post.id"
              :post="post"
            />
          </div>
        </div>

        <!-- Empty state -->
        <div v-else class="text-center py-12">
          <div class="bg-gray-50 border border-gray-200 rounded-lg p-8">
            <h3 class="text-lg font-medium text-gray-900 mb-2">No posts yet</h3>
            <p class="text-gray-600">Check back soon for updates and announcements!</p>
          </div>
        </div>
      </div>

      <!-- Sidebar -->
      <div class="lg:col-span-1">
        <div class="sticky top-8 space-y-6">
          <!-- Newsletter signup -->
          <div class="bg-indigo-50 border border-indigo-200 rounded-lg p-6">
            <h3 class="text-lg font-semibold text-indigo-900 mb-3">Stay Updated</h3>
            <p class="text-indigo-700 mb-4 text-sm">
              Want to be the first to know about new features and updates?
            </p>
            <div class="mb-4">
              <input
                type="email"
                placeholder="Enter your email"
                class="w-full px-3 py-2 border border-indigo-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                disabled
              />
            </div>
            <button
              class="w-full bg-indigo-600 text-white px-4 py-2 rounded-md text-sm font-medium opacity-50 cursor-not-allowed"
              disabled
            >
              Coming Soon
            </button>
            <p class="text-xs text-indigo-600 mt-2">
              💡 Newsletter signup functionality coming soon!
            </p>
          </div>

          <!-- Sponsor section -->
          <div class="bg-gradient-to-br from-yellow-50 to-orange-50 border border-yellow-200 rounded-lg p-6">
            <h3 class="text-lg font-semibold text-yellow-900 mb-3">Support TibiaCores</h3>
            <p class="text-yellow-800 mb-4 text-sm">
              Help us keep TibiaCores running and add new features by becoming a sponsor.
            </p>
            <RouterLink
              to="/sponsor"
              class="inline-flex items-center w-full justify-center bg-yellow-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-yellow-700 transition-colors"
            >
              <svg class="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/>
              </svg>
              Become a Sponsor
            </RouterLink>
            <p class="text-xs text-yellow-700 mt-2">
              Your support helps us improve TibiaCores for everyone!
            </p>
          </div>

          <!-- Quick links -->
          <div class="bg-gray-50 border border-gray-200 rounded-lg p-6">
            <h3 class="text-lg font-semibold text-gray-900 mb-3">Quick Links</h3>
            <div class="space-y-2">
              <RouterLink to="/highscores" class="block text-sm text-gray-600 hover:text-indigo-600 transition-colors">
                📊 Highscores
              </RouterLink>
              <RouterLink to="/about" class="block text-sm text-gray-600 hover:text-indigo-600 transition-colors">
                ℹ️ About TibiaCores
              </RouterLink>
              <RouterLink to="/privacy" class="block text-sm text-gray-600 hover:text-indigo-600 transition-colors">
                🔒 Privacy Policy
              </RouterLink>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
