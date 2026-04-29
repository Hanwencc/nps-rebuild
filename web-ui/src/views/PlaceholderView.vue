<script setup lang="ts">
import { useRoute } from 'vue-router'
import { NCard, NEmpty } from 'naive-ui'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const { t } = useI18n()
const title = computed(() => {
  if (route.name === 'tunnels') return t('nav.' + (route.params.mode as string))
  return String(route.name ?? '')
})
</script>

<template>
  <NCard :title="title">
    <NEmpty description="Phase 1 中即将提供（敬请期待）" />
  </NCard>
</template>
