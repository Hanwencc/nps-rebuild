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
  <div class="nps-login-bg min-h-screen flex items-center justify-center p-4">
    <div class="nps-login-orbs" aria-hidden="true">
      <span class="orb orb-a" />
      <span class="orb orb-b" />
      <span class="orb orb-c" />
    </div>
    <NCard
      class="w-full max-w-sm nps-login-card"
      :bordered="false"
    >
      <template #header>
        <div class="flex items-center gap-3">
          <div class="nps-login-logo">N</div>
          <div class="leading-tight">
            <div class="text-base font-semibold nps-gradient-text">
              {{ t('login.title') }}
            </div>
            <div class="text-xs opacity-60">NPS · Control Panel</div>
          </div>
        </div>
      </template>
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

<style scoped>
.nps-login-bg {
  position: relative;
  overflow: hidden;
  background:
    radial-gradient(circle at 20% 20%, #312e81 0%, transparent 55%),
    radial-gradient(circle at 80% 80%, #0e7490 0%, transparent 55%),
    linear-gradient(135deg, #0b0d12 0%, #1e1b4b 100%);
}
.nps-login-bg::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.05) 1px, transparent 1px);
  background-size: 40px 40px;
  mask-image: radial-gradient(circle at 50% 50%, #000 40%, transparent 80%);
  pointer-events: none;
}
.nps-login-orbs .orb {
  position: absolute;
  border-radius: 999px;
  filter: blur(60px);
  opacity: 0.55;
  pointer-events: none;
  animation: nps-float 16s ease-in-out infinite;
}
.orb-a {
  width: 360px;
  height: 360px;
  background: #6366f1;
  top: -80px;
  left: -80px;
}
.orb-b {
  width: 280px;
  height: 280px;
  background: #06b6d4;
  bottom: -60px;
  right: -40px;
  animation-delay: -5s;
}
.orb-c {
  width: 200px;
  height: 200px;
  background: #a855f7;
  top: 40%;
  right: 20%;
  animation-delay: -10s;
}
@keyframes nps-float {
  0%,
  100% {
    transform: translate3d(0, 0, 0);
  }
  50% {
    transform: translate3d(20px, -30px, 0);
  }
}
.nps-login-card {
  position: relative;
  z-index: 1;
  background: rgba(20, 22, 30, 0.55) !important;
  backdrop-filter: saturate(180%) blur(18px);
  -webkit-backdrop-filter: saturate(180%) blur(18px);
  border: 1px solid rgba(255, 255, 255, 0.08) !important;
  box-shadow:
    0 20px 60px rgba(0, 0, 0, 0.45),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
  color: #e5e7eb;
}
.nps-login-card :deep(.n-card-header) {
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}
.nps-login-logo {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  color: #fff;
  font-weight: 700;
  background: linear-gradient(135deg, #6366f1 0%, #06b6d4 100%);
  box-shadow:
    0 8px 22px rgba(99, 102, 241, 0.45),
    inset 0 1px 0 rgba(255, 255, 255, 0.25);
}
</style>
