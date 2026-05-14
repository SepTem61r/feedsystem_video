import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { accountApi } from '../api/account'
import type { AccountInfo } from '../types'

const TOKEN_KEY = 'feedsystem_token'
const USER_KEY = 'feedsystem_user'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const currentUser = ref<AccountInfo | null>(
    JSON.parse(localStorage.getItem(USER_KEY) || 'null'),
  )

  const isLoggedIn = computed(() => !!token.value)

  function setAuth(newToken: string, user: AccountInfo) {
    token.value = newToken
    currentUser.value = user
    localStorage.setItem(TOKEN_KEY, newToken)
    localStorage.setItem(USER_KEY, JSON.stringify(user))
  }

  function clearAuth() {
    token.value = null
    currentUser.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  }

  function parseToken(t: string): AccountInfo | null {
    try {
      const payload = JSON.parse(atob(t.split('.')[1]))
      if (payload.account_id && payload.username) {
        return { id: payload.account_id, username: payload.username }
      }
    } catch { /* invalid token format */ }
    return null
  }

  async function login(username: string, password: string) {
    const res = await accountApi.login({ username, password })
    const user = parseToken(res.data.token)
    if (!user) throw new Error('无法解析用户信息')
    setAuth(res.data.token, user)
    return user
  }

  async function register(username: string, password: string) {
    await accountApi.register({ username, password })
    await login(username, password)
  }

  async function logout() {
    try {
      await accountApi.logout()
    } catch { /* ignore logout API errors */ }
    clearAuth()
  }

  async function rename(newUsername: string) {
    const res = await accountApi.rename({ new_username: newUsername })
    const user = parseToken(res.data.token)
    if (!user) throw new Error('无法解析用户信息')
    setAuth(res.data.token, user)
    return user
  }

  return { token, currentUser, isLoggedIn, setAuth, clearAuth, login, register, logout, rename, parseToken }
})
