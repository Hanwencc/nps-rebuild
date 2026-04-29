<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  NCard,
  NCollapse,
  NCollapseItem,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSwitch,
  NSelect,
  NTag,
  NSpace,
  NButton,
  NAlert,
  NEmpty,
  useMessage,
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { settingsApi, type SettingItem } from '@/api/settings'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const message = useMessage()
const auth = useAuthStore()

const loading = ref(false)
const saving = ref(false)
const items = ref<SettingItem[]>([])
// staged values keyed by item.key — only sent on save
const draft = reactive<Record<string, string>>({})
// keys that the user has touched; submitted on Save
const dirty = reactive<Set<string>>(new Set())

const groups = computed(() => {
  const map = new Map<string, SettingItem[]>()
  for (const it of items.value) {
    if (!map.has(it.group)) map.set(it.group, [])
    map.get(it.group)!.push(it)
  }
  return Array.from(map.entries())
})

const dirtyCount = computed(() => dirty.size)

async function load() {
  if (!auth.isAdmin) return
  loading.value = true
  try {
    const list = await settingsApi.list()
    items.value = list
    for (const it of list) {
      draft[it.key] = it.value
    }
    dirty.clear()
  } catch (e: unknown) {
    message.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

function onChange(key: string, val: string | number | boolean | null) {
  let s: string
  if (typeof val === 'boolean') s = val ? 'true' : 'false'
  else if (val === null || val === undefined) s = ''
  else s = String(val)
  draft[key] = s
  dirty.add(key)
}

function valueOf(it: SettingItem): string | number | boolean | null {
  const raw = draft[it.key] ?? ''
  if (it.type === 'bool') return raw === 'true' || raw === '1'
  if (it.type === 'int') {
    const n = parseInt(raw, 10)
    return Number.isNaN(n) ? null : n
  }
  return raw
}

async function save() {
  if (dirty.size === 0) {
    message.info(t('settings.nothingToSave'))
    return
  }
  saving.value = true
  try {
    const payload: Record<string, string> = {}
    for (const k of dirty) {
      // strip bootstrap keys client-side too — server also rejects them
      const it = items.value.find((x) => x.key === k)
      if (!it || it.bootstrap) continue
      payload[k] = draft[k]
    }
    if (Object.keys(payload).length === 0) {
      message.info(t('settings.nothingToSave'))
      return
    }
    const res = await settingsApi.update(payload)
    message.success(
      t('settings.savedN', { n: res.applied }) +
        (res.rejected && res.rejected.length
          ? ' (' + t('settings.rejected') + ': ' + res.rejected.join(', ') + ')'
          : ''),
    )
    await load()
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
      {{ t('settings.adminOnly') }}
    </NAlert>

    <template v-else>
      <NAlert type="info" :show-icon="false">
        {{ t('settings.intro') }}
      </NAlert>

      <NCard :loading="loading">
        <template #header>
          <span>{{ t('nav.settings') }}</span>
        </template>
        <template #header-extra>
          <NSpace>
            <NTag v-if="dirtyCount > 0" type="warning" size="small">
              {{ t('settings.dirtyN', { n: dirtyCount }) }}
            </NTag>
            <NButton
              type="primary"
              :loading="saving"
              :disabled="dirtyCount === 0"
              @click="save"
            >
              {{ t('common.save') }}
            </NButton>
            <NButton @click="load" :disabled="saving">
              {{ t('common.refresh') }}
            </NButton>
          </NSpace>
        </template>

        <NEmpty v-if="!loading && items.length === 0" />

        <NCollapse :default-expanded-names="groups.map((g) => g[0])">
          <NCollapseItem
            v-for="[group, list] in groups"
            :key="group"
            :title="group"
            :name="group"
          >
            <NForm label-placement="left" label-width="220">
              <NFormItem
                v-for="it in list"
                :key="it.key"
                :label="it.label"
              >
                <NSpace vertical :size="2" style="width: 100%">
                  <NSpace>
                    <!-- bool -->
                    <NSwitch
                      v-if="it.type === 'bool'"
                      :value="valueOf(it) as boolean"
                      :disabled="it.bootstrap"
                      @update:value="(v) => onChange(it.key, v)"
                    />
                    <!-- int -->
                    <NInputNumber
                      v-else-if="it.type === 'int'"
                      :value="valueOf(it) as number | null"
                      :disabled="it.bootstrap"
                      style="width: 220px"
                      @update:value="(v) => onChange(it.key, v)"
                    />
                    <!-- enum -->
                    <NSelect
                      v-else-if="it.type === 'enum'"
                      :value="valueOf(it) as string"
                      :disabled="it.bootstrap"
                      :options="(it.enum ?? []).map((e) => ({ label: e, value: e }))"
                      style="width: 220px"
                      @update:value="(v) => onChange(it.key, v)"
                    />
                    <!-- password -->
                    <NInput
                      v-else-if="it.type === 'password'"
                      :value="valueOf(it) as string"
                      type="password"
                      show-password-on="click"
                      :disabled="it.bootstrap"
                      style="width: 320px"
                      @update:value="(v) => onChange(it.key, v)"
                    />
                    <!-- string -->
                    <NInput
                      v-else
                      :value="valueOf(it) as string"
                      :disabled="it.bootstrap"
                      style="width: 320px"
                      @update:value="(v) => onChange(it.key, v)"
                    />
                    <NTag v-if="it.bootstrap" size="small" type="error">
                      {{ t('settings.bootstrap') }}
                    </NTag>
                    <NTag v-else-if="it.needsRestart" size="small" type="warning">
                      {{ t('settings.needsRestart') }}
                    </NTag>
                  </NSpace>
                  <span
                    v-if="it.help"
                    style="color: var(--n-text-color-3); font-size: 12px"
                  >
                    {{ it.help }}
                  </span>
                </NSpace>
              </NFormItem>
            </NForm>
          </NCollapseItem>
        </NCollapse>
      </NCard>
    </template>
  </div>
</template>
