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
  type GlobalThemeOverrides,
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

// Light/dark themed brand overrides for a more modern, "tech" feel.
const themeOverrides = computed<GlobalThemeOverrides>(() => {
  const dark = prefs.dark
  const accent = dark ? '#818cf8' : '#6366f1'
  const accentHover = dark ? '#a5b4fc' : '#818cf8'
  const accentPressed = dark ? '#6366f1' : '#4f46e5'
  return {
    common: {
      primaryColor: accent,
      primaryColorHover: accentHover,
      primaryColorPressed: accentPressed,
      primaryColorSuppl: accentHover,
      borderRadius: '10px',
      borderRadiusSmall: '8px',
      fontFamily:
        "'Inter','PingFang SC','Helvetica Neue',system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif",
      fontFamilyMono:
        "'JetBrains Mono','Fira Code',ui-monospace,SFMono-Regular,Menlo,Consolas,monospace",
      bodyColor: 'transparent',
    },
    Layout: {
      color: 'transparent',
      siderColor: dark
        ? 'rgba(20,22,30,0.66)'
        : 'rgba(255,255,255,0.72)',
      headerColor: dark
        ? 'rgba(20,22,30,0.66)'
        : 'rgba(255,255,255,0.72)',
    },
    Card: {
      borderRadius: '14px',
      paddingMedium: '18px 20px',
    },
    Button: {
      borderRadiusMedium: '10px',
    },
    Menu: {
      itemHeight: '40px',
      borderRadius: '10px',
    },
  }
})

// Reflect the dark preference on <html> so global CSS (body bg, etc.)
// can react via the `.dark` class.
watchEffect(() => {
  document.documentElement.classList.toggle('dark', prefs.dark)
})
</script>

<template>
  <NConfigProvider
    :theme="theme"
    :theme-overrides="themeOverrides"
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
