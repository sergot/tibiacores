<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'

interface BlogPost {
  id: string
  title: string
  excerpt: string
  date: string
  author: string
  image?: string
  icon?: string
  tags: string[]
}

const recentPosts = ref<BlogPost[]>([])
const loading = ref(true)

const loadRecentPosts = async () => {
  try {
    const response = await fetch('/assets/blog/index.json')
    if (!response.ok) {
      throw new Error('Failed to load blog posts')
    }
    const data = await response.json()

    // Sort by date and take the 3 most recent posts
    const sortedPosts = data.sort((a: BlogPost, b: BlogPost) =>
      new Date(b.date).getTime() - new Date(a.date).getTime()
    )

    recentPosts.value = sortedPosts.slice(0, 3)
    loading.value = false
  } catch (err) {
    console.error('Error loading recent blog posts:', err)
    loading.value = false
  }
}

onMounted(() => {
  loadRecentPosts()
})
</script>

<template>
  <div v-if="!loading && recentPosts.length > 0" class="bg-gradient-to-r from-indigo-50 to-blue-50 border-b border-gray-200">
    <div class="max-w-6xl mx-auto px-8 py-4">
      <div class="flex items-center justify-between mb-3">
        <div class="flex items-center space-x-2">
          <svg class="h-6 w-6 text-indigo-600" fill="currentColor" viewBox="0 0 24 24">
            <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z"/>
          </svg>
          <h2 class="text-lg font-bold text-gray-900">Latest News</h2>
        </div>
        <RouterLink
          to="/blog"
          class="text-sm font-medium text-indigo-600 hover:text-indigo-800 transition-colors"
        >
          View All →
        </RouterLink>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        <RouterLink
          v-for="post in recentPosts"
          :key="post.id"
          :to="`/blog/${post.id}`"
          class="bg-white rounded-lg p-3 shadow-sm border border-gray-200 hover:shadow-md hover:border-indigo-200 transition-all duration-200 group"
        >
          <div class="flex items-start space-x-2">
            <div class="flex-shrink-0">
              <div class="w-8 h-8 bg-indigo-100 rounded-lg flex items-center justify-center">
                <span v-if="post.icon" class="text-lg">{{ post.icon }}</span>
                <svg v-else class="h-4 w-4 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.746 0 3.332.477 4.5 1.253v13C19.832 18.477 18.246 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/>
                </svg>
              </div>
            </div>
            <div class="flex-1 min-w-0">
              <h3 class="text-sm font-semibold text-gray-900 group-hover:text-indigo-600 transition-colors leading-tight">
                {{ post.title }}
              </h3>
              <p class="text-xs text-gray-500 mt-1">
                {{ new Date(post.date).toLocaleDateString('en-US', {
                  month: 'short',
                  day: 'numeric'
                }) }}
              </p>
            </div>
          </div>
        </RouterLink>
      </div>
    </div>
  </div>
</template>
