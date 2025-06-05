<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useSEO } from '@/composables/useSEO'
import BreadcrumbNavigation from '@/components/BreadcrumbNavigation.vue'

const { setBlogPostSEO } = useSEO()

interface BlogPost {
  id: string
  title: string
  excerpt: string
  date: string
  author: string
  image?: string
  tags: string[]
}

const route = useRoute()
const router = useRouter()

const post = ref<BlogPost | null>(null)
const content = ref('')
const loading = ref(true)
const error = ref('')

const formattedDate = computed(() => {
  if (!post.value) return ''
  return new Date(post.value.date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
})

const tagColors = {
  feature: 'bg-gradient-to-r from-blue-500 to-blue-600 text-white shadow-lg',
  collaboration: 'bg-gradient-to-r from-green-500 to-emerald-600 text-white shadow-lg',
  communication: 'bg-gradient-to-r from-purple-500 to-violet-600 text-white shadow-lg',
  improvement: 'bg-gradient-to-r from-orange-500 to-amber-600 text-white shadow-lg',
  bugfix: 'bg-gradient-to-r from-red-500 to-rose-600 text-white shadow-lg'
}

const getTagColor = (tag: string) => {
  return tagColors[tag as keyof typeof tagColors] || 'bg-gradient-to-r from-gray-500 to-slate-600 text-white shadow-lg'
}

const loadBlogPost = async () => {
  try {
    const slug = route.params.slug as string

    // Load blog index to find the post
    const indexResponse = await fetch('/assets/blog/index.json')
    const posts: BlogPost[] = await indexResponse.json()

    const foundPost = posts.find(p => p.id === slug)
    if (!foundPost) {
      error.value = 'Blog post not found'
      loading.value = false
      return
    }

    post.value = foundPost

    // Set SEO data for the blog post
    setBlogPostSEO(foundPost)

    // Load the markdown content
    const contentResponse = await fetch(`/assets/blog/posts/${foundPost.date}-${foundPost.id}.md`)
    if (!contentResponse.ok) {
      throw new Error('Failed to load post content')
    }

    const markdownContent = await contentResponse.text()
    // Simple markdown to HTML conversion (basic)
    content.value = markdownToHtml(markdownContent)

    loading.value = false
  } catch (err) {
    console.error('Error loading blog post:', err)
    error.value = 'Failed to load blog post'
    loading.value = false
  }
}

// Simple markdown to HTML converter (basic implementation)
const markdownToHtml = (markdown: string): string => {
  return markdown
    // Images with responsive styling - constrained to reasonable sizes
    .replace(/!\[([^\]]*)\]\(([^)]+)\)/g, '<img src="$2" alt="$1" class="max-w-lg w-full h-auto rounded-lg shadow-md my-4 mx-auto block" style="max-height: 400px; object-fit: contain;" />')

    // Icons (using emoji or simple text icons)
    .replace(/:icon-([^:]+):/g, '<span class="inline-flex items-center justify-center w-5 h-5 text-sm">$1</span>')

    // Headers with emojis - much tighter spacing
    .replace(/^## (✨|🔧|🚀|💬|📊|🎯|🔥|⚡|🌟|🛡️) (.*$)/gm, '<h2 class="text-lg font-bold text-gray-900 mb-2 mt-4 flex items-center gap-2"><span class="text-xl">$1</span><span>$2</span></h2>')
    .replace(/^## (.*$)/gm, '<h2 class="text-lg font-bold text-gray-900 mb-2 mt-4">$1</h2>')
    .replace(/^# (.*$)/gm, '<h1 class="text-xl font-bold text-gray-900 mb-2">$1</h1>')
    .replace(/^### (.*$)/gm, '<h3 class="text-base font-semibold text-gray-900 mb-1 mt-3">$1</h3>')

    // Bold text with minimal styling
    .replace(/\*\*(.*?)\*\*/g, '<strong class="font-semibold text-gray-900">$1</strong>')

    // Italic text
    .replace(/\*(.*?)\*/g, '<em class="italic text-gray-600">$1</em>')

    // Links
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" class="text-indigo-600 hover:text-indigo-800 underline font-medium">$1</a>')

    // Code blocks (inline)
    .replace(/`([^`]+)`/g, '<code class="bg-gray-100 text-gray-800 px-1 py-0.5 rounded text-sm">$1</code>')

    // Numbered lists - process each item separately to avoid <br> issues
    .replace(/^(\d+)\. (.*)$/gm, (match, num, content) => {
      return `<li class="mb-1 pl-1"><span class="font-medium text-indigo-600 mr-1">${num}.</span>${content}</li>`
    })
    .replace(/(<li class="mb-1 pl-1">.*?<\/li>(\s*<li class="mb-1 pl-1">.*?<\/li>)*)/gs, '<ol class="list-none mb-3 space-y-0.5 bg-gray-50 rounded p-3 border-l-2 border-indigo-400">$1</ol>')

    // Bullet lists - process each item separately to avoid <br> issues
    .replace(/^- (.*)$/gm, (match, content) => {
      return `<li class="mb-0.5 flex items-start"><span class="text-indigo-500 mr-1 mt-0.5 text-sm">•</span><span>${content}</span></li>`
    })
    .replace(/(<li class="mb-0.5 flex items-start">.*?<\/li>(\s*<li class="mb-0.5 flex items-start">.*?<\/li>)*)/gs, '<ul class="list-none mb-3 space-y-0.5 bg-blue-50 rounded p-3">$1</ul>')

    // Split into paragraphs by double newlines first
    .split(/\n\s*\n/)
    .map(paragraph => {
      const trimmed = paragraph.trim()
      // Skip if it's already an HTML element (headers, lists, images, etc.)
      if (trimmed.match(/^<[hluoi]/)) {
        return trimmed
      }
      // For regular text paragraphs, only add <br> for intentional line breaks
      if (trimmed && !trimmed.includes('<li>')) {
        const content = trimmed.replace(/\n/g, '<br>')
        return `<p class="mb-3 text-gray-700 leading-relaxed">${content}</p>`
      }
      return trimmed
    })
    .filter(p => p.trim())
    .join('')

    // Clean up any remaining issues
    .replace(/<p class="mb-3 text-gray-700 leading-relaxed"><\/p>/g, '')
}

onMounted(() => {
  loadBlogPost()
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 py-8">
    <!-- Breadcrumb Navigation -->
    <BreadcrumbNavigation :page-title="post?.title" />

    <div class="grid lg:grid-cols-3 gap-8">
      <!-- Main content column -->
      <div class="lg:col-span-2">
        <!-- Back button -->
        <button
          @click="router.back()"
          class="mb-6 inline-flex items-center text-indigo-600 hover:text-indigo-800 transition-colors"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
          Back to Blog
        </button>

        <!-- Loading state -->
        <div v-if="loading" class="flex justify-center items-center py-12">
          <div class="animate-spin h-8 w-8 border-2 border-indigo-500 border-t-transparent rounded-full"></div>
          <span class="ml-3 text-gray-600">Loading post...</span>
        </div>

        <!-- Error state -->
        <div v-else-if="error" class="text-center py-12">
          <div class="bg-red-50 border border-red-200 rounded-lg p-6">
            <p class="text-red-700">{{ error }}</p>
          </div>
        </div>

        <!-- Blog post content -->
        <article v-else-if="post" class="bg-white rounded-xl shadow-lg border border-gray-200 overflow-hidden">
          <!-- Content -->
          <div class="p-4">
            <!-- Meta info -->
            <div class="flex items-center justify-between mb-3 text-sm text-gray-500 border-b border-gray-100 pb-3">
              <div class="flex items-center gap-2">
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clip-rule="evenodd" />
                </svg>
                <span>{{ formattedDate }}</span>
              </div>
              <div class="flex items-center gap-2">
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd" />
                </svg>
                <span>by {{ post.author }}</span>
              </div>
            </div>

            <!-- Tags -->
            <div class="flex flex-wrap gap-1.5 mb-4">
              <span
                v-for="tag in post.tags"
                :key="tag"
                :class="['inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium shadow-sm', getTagColor(tag)]"
              >
                {{ tag }}
              </span>
            </div>

            <!-- Title -->
            <h1 class="text-2xl font-bold text-gray-900 mb-4 leading-tight">{{ post.title }}</h1>

            <!-- Content with enhanced styling -->
            <div class="prose max-w-none text-gray-700 text-sm" v-html="content"></div>

            <!-- Footer call-to-action -->
            <div class="mt-6 p-3 bg-gradient-to-r from-indigo-50 to-blue-50 rounded-lg border border-indigo-200">
              <div class="flex items-center gap-3">
                <div class="flex-shrink-0">
                  <svg class="w-6 h-6 text-indigo-600" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10c0 3.866-3.582 7-8 7a8.841 8.841 0 01-4.083-.98L2 17l1.338-3.123C2.493 12.767 2 11.434 2 10c0-3.866 3.582-7 8-7s8 3.134 8 7zM7 9H5v2h2V9zm8 0h-2v2h2V9zM9 9h2v2H9V9z" clip-rule="evenodd" />
                  </svg>
                </div>
                <div>
                  <p class="text-sm font-medium text-indigo-900 mb-1">Try the new chat feature now!</p>
                  <p class="text-xs text-indigo-700">Visit your soul core lists and look for the chat bubble in the bottom-right corner.</p>
                </div>
              </div>
            </div>
          </div>
        </article>
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
