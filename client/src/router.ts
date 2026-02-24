import { createRouter, createWebHistory } from 'vue-router';
import { useAuth } from './composables/useAuth';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: () => import('./pages/HomePage.vue'),
    },
    {
      path: '/my-songs',
      component: () => import('./pages/MySongsPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/users',
      component: () => import('./pages/UsersPage.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    {
      path: '/users/:id',
      component: () => import('./pages/UserDetailPage.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
  ],
});

router.beforeEach((to) => {
  const { isAuthenticated, isAdmin } = useAuth();

  if (to.meta.requiresAuth && !isAuthenticated.value) {
    return '/';
  }

  if (to.meta.requiresAdmin && !isAdmin.value) {
    return '/';
  }
});

export default router;
