<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

interface BlogPost {
  id: string
  title: string
  excerpt: string
  date: string
  author: string
  image?: string
  tags: string[]
}

const props = defineProps<{
  post: BlogPost
  compact?: boolean
}>()

const formattedDate = computed(() => {
  return new Date(props.post.date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
})

const tagColors = {
  feature: 'bg-blue-100 text-blue-800',
  collaboration: 'bg-green-100 text-green-800',
  communication: 'bg-purple-100 text-purple-800',
  improvement: 'bg-orange-100 text-orange-800',
  bugfix: 'bg-red-100 text-red-800'
}

const getTagColor = (tag: string) => {
  return tagColors[tag as keyof typeof tagColors] || 'bg-gray-100 text-gray-800'
}
</script>

<template>
  <article
    :class="[
      'bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow duration-200',
      compact ? 'flex gap-4' : ''
    ]"
  >

    <!-- Content -->
    <div :class="['p-6', compact ? 'flex-1' : '']">


      <!-- Tags -->
      <div v-if="!compact" class="flex flex-wrap gap-2 mb-3">
        <span
          v-for="tag in post.tags"
          :key="tag"
          :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium', getTagColor(tag)]"
        >
          {{ tag }}
        </span>
      </div>

      <!-- Title -->
      <h3 :class="[compact ? 'text-lg' : 'text-xl', 'font-semibold text-gray-900 mb-2']">
        <RouterLink
          :to="{ name: 'blog-post', params: { slug: post.id } }"
          class="hover:text-indigo-600 transition-colors"
        >
          {{ post.title }}
        </RouterLink>
      </h3>

      <!-- Excerpt -->
      <p :class="['text-gray-600 mb-4', compact ? 'text-sm' : '']">
        {{ post.excerpt }}
      </p>

      <!-- Meta info -->
      <div class="flex items-center justify-between text-sm text-gray-500">
        <span>{{ formattedDate }}</span>
        <span>by {{ post.author }}</span>
      </div>

      <!-- Read more link for compact view -->
      <div v-if="compact" class="mt-3">
        <RouterLink
          :to="{ name: 'blog-post', params: { slug: post.id } }"
          class="text-indigo-600 hover:text-indigo-800 text-sm font-medium"
        >
          Read more →
        </RouterLink>
      </div>
    </div>
  </article>
</template>
