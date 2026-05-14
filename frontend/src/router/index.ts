import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../pages/LoginPage.vue'),
      meta: { guest: true },
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('../pages/RegisterPage.vue'),
      meta: { guest: true },
    },
    {
      path: '/',
      name: 'Feed',
      component: () => import('../pages/SwipeFeed.vue'),
      meta: { hideNav: true },
    },
    {
      path: '/explore',
      name: 'Explore',
      component: () => import('../pages/FeedPage.vue'),
    },
    {
      path: '/video/:id',
      name: 'VideoDetail',
      component: () => import('../pages/VideoDetailPage.vue'),
    },
    {
      path: '/profile/:id',
      name: 'Profile',
      component: () => import('../pages/ProfilePage.vue'),
    },
    {
      path: '/upload',
      name: 'Upload',
      component: () => import('../pages/UploadPage.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else if (to.meta.guest && auth.isLoggedIn) {
    next({ name: 'Feed' })
  } else {
    next()
  }
})

export default router
