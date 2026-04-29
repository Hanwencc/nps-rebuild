<script setup lang="ts">
import { computed, h } from 'vue'
import { useRoute, useRouter, RouterView } from 'vue-router'
import {
  NLayout,
  NLayoutSider,
  NLayoutHeader,
  NLayoutContent,
  NMenu,
  NIcon,
  NSpace,
  NSwitch,
  NSelect,
  NButton,
  NDropdown,
  type MenuOption,
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { usePrefsStore } from '@/stores/prefs'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const prefs = usePrefsStore()

function icon(name: string) {
  // Lightweight inline SVG icons (avoid heavy icon set dep for phase 0)
  const map: Record<string, string> = {
    dashboard:
      '<path d="M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z"/>',
    clients:
      '<path d="M21 4H3a1 1 0 0 0-1 1v12a1 1 0 0 0 1 1h7v2H7v2h10v-2h-3v-2h7a1 1 0 0 0 1-1V5a1 1 0 0 0-1-1zm-1 12H4V6h16v10z"/>',
    hosts:
      '<path d="M12 2a10 10 0 1 0 10 10A10.011 10.011 0 0 0 12 2zm6.93 9h-3.96a14 14 0 0 0-1.21-5.06A8.014 8.014 0 0 1 18.93 11zM12 4c.81 0 1.97 2.04 2.45 5h-4.9C10.03 6.04 11.19 4 12 4zM4.07 11A8.014 8.014 0 0 1 9.24 5.94 14 14 0 0 0 8.03 11zm0 2h3.96a14 14 0 0 0 1.21 5.06A8.014 8.014 0 0 1 4.07 13zM12 20c-.81 0-1.97-2.04-2.45-5h4.9C13.97 17.96 12.81 20 12 20zm2.76-1.94A14 14 0 0 0 15.97 13h3.96a8.014 8.014 0 0 1-5.17 5.06z"/>',
    tcp: '<path d="M5 3h14v4H5zM5 10h14v4H5zM5 17h14v4H5z"/>',
    udp: '<path d="M3 13h2v-2H3v2zm4 0h2v-2H7v2zm4 0h2v-2h-2v2zm4 0h2v-2h-2v2zm4 0h2v-2h-2v2z"/>',
    http: '<path d="M3 5h18v2H3zM3 11h18v2H3zM3 17h18v2H3z"/>',
    socks5: '<path d="M12 2 4 6v6c0 5 3.5 9.7 8 10 4.5-.3 8-5 8-10V6l-8-4z"/>',
    secret:
      '<path d="M12 1 3 5v6c0 5.6 3.8 10.7 9 12 5.2-1.3 9-6.4 9-12V5l-9-4zm0 11h7c-.5 4-3.4 7.7-7 9V12H5V6l7-3v9z"/>',
    p2p: '<path d="M3 7h6v2H3zM3 11h10v2H3zM3 15h6v2H3zM15 7h6v2h-6zM15 11h6v2h-6zM15 15h6v2h-6z"/>',
    file: '<path d="M14 2H6a2 2 0 0 0-2 2v16c0 1.1.9 2 2 2h12a2 2 0 0 0 2-2V8l-6-6zm4 18H6V4h7v5h5v11z"/>',
    global: '<path d="M19.43 12.98c.04-.32.07-.65.07-.98s-.03-.66-.07-.98l2.11-1.65a.49.49 0 0 0 .12-.61l-2-3.46a.5.5 0 0 0-.61-.22l-2.49 1a7.03 7.03 0 0 0-1.7-.98l-.38-2.65A.49.49 0 0 0 14 2h-4a.49.49 0 0 0-.49.42l-.38 2.65c-.61.25-1.18.58-1.7.98l-2.49-1a.5.5 0 0 0-.61.22l-2 3.46a.49.49 0 0 0 .12.61l2.11 1.65c-.04.32-.07.65-.07.98s.03.66.07.98l-2.11 1.65a.49.49 0 0 0-.12.61l2 3.46c.14.24.43.34.69.22l2.49-1c.52.4 1.09.73 1.7.98l.38 2.65c.04.24.25.42.49.42h4c.24 0 .45-.18.49-.42l.38-2.65c.61-.25 1.18-.58 1.7-.98l2.49 1c.26.12.55.02.69-.22l2-3.46a.49.49 0 0 0-.12-.61l-2.11-1.65zM12 15.5a3.5 3.5 0 1 1 0-7 3.5 3.5 0 0 1 0 7z"/>',
    tokens: '<path d="M12 1 3 5v6c0 5.5 3.8 10.7 9 12 5.2-1.3 9-6.5 9-12V5l-9-4zm-1 6h2v6h-2V7zm0 8h2v2h-2v-2z"/>',
  }
  return () =>
    h(NIcon, { size: 18 }, () =>
      h('svg', {
        viewBox: '0 0 24 24',
        width: 18,
        height: 18,
        fill: 'currentColor',
        innerHTML: map[name] ?? '',
      }),
    )
}

const menu = computed<MenuOption[]>(() => [
  {
    label: t('nav.dashboard'),
    key: 'dashboard',
    icon: icon('dashboard'),
  },
  {
    label: t('nav.clients'),
    key: 'clients',
    icon: icon('clients'),
  },
  {
    type: 'group',
    label: t('nav.tunnels'),
    key: 'g-tunnels',
    children: [
      { label: t('nav.hosts'), key: 'hosts', icon: icon('hosts') },
      { label: t('nav.tcp'), key: 'tunnels-tcp', icon: icon('tcp') },
      { label: t('nav.udp'), key: 'tunnels-udp', icon: icon('udp') },
      { label: t('nav.http'), key: 'tunnels-http', icon: icon('http') },
      { label: t('nav.socks5'), key: 'tunnels-socks5', icon: icon('socks5') },
      { label: t('nav.secret'), key: 'tunnels-secret', icon: icon('secret') },
      { label: t('nav.p2p'), key: 'tunnels-p2p', icon: icon('p2p') },
      { label: t('nav.file'), key: 'tunnels-file', icon: icon('file') },
    ],
  },
  {
    type: 'group',
    label: t('nav.system'),
    key: 'g-system',
    children: [
      { label: t('nav.global'), key: 'global', icon: icon('global') },
      { label: t('nav.tokens'), key: 'tokens', icon: icon('tokens') },
    ],
  },
])

const activeKey = computed(() => {
  const n = route.name as string | undefined
  if (n === 'tunnels') return 'tunnels-' + (route.params.mode as string)
  return n ?? ''
})

function onMenu(key: string) {
  if (key.startsWith('tunnels-')) {
    router.push({
      name: 'tunnels',
      params: { mode: key.slice('tunnels-'.length) },
    })
  } else {
    router.push({ name: key })
  }
}

const langOptions = [
  { label: '简体中文', value: 'zh-CN' },
  { label: 'English', value: 'en' },
]
function onLangChange(v: string) {
  locale.value = v as 'zh-CN' | 'en'
  prefs.setLang(v)
}

const userMenuOptions = [{ label: t('app.logout'), key: 'logout' }]
async function onUserMenu(key: string) {
  if (key === 'logout') {
    await auth.logout()
    router.replace({ name: 'login' })
  }
}
</script>

<template>
  <NLayout has-sider class="h-screen">
    <NLayoutSider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="240"
      show-trigger
      :collapsed="prefs.sidebarCollapsed"
      @update:collapsed="(v: boolean) => prefs.setSidebar(v)"
    >
      <div class="nps-brand flex items-center gap-3 px-4 h-14">
        <div class="nps-brand-logo">N</div>
        <span
          v-show="!prefs.sidebarCollapsed"
          class="font-semibold tracking-wide nps-gradient-text text-[15px]"
        >
          {{ t('app.title') }}
        </span>
      </div>
      <NMenu
        :collapsed="prefs.sidebarCollapsed"
        :collapsed-width="64"
        :collapsed-icon-size="20"
        :indent="16"
        :options="menu"
        :value="activeKey"
        @update:value="onMenu"
      />
    </NLayoutSider>

    <NLayout>
      <NLayoutHeader
        bordered
        class="px-5 h-14 flex items-center justify-between nps-header"
      >
        <span class="text-sm opacity-70 flex items-center gap-2">
          <span class="nps-status-dot" />
          {{ t('app.welcome') }}
          <span class="font-semibold nps-gradient-text">
            {{ auth.user?.username || '-' }}
          </span>
        </span>
        <NSpace align="center" :size="14">
          <NSwitch
            :value="prefs.dark"
            @update:value="(v: boolean) => prefs.setDark(v)"
          >
            <template #checked>🌙</template>
            <template #unchecked>☀️</template>
          </NSwitch>
          <NSelect
            :value="locale"
            :options="langOptions"
            size="small"
            style="width: 120px"
            @update:value="onLangChange"
          />
          <NDropdown :options="userMenuOptions" @select="onUserMenu">
            <NButton size="small" quaternary>
              {{ auth.user?.username || '-' }} ▾
            </NButton>
          </NDropdown>
        </NSpace>
      </NLayoutHeader>

      <NLayoutContent class="p-4">
        <RouterView />
      </NLayoutContent>
    </NLayout>
  </NLayout>
</template>

<style scoped>
.nps-brand {
  position: relative;
  border-bottom: 1px solid var(--n-border-color);
}
.nps-brand::after {
  content: '';
  position: absolute;
  left: 16px;
  right: 16px;
  bottom: -1px;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent,
    var(--nps-accent),
    var(--nps-accent-2),
    transparent
  );
  opacity: 0.55;
}
.nps-brand-logo {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  color: #fff;
  font-weight: 700;
  background: linear-gradient(135deg, #6366f1 0%, #06b6d4 100%);
  box-shadow:
    0 6px 18px rgba(99, 102, 241, 0.35),
    inset 0 1px 0 rgba(255, 255, 255, 0.25);
  letter-spacing: 0.5px;
}
.nps-header {
  position: relative;
  backdrop-filter: saturate(160%) blur(12px);
  -webkit-backdrop-filter: saturate(160%) blur(12px);
}
.nps-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: #22c55e;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.18);
  animation: nps-pulse 2.4s ease-in-out infinite;
}
@keyframes nps-pulse {
  0%,
  100% {
    box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.18);
  }
  50% {
    box-shadow: 0 0 0 6px rgba(34, 197, 94, 0.05);
  }
}
</style>
