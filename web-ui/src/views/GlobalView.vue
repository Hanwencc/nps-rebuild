<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import {
  NCard,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NAlert,
  useMessage,
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { globalApi, type GlobalConfig } from '@/api/global'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const message = useMessage()
const auth = useAuthStore()

const loading = ref(false)
const saving = ref(false)
const serverUrl = ref('')
const blackIpText = ref('')

const blackIpList = computed<string[]>(() =>
  blackIpText.value
    .split(/\r?\n/)
    .map((s) => s.trim())
    .filter(Boolean),
)

async function load() {
  if (!auth.isAdmin) return
  loading.value = true
  try {
    const cfg = await globalApi.get()
    serverUrl.value = cfg.serverUrl ?? ''
    blackIpText.value = (cfg.blackIpList ?? []).join('\n')
  } catch (e: unknown) {
    message.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    const payload: GlobalConfig = {
      serverUrl: serverUrl.value.trim(),
      blackIpList: blackIpList.value,
    }
    await globalApi.update(payload)
    message.success(t('common.saved'))
  } catch (e: unknown) {
    message.error((e as Error).message)
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <NAlert v-if="!auth.isAdmin" type="warning">
      {{ t('global.adminOnly') }}
    </NAlert>

    <NCard v-else :title="t('nav.global')" :loading="loading">
      <NForm label-placement="top" :disabled="saving">
        <NFormItem :label="t('global.serverUrl')">
          <NInput
            v-model:value="serverUrl"
            placeholder="https://example.com"
            clearable
          />
        </NFormItem>
        <NFormItem :label="t('global.blackIpList')">
          <NInput
            v-model:value="blackIpText"
            type="textarea"
            :autosize="{ minRows: 6, maxRows: 16 }"
            :placeholder="t('global.blackIpListTip')"
          />
        </NFormItem>
        <NSpace>
          <NButton type="primary" :loading="saving" @click="save">
            {{ t('common.save') }}
          </NButton>
          <NButton @click="load" :disabled="saving">{{ t('common.refresh') }}</NButton>
        </NSpace>
      </NForm>
    </NCard>
  </div>
</template>
