<script setup lang="ts">
import { ref } from 'vue'
import type { UptimeCell, UptimeSeries } from '../../lib/models'
import { date, number } from '../../lib/api'
import { t } from '../../lib/i18n'
defineProps<{ series: UptimeSeries; label?: string; compact?: boolean }>()
const emit = defineEmits<{ select: [cell: UptimeCell] }>()
const description = (cell: UptimeCell) =>
  `${date(cell.at)} – ${date(cell.end)}\n${t('uptime_' + cell.status)}\n${t('availability')}: ${number(cell.availability, '%')} / ${t('coverage')}: ${number(cell.coverage, '%')}\n${t('success')}: ${cell.success}, ${t('failure')}: ${cell.failed}, ${t('error')}: ${cell.errors}`
const hover = ref<{ cell: UptimeCell; x: number; y: number }>()
function show(event: Event, cell: UptimeCell) {
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  const width = Math.min(280, window.innerWidth - 16)
  hover.value = {
    cell,
    x: Math.max(8, Math.min(rect.left, window.innerWidth - width - 8)),
    y: rect.top > 100 ? rect.top - 8 : rect.bottom + 8,
  }
  above.value = rect.top > 100
}
const above = ref(false)
</script>
<template>
  <div v-if="compact" class="preview">
    <span
      v-for="cell in series.cells.slice(-24)"
      :key="cell.at"
      :class="cell.status"
      @mouseenter="show($event, cell)"
      @mouseleave="hover = undefined"
    />
  </div>
  <div v-else class="uptime-history">
    <div class="uptime-stats">
      <span
        >{{ t('availability') }} <strong>{{ number(series.availability, '%') }}</strong></span
      ><span
        >{{ t('coverage') }} <strong>{{ number(series.coverage, '%') }}</strong></span
      ><small>{{ series.bucket_seconds / 60 }} {{ t('minutesPerCell') }}</small>
    </div>
    <div class="uptime-grid" role="group" :aria-label="label || t('uptimeHistory')">
      <button
        v-for="cell in series.cells"
        :key="cell.at"
        type="button"
        :class="cell.status"
        @mouseenter="show($event, cell)"
        @mouseleave="hover = undefined"
        :aria-label="description(cell)"
        @focus="show($event, cell)"
        @blur="hover = undefined"
        @keydown.esc="hover = undefined"
        @click="emit('select', cell)"
      />
    </div>
    <div class="uptime-axis">
      <small>{{ date(series.from) }}</small
      ><small>{{ date(series.until) }}</small>
    </div>
    <div class="uptime-legend">
      <span v-for="state in ['healthy', 'partial', 'down', 'incomplete', 'unknown']" :key="state"
        ><i :class="state" />{{ t('uptime_' + state) }}</span
      >
    </div>
  </div>
  <Teleport to="body">
    <div
      v-if="hover"
      class="bucket-tooltip"
      role="tooltip"
      :style="{
        left: hover.x + 'px',
        top: hover.y + 'px',
        transform: above ? 'translateY(-100%)' : undefined,
      }"
    >
      <strong>{{ date(hover.cell.at) }} – {{ date(hover.cell.end) }}</strong>
      <div>
        {{ t('uptime_' + hover.cell.status) }} · {{ t('coverage') }}
        {{ number(hover.cell.coverage, '%') }}
      </div>
    </div>
  </Teleport>
</template>
<style scoped>
.bucket-tooltip {
  position: fixed;
  pointer-events: none;
  z-index: 4000;
  box-sizing: border-box;
  width: min(280px, calc(100vw - 16px));
  padding: 8px 12px;
  border-radius: 5px;
  color: white;
  background: #17343b;
  font-size: 13px;
}
.bucket-tooltip strong {
  color: inherit;
  margin: 0;
}

.preview {
  display: flex;
  gap: 2px;
  margin: 6px 0;
  max-width: 240px;
}
.preview span {
  flex: 1;
  height: 12px;
  border-radius: 2px;
}
.uptime-stats,
.uptime-axis {
  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 12px;
}
.uptime-stats {
  margin-bottom: 14px;
  color: var(--muted);
}
strong {
  color: var(--ink);
  font-variant-numeric: tabular-nums;
  margin-left: 6px;
}
.uptime-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(7px, 1fr));
  gap: 3px;
}
.uptime-grid button {
  border: 0;
  border-radius: 2px;
  padding: 0;
  height: 28px;
  min-width: 0;
  transition: transform 0.12s;
}
.uptime-grid button:hover,
.uptime-grid button:focus-visible {
  transform: translateY(-3px);
  outline: 2px solid var(--ink);
  outline-offset: 1px;
}
.healthy {
  background: #19967e;
}
.partial {
  background: #d9a23d;
}
.down {
  background: #cb5753;
}
.incomplete {
  background: repeating-linear-gradient(135deg, #8bbeb1 0 3px, #dcebe5 3px 6px);
}
.unknown {
  background: #d9e1df;
}
.uptime-axis {
  margin-top: 10px;
}
.uptime-legend {
  display: flex;
  gap: 14px;
  flex-wrap: wrap;
  margin-top: 14px;
  color: var(--muted);
  font-size: 11px;
}
.uptime-legend span {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.uptime-legend i {
  width: 9px;
  height: 9px;
  border-radius: 2px;
}
@media (prefers-reduced-motion: reduce) {
  .uptime-grid button {
    transition: none;
  }
}
</style>
