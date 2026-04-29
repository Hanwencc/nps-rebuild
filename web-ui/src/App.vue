<script setup lang="ts">
import { computed, watchEffect } from 'vue'
import {
  NConfigProvider,
  NMessageProvider,
  NDialogProvider,
  NLoadingBarProvider,
  NNotificationProvider,
  zhCN,
  enUS,
  dateZhCN,
  dateEnUS,
  darkTheme,
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { usePrefsStore } from '@/stores/prefs'

const prefs = usePrefsStore()
const { locale } = useI18n()

const naiveLocale = computed(() => (locale.value === 'en' ? enUS : zhCN))
const naiveDateLocale = computed(() =>
  locale.value === 'en' ? dateEnUS : dateZhCN,
)
const theme = computed(() => (prefs.dark ? darkTheme : null))

// Reflect the dark preference on <html> so global CSS (body bg, etc.)
// can react via the `.dark` class.
watchEffect(() => {
  document.documentElement.classList.toggle('dark', prefs.dark)
})
</script>

<template>
  <NConfigProvider
    :theme="theme"
    :locale="naiveLocale"
    :date-locale="naiveDateLocale"
    inline-theme-disabled
  >
    <NLoadingBarProvider>
      <NDialogProvider>
        <NNotificationProvider>
          <NMessageProvider>
            <RouterView />
          </NMessageProvider>
        </NNotificationProvider>
      </NDialogProvider>
    </NLoadingBarProvider>
  </NConfigProvider>
</template>
