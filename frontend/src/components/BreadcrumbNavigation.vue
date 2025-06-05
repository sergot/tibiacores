<template>
  <nav aria-label="Breadcrumb" class="py-4">
    <ol class="flex items-center space-x-2 text-sm" itemscope itemtype="https://schema.org/BreadcrumbList">
      <li
        v-for="(item, index) in breadcrumbs"
        :key="index"
        class="flex items-center"
        itemprop="itemListElement"
        itemscope
        itemtype="https://schema.org/ListItem"
      >
        <span v-if="index > 0" class="mx-2 text-gray-400">/</span>

        <router-link
          v-if="!item.isLast"
          :to="item.path"
          class="text-indigo-600 hover:text-indigo-800 transition-colors duration-200"
          itemprop="item"
        >
          <span itemprop="name">{{ item.name }}</span>
        </router-link>

        <span
          v-else
          class="text-gray-500 font-medium"
          itemprop="name"
        >
          {{ item.name }}
        </span>

        <meta itemprop="position" :content="(index + 1).toString()">
      </li>
    </ol>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const { t } = useI18n()

interface Props {
  pageTitle?: string // Custom title to override default breadcrumb title
}

const props = withDefaults(defineProps<Props>(), {
  pageTitle: undefined
})

interface BreadcrumbItem {
  name: string
  path: string
  isLast: boolean
}

const breadcrumbs = computed((): BreadcrumbItem[] => {
  const crumbs: BreadcrumbItem[] = []

  // Always start with home
  crumbs.push({
    name: t('breadcrumb.home'),
    path: '/',
    isLast: false
  })

  // Handle specific routes
  if (route.name === 'about') {
    crumbs.push({
      name: t('breadcrumb.about'),
      path: '/about',
      isLast: true
    })
  } else if (route.name === 'blog') {
    crumbs.push({
      name: t('breadcrumb.blog'),
      path: '/blog',
      isLast: true
    })
  } else if (route.name === 'blog-post') {
    crumbs.push({
      name: t('breadcrumb.blog'),
      path: '/blog',
      isLast: false
    })
    crumbs.push({
      name: props.pageTitle || route.params.slug as string || t('breadcrumb.blogPost'),
      path: route.path,
      isLast: true
    })
  } else if (route.name === 'highscores') {
    crumbs.push({
      name: t('breadcrumb.highscores'),
      path: '/highscores',
      isLast: true
    })
  } else if (route.name === 'public-character') {
    crumbs.push({
      name: t('breadcrumb.characters'),
      path: '/highscores',
      isLast: false
    })
    crumbs.push({
      name: route.params.name as string || t('breadcrumb.character'),
      path: route.path,
      isLast: true
    })
  } else if (route.name === 'sponsor') {
    crumbs.push({
      name: t('breadcrumb.sponsor'),
      path: '/sponsor',
      isLast: true
    })
  } else if (route.name === 'privacy') {
    crumbs.push({
      name: t('breadcrumb.privacy'),
      path: '/privacy',
      isLast: true
    })
  }

  // Mark the last item as last
  if (crumbs.length > 1) {
    crumbs.forEach((crumb, index) => {
      crumb.isLast = index === crumbs.length - 1
    })
  }

  return crumbs
})
</script>
