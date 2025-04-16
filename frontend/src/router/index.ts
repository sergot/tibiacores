/// <reference types="vite/client" />

import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import HomeView from '@/views/HomeView.vue'
import CharacterClaimView from '@/views/CharacterClaimView.vue'
import SignUpView from '@/views/SignUpView.vue'
import SignInView from '@/views/SignInView.vue'
import ProfileView from '@/views/ProfileView.vue'
import CreateListView from '@/views/CreateListView.vue'
import ListDetailView from '@/views/ListDetailView.vue'
import JoinListView from '@/views/JoinListView.vue'
import CharacterDetailsView from '@/views/CharacterDetailsView.vue'
import PrivacyView from '@/views/PrivacyView.vue'
import PublicCharacterView from '@/views/PublicCharacterView.vue'
import HighscoreView from '@/views/HighscoreView.vue'
import SponsorView from '@/views/SponsorView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('../views/AboutView.vue'),
    },
    {
      path: '/highscores',
      name: 'highscores',
      component: HighscoreView,
    },
    {
      path: '/sponsor',
      name: 'sponsor',
      component: SponsorView,
    },
    {
      path: '/signin',
      name: 'signin',
      component: SignInView,
    },
    {
      path: '/signup',
      name: 'signup',
      component: SignUpView,
    },
    {
      path: '/create-list',
      name: 'create-list',
      component: CreateListView,
      props: (route) => ({ character: route.query.character }),
    },
    {
      path: '/profile',
      name: 'profile',
      component: ProfileView,
      meta: { requiresAuth: true },
    },
    {
      path: '/lists/:id',
      name: 'list-detail',
      component: ListDetailView,
      props: true,
      meta: { requiresAuth: true },
    },
    {
      path: '/join/:share_code',
      name: 'join-list',
      component: JoinListView,
      props: true,
    },
    {
      path: '/claim-character',
      name: 'claim-character',
      component: CharacterClaimView,
      meta: { requiresAuth: true },
    },
    {
      path: '/characters/:id',
      name: 'character-details',
      component: CharacterDetailsView,
      meta: { requiresAuth: true },
    },
    {
      path: '/characters/public/:name',
      name: 'public-character',
      component: PublicCharacterView,
    },
    {
      path: '/verify-email',
      name: 'verify-email',
      component: () => import('../views/EmailVerificationView.vue'),
    },
    {
      path: '/oauth/:provider/callback',
      name: 'oauth-callback',
      component: () => import('../views/OAuthCallbackView.vue'),
    },
    {
      path: '/oauth/:provider',
      name: 'oauth-initiate',
      component: () => import('../views/OAuthInitiateView.vue'),
    },
    {
      path: '/privacy',
      name: 'privacy',
      component: PrivacyView,
    },
  ],
})

// Navigation guard to check authentication
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()

  // Check if route requires authentication
  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth)

  // Allow navigation to public routes or if user is authenticated
  if (!requiresAuth) {
    return next()
  }

  // Check if user is authenticated
  if (userStore.isAuthenticated) {
    return next()
  }

  // Redirect to signin page if not authenticated
  return next('/signin')
})

export default router
