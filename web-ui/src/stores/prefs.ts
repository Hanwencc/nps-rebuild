import { defineStore } from 'pinia'
import { ref } from 'vue'

/**
 * Generic UI preferences (theme, language, sidebar collapsed) persisted
 * to localStorage so they survive page reloads.
 */
export const usePrefsStore = defineStore('prefs', () => {
  const dark = ref<boolean>(localStorage.getItem('nps:dark') === '1')
  const lang = ref<string>(localStorage.getItem('nps:lang') ?? 'zh-CN')
  const sidebarCollapsed = ref<boolean>(
    localStorage.getItem('nps:sidebar') === '1',
  )

  function setDark(v: boolean) {
    dark.value = v
    localStorage.setItem('nps:dark', v ? '1' : '0')
  }
  function setLang(v: string) {
    lang.value = v
    localStorage.setItem('nps:lang', v)
  }
  function setSidebar(v: boolean) {
    sidebarCollapsed.value = v
    localStorage.setItem('nps:sidebar', v ? '1' : '0')
  }

  return { dark, lang, sidebarCollapsed, setDark, setLang, setSidebar }
})
