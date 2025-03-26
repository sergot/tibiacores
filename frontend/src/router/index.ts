/// <reference types="vite/client" />

import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import CharacterClaimView from '../views/CharacterClaimView.vue'
import SignUpView from '../views/SignUpView.vue'
import SignInView from '../views/SignInView.vue'
import ProfileView from '../views/ProfileView.vue'
import CreateListView from '../views/CreateListView.vue'
import ListDetailView from '../views/ListDetailView.vue'
import JoinListView from '../views/JoinListView.vue'
import CharacterDetailsView from '../views/CharacterDetailsView.vue'

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
    },
    {
      path: '/lists/:id',
      name: 'list-detail',
      component: ListDetailView,
      props: true,
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
      meta: { requiresAuth: true }
    },
    {
      path: '/characters/:id',
      name: 'character-details',
      component: CharacterDetailsView,
      meta: { requiresAuth: true }
    },
    {
      path: '/verify-email',
      name: 'verify-email',
      component: () => import('../views/EmailVerificationView.vue')
    }
  ]
})

export default router
