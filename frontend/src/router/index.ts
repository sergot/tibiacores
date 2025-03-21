import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

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
      component: () => import('../views/SignInView.vue'),
    },
    {
      path: '/signup',
      name: 'signup',
      component: () => import('../views/SignUpView.vue'),
    },
    {
      path: '/create-list',
      name: 'create-list',
      component: () => import('../views/CreateListView.vue'),
      props: (route) => ({ character: route.query.character }),
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('../views/ProfileView.vue'),
    },
    {
      path: '/lists/:id',
      name: 'list-detail',
      component: () => import('../views/ListDetailView.vue'),
      props: true,
    },
    {
      path: '/join/:share_code',
      name: 'join-list',
      component: () => import('../views/JoinListView.vue'),
      props: true,
    },
  ],
})

export default router
