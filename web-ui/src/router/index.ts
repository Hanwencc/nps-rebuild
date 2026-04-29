import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true },
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/DashboardView.vue'),
      },
      {
        path: 'clients',
        name: 'clients',
        component: () => import('@/views/clients/ClientListView.vue'),
      },
      {
        path: 'clients/new',
        name: 'client-new',
        component: () => import('@/views/clients/ClientFormView.vue'),
      },
      {
        path: 'clients/:id/edit',
        name: 'client-edit',
        component: () => import('@/views/clients/ClientFormView.vue'),
        props: true,
      },
      // Phase 1+ placeholders so the sidebar links don't 404 right away
      {
        path: 'tunnels/:mode',
        name: 'tunnels',
        component: () => import('@/views/tunnels/TunnelListView.vue'),
        props: true,
      },
      {
        path: 'hosts',
        name: 'hosts',
        component: () => import('@/views/hosts/HostListView.vue'),
      },
      {
        path: 'global',
        name: 'global',
        component: () => import('@/views/GlobalView.vue'),
      },
      {
        path: 'settings',
        name: 'settings',
        component: () => import('@/views/SettingsView.vue'),
      },
      {
        path: 'tokens',
        name: 'tokens',
        component: () => import('@/views/TokenListView.vue'),
      },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
]

const router = createRouter({
  history: createWebHashHistory('/ui/'),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.ready) {
    await auth.refresh()
  }
  if (to.meta.public) {
    if (auth.isAuthed && to.name === 'login') return { name: 'dashboard' }
    return true
  }
  if (!auth.isAuthed) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  return true
})

export default router
