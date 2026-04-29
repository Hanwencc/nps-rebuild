<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { ApiError } from '@/api/request'

const { t } = useI18n()
const message = useMessage()
const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const form = reactive({ username: '', password: '' })
const loading = ref(false)

async function submit() {
  if (!form.username || !form.password) {
    message.warning(t('login.username') + ' / ' + t('login.password'))
    return
  }
  loading.value = true
  try {
    await auth.login(form.username, form.password)
    message.success(t('login.success'))
    const redirect = (route.query.redirect as string) || '/dashboard'
    router.replace(redirect)
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-500 to-violet-700 p-4">
    <NCard
      :title="t('login.title')"
      class="w-full max-w-sm shadow-2xl"
      :bordered="false"
    >
      <NForm @submit.prevent="submit">
        <NFormItem :label="t('login.username')">
          <NInput v-model:value="form.username" autofocus />
        </NFormItem>
        <NFormItem :label="t('login.password')">
          <NInput
            v-model:value="form.password"
            type="password"
            show-password-on="mousedown"
            @keyup.enter="submit"
          />
        </NFormItem>
        <NButton
          type="primary"
          block
          :loading="loading"
          attr-type="submit"
        >
          {{ t('login.submit') }}
        </NButton>
      </NForm>
    </NCard>
  </div>
</template>
