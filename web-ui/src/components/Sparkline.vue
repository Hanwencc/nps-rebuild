<script setup lang="ts">
import { computed } from 'vue'

/**
 * Lightweight inline-SVG sparkline. Zero dependencies — keeps the
 * dashboard bundle small while giving us a clean "tech" look.
 *
 * - `data` : numeric series (oldest -> newest).
 * - `min`/`max` : optional clamp; otherwise auto from data.
 * - `color` : stroke color (CSS).
 * - `gradient` : second color for the area-fill gradient (top stop).
 * - `height` : px; width is responsive (viewBox preserveAspectRatio=none).
 */
const props = withDefaults(
  defineProps<{
    data: number[]
    min?: number
    max?: number
    color?: string
    gradient?: string
    height?: number
    smooth?: boolean
  }>(),
  {
    color: 'var(--nps-accent, #6366f1)',
    gradient: 'var(--nps-accent-2, #06b6d4)',
    height: 56,
    smooth: true,
  },
)

const W = 100 // viewBox width units; SVG scales to container
const gid = 'spark-' + Math.random().toString(36).slice(2, 9)

const path = computed(() => {
  const d = props.data
  if (!d || d.length === 0) return { line: '', area: '' }
  const lo = props.min ?? Math.min(...d)
  const hiRaw = props.max ?? Math.max(...d)
  const hi = hiRaw === lo ? lo + 1 : hiRaw
  const n = d.length
  const step = n > 1 ? W / (n - 1) : 0
  const pts = d.map((v, i) => {
    const x = i * step
    const y = 100 - ((v - lo) / (hi - lo)) * 100
    return [x, y] as [number, number]
  })

  let line = `M ${pts[0][0]} ${pts[0][1]}`
  if (props.smooth && pts.length > 1) {
    for (let i = 1; i < pts.length; i++) {
      const [x0, y0] = pts[i - 1]
      const [x1, y1] = pts[i]
      const cx = (x0 + x1) / 2
      line += ` C ${cx} ${y0}, ${cx} ${y1}, ${x1} ${y1}`
    }
  } else {
    for (let i = 1; i < pts.length; i++) {
      line += ` L ${pts[i][0]} ${pts[i][1]}`
    }
  }
  const area = `${line} L ${W} 100 L 0 100 Z`
  return { line, area }
})

const lastPoint = computed(() => {
  const d = props.data
  if (!d || d.length === 0) return null
  const lo = props.min ?? Math.min(...d)
  const hiRaw = props.max ?? Math.max(...d)
  const hi = hiRaw === lo ? lo + 1 : hiRaw
  const n = d.length
  const step = n > 1 ? W / (n - 1) : 0
  const x = (n - 1) * step
  const y = 100 - ((d[n - 1] - lo) / (hi - lo)) * 100
  return { x, y }
})
</script>

<template>
  <div class="nps-spark" :style="{ height: height + 'px' }">
    <svg
      :viewBox="`0 0 ${W} 100`"
      preserveAspectRatio="none"
      width="100%"
      :height="height"
      aria-hidden="true"
    >
      <defs>
        <linearGradient :id="gid" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" :stop-color="gradient" stop-opacity="0.45" />
          <stop offset="100%" :stop-color="gradient" stop-opacity="0" />
        </linearGradient>
      </defs>
      <path
        v-if="path.area"
        :d="path.area"
        :fill="`url(#${gid})`"
        stroke="none"
      />
      <path
        v-if="path.line"
        :d="path.line"
        fill="none"
        :stroke="color"
        stroke-width="1.6"
        stroke-linecap="round"
        stroke-linejoin="round"
        vector-effect="non-scaling-stroke"
      />
      <circle
        v-if="lastPoint"
        :cx="lastPoint.x"
        :cy="lastPoint.y"
        r="1.8"
        :fill="color"
        vector-effect="non-scaling-stroke"
      />
    </svg>
  </div>
</template>

<style scoped>
.nps-spark {
  width: 100%;
  display: block;
}
.nps-spark svg {
  display: block;
}
</style>
