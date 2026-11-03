<script setup lang="ts">
import { computed } from 'vue'
import { usePolling, date } from '../lib/api'
import type { Monitor, Overview } from '../lib/models'
import { t } from '../lib/i18n'
import HealthPill from '../components/HealthPill.vue'
import { endpoint } from '../lib/endpoint'
import { checkTypeLabel } from '../features/checks/catalog'
const { data, error } = usePolling<Overview>(() => '/overview', {
  monitors: [],
  nodes: [],
  incidents: [],
  summary: {},
  online_nodes: 0,
  time: 0,
})
const targets = computed(() =>
  data.value.nodes.filter((node) => node.role !== 'panel' && !node.revoked),
)
const online = computed(
  () =>
    targets.value.filter((node) => node.enabled && data.value.time - node.last_seen <= 60).length,
)
const panel = computed(() =>
  data.value.nodes.find((node) => node.role === 'panel' && !node.revoked),
)
const active = computed(() => data.value.monitors.filter((m) => !m.archived))
function groupHealth(checks: Monitor[]) {
  const live = checks.filter((m) => m.enabled)
  if (!live.length) return checks.length ? 'paused' : 'noChecks'
  if (live.every((m) => m.status === 'down')) return 'down'
  if (live.some((m) => ['down', 'partial'].includes(m.status))) return 'mixedHealth'
  return live.every((m) => m.status === 'healthy' && m.coverage_complete) ? 'healthy' : 'unknown'
}
// Names are an association only when unique; executor IDs identify the source, not the target.
const servers = computed(() =>
  targets.value.map((node) => {
    const checks =
      targets.value.filter((n) => n.name === node.name).length === 1
        ? active.value.filter((m) => m.name === node.name)
        : []
    return { node, checks, status: groupHealth(checks) }
  }),
)
const remaining = computed(() => {
  const assigned = new Set(servers.value.flatMap((s) => s.checks.map((m) => m.id)))
  return active.value.filter((m) => !assigned.has(m.id))
})
function panelCheck(m: Monitor) {
  try {
    return m.check.type === 'http' && new URL(m.check.target).host === location.host
  } catch {
    return false
  }
}
const sections = computed(() => [
  { key: 'panelHealth', checks: remaining.value.filter(panelCheck) },
  { key: 'otherTargets', checks: remaining.value.filter((m) => !panelCheck(m)) },
])
</script>

<template>
  <div class="page-title">
    <div>
      <p class="eyebrow">{{ t('live') }}</p>
      <h1>{{ t('overview') }}</h1>
      <p v-if="panel" class="muted">{{ t('probeSource') }}：{{ panel.name }}</p>
    </div>
    <router-link to="/monitors">{{ t('monitors') }} →</router-link>
  </div>
  <el-alert
    v-if="error"
    :title="t('offlineCache') + ' - ' + error"
    type="warning"
    :closable="false"
  />
  <section class="panel">
    <div class="section-title">
      <h2>{{ t('serverHealth') }}</h2>
      <span>{{ t('onlineNodes') }} {{ online }} / {{ targets.length }}</span>
    </div>
    <article v-for="s in servers" :key="s.node.id" class="server-row">
      <div>
        <strong>{{ s.node.name }}</strong>
        <p class="muted">
          {{ t('heartbeat') }}：{{
            t(s.node.enabled && data.time - s.node.last_seen <= 60 ? 'online' : 'offline')
          }}
          · {{ date(s.node.last_seen) }}
        </p>
      </div>
      <HealthPill :status="s.status" />
      <div class="check-links">
        <router-link
          v-for="m in s.checks"
          :key="m.id"
          :to="'/monitors/' + m.id"
          :title="endpoint(m.check)"
          >{{ checkTypeLabel(m.check.type) }} {{ t(m.status) }} →</router-link
        >
      </div>
    </article>
    <el-empty v-if="!servers.length" />
  </section>
  <template v-for="section in sections" :key="section.key">
    <section v-if="section.checks.length" class="panel">
      <h2>{{ t(section.key) }}</h2>
      <article v-for="m in section.checks" :key="m.id" class="server-row">
        <router-link :to="'/monitors/' + m.id">{{ m.name }} · {{ endpoint(m.check) }}</router-link
        ><HealthPill :status="m.status" />
      </article>
    </section>
  </template>
  <section class="panel">
    <div class="section-title">
      <h3>{{ t('incidents') }}</h3>
      <router-link to="/incidents">{{ t('view') }} →</router-link>
    </div>
    <p v-if="!data.incidents.length" class="muted">{{ t('noIncident') }}</p>
    <router-link
      v-for="incident in data.incidents"
      :key="incident.id"
      :to="'/incidents?id=' + incident.id"
      class="incident-line"
    >
      <span
        >{{ incident.name }}<small> - {{ date(incident.opened_at) }}</small></span
      ><HealthPill :status="incident.node_id ? 'offline' : 'failed'" />
    </router-link>
  </section>
</template>

<style scoped>
.server-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px 24px;
  padding: 16px 0;
  border-top: 1px solid var(--el-border-color-light);
  overflow-wrap: anywhere;
}
.server-row p {
  margin: 4px 0 0;
}
.check-links {
  margin-inline-start: auto;
  display: flex;
  gap: 8px 20px;
  flex-wrap: wrap;
}
@media (max-width: 700px) {
  .check-links {
    width: 100%;
    margin: 0;
  }
}
</style>
