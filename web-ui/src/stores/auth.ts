import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { authApi } from '@/api/auth'
import type { UserInfo } from '@/api/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserInfo | null>(null)
  const ready = ref(false)

  const isAuthed = computed(() => user.value?.authed === true)
  const isAdmin = computed(() => user.value?.isAdmin === true)

  async function refresh() {
    try {
      user.value = await authApi.me()
    } catch {
      user.value = null
    } finally {
      ready.value = true
    }
  }

  async function login(username: string, password: string) {
    user.value = await authApi.login(username, password)
  }

  async function logout() {
    try {
      await authApi.logout()
    } finally {
      user.value = null
    }
  }

  return { user, ready, isAuthed, isAdmin, refresh, login, logout }
})
