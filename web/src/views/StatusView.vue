<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { usePolling, date, number, displayTimezone } from '../lib/api'
import { t } from '../lib/i18n'
import { loadUptimeRange, uptimeQuery } from '../features/history/range'
import type { UptimeCell, UptimeSeries } from '../lib/models'
import HealthPill from '../components/HealthPill.vue'
import HistoryRange from '../features/history/HistoryRange.vue'
import UptimeStrip from '../features/history/UptimeStrip.vue'
interface PublicMonitor {
  id: string
  name: string
  status: string
  maintenance: boolean
  coverage_complete: boolean
  interval: number
  uptime: UptimeSeries
}
interface PublicIncident {
  monitor_id: string
  opened_at: number
  recovered_at: number
  suppressed: boolean
}
const range = ref(loadUptimeRange())
const selected = ref<{ monitor: PublicMonitor; cell: UptimeCell }>()
const { data, error, updated, reload } = usePolling<{
  name: string
  timezone?: string
  monitors: PublicMonitor[]
  incidents: PublicIncident[]
}>(
  () => `/status?${uptimeQuery(range.value)}`,
  { name: 'Probelune', monitors: [], incidents: [] },
  30000,
)
watch(range, () => {
  selected.value = undefined
  void reload()
})
watch(
  () => data.value.timezone,
  (value) => {
    if (value) displayTimezone.value = value
  },
)
const operational = computed(
  () =>
    data.value.monitors.length > 0 &&
    data.value.monitors.every(
      (m) => m.status === 'healthy' && m.coverage_complete && !m.maintenance,
    ),
)
const events = computed(() =>
  selected.value
    ? data.value.incidents.filter(
        (v) =>
          v.monitor_id === selected.value!.monitor.id &&
          v.opened_at < selected.value!.cell.end &&
          (!v.recovered_at || v.recovered_at >= selected.value!.cell.at),
      )
    : [],
)
</script>
<template>
  <main class="public-status">
    <header class="status-heading">
      <div>
        <p class="eyebrow">{{ t('statusEyebrow') }}</p>
        <h1>{{ data.name }}</h1>
      </div>
      <span class="status-mark">◉</span>
    </header>
    <div class="status-summary" :class="{ operational: operational && !error }" role="status">
      <span class="status-dot" /><strong>{{
        error
          ? t('offlineCache')
          : !data.monitors.length
            ? t('noPublicMonitors')
            : operational
              ? t('allOperational')
              : t('attentionNeeded')
      }}</strong>
      <small>{{ t('updated') }} {{ date(updated / 1000) }}</small>
    </div>
    <el-alert v-if="error" :title="error" type="warning" :closable="false" />
    <section class="status-controls">
      <h2>{{ t('uptimeHistory') }}</h2>
      <p class="muted">{{ t('uptimeHint') }}</p>
      <HistoryRange v-model="range" />
    </section>
    <section v-if="!data.monitors.length" class="panel">
      <el-empty :description="t('noPublicMonitors')" />
    </section>
    <article v-for="m in data.monitors" :key="m.id" class="panel service-history">
      <div class="section-title">
        <div>
          <h3>{{ m.name }}</h3>
          <small>{{ t('probeEvery') }} {{ m.interval }} {{ t('seconds') }}</small>
        </div>
        <HealthPill :status="m.maintenance ? 'maintenance' : m.status" />
      </div>
      <UptimeStrip
        :series="m.uptime"
        :label="m.name"
        @select="(cell) => (selected = { monitor: m, cell })"
      />
    </article>
    <footer>
      <span>{{ t('tagline') }}</span
      ><router-link to="/login">{{ t('login') }} ↗</router-link>
    </footer>
    <el-dialog
      :model-value="!!selected"
      @close="selected = undefined"
      :title="selected?.monitor.name"
      width="580px"
    >
      <template v-if="selected"
        ><p class="mono">{{ date(selected.cell.at) }} — {{ date(selected.cell.end) }}</p>
        <dl class="result-meta">
          <dt>{{ t('availability') }}</dt>
          <dd>{{ number(selected.cell.availability, '%') }}</dd>
          <dt>{{ t('coverage') }}</dt>
          <dd>{{ number(selected.cell.coverage, '%') }}</dd>
          <dt>{{ t('sampleCounts') }}</dt>
          <dd>
            {{ selected.cell.success }} / {{ selected.cell.failed }} / {{ selected.cell.errors }}
          </dd>
          <dt>{{ t('expectedSamples') }}</dt>
          <dd>{{ selected.cell.expected }}</dd>
        </dl>
        <h3>{{ t('incidents') }}</h3>
        <p v-if="!events.length" class="muted">{{ t('noConfirmedIncident') }}</p>
        <p v-for="event in events" :key="event.opened_at">
          {{ date(event.opened_at) }} —
          {{ event.recovered_at ? date(event.recovered_at) : t('active') }}
          <small v-if="event.suppressed">{{ t('suppressed') }}</small>
        </p></template
      >
    </el-dialog>
  </main>
</template>
<style scoped>
.public-status {
  max-width: 1000px;
  margin: 64px auto;
  padding: 0 28px;
}
.status-heading {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 28px;
  border-bottom: 1px solid var(--line);
}
.status-heading h1 {
  margin-bottom: 0;
  font-size: clamp(26px, 5vw, 42px);
}
.status-mark {
  color: var(--teal);
  font-size: 42px;
}
.status-summary {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  background: #edf0ed;
  padding: 22px;
  margin: 28px 0;
  border-radius: 9px;
}
.status-summary.operational {
  background: #e1f1e9;
  color: #08694f;
}
.status-summary small {
  margin-left: auto;
}
.status-controls {
  margin: 32px 0 24px;
}
.status-controls h2 {
  margin-bottom: 6px;
}
.service-history {
  padding: 28px;
}
footer {
  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  padding: 10px 0 32px;
  color: var(--muted);
  font-size: 12px;
}
@media (max-width: 600px) {
  .public-status {
    margin: 28px auto;
    padding: 0 16px;
  }
  .service-history {
    padding: 18px;
  }
  .status-summary small {
    width: 100%;
    margin: 0;
  }
}
</style>
