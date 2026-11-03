<script setup lang="ts">
import { computed } from 'vue'
import { usePolling, date, number, canEdit } from '../lib/api'
import type { Monitor, Overview } from '../lib/models'
import { t } from '../lib/i18n'
import HealthPill from '../components/HealthPill.vue'
import MonitorTypeFilter from '../features/monitors/MonitorTypeFilter.vue'
import { monitorType } from '../features/monitors/typeFilter'
const { data, error, updated } = usePolling<Overview>(() => '/overview', {
  monitors: [],
  nodes: [],
  incidents: [],
  summary: {},
  online_nodes: 0,
  time: 0,
})
const visible = computed(() =>
  data.value.monitors.filter(
    (monitor) => !monitorType.value || monitor.check.type === monitorType.value,
  ),
)
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
const summary = computed(() => {
  const counts: Record<string, number> = {}
  for (const monitor of visible.value) counts[monitor.status] = (counts[monitor.status] || 0) + 1
  return counts
})
function latest(monitor: Monitor) {
  return monitor.states.reduce<(typeof monitor.states)[number] | undefined>(
    (last, state) => (!last || state.last_at > last.last_at ? state : last),
    undefined,
  )
}
</script>

<template>
  <div class="page-title">
    <div>
      <p class="eyebrow">{{ t('live') }}</p>
      <h1>{{ t('overview') }}</h1>
      <p class="muted">{{ t('panelOverviewHint') }}</p>
    </div>
    <router-link v-if="canEdit" to="/monitors"
      ><el-button type="primary">+ {{ t('monitors') }}</el-button></router-link
    >
  </div>
  <el-alert
    v-if="error"
    :title="t('offlineCache') + ' - ' + error"
    type="warning"
    :closable="false"
  />
  <section class="panel">
    <MonitorTypeFilter :monitors="data.monitors" />
    <div class="stat-grid">
      <div class="stat-card primary">
        <span class="eyebrow">{{ t('healthy') }}</span>
        <div class="value">{{ summary.healthy || 0 }}</div>
        <small>{{ t('filteredMonitors') }}</small>
      </div>
      <div class="stat-card">
        <span class="eyebrow">{{ t('total') }}</span>
        <div class="value">{{ visible.length }}</div>
        <small>{{ t('filteredMonitors') }}</small>
      </div>
      <div class="stat-card">
        <span class="eyebrow">{{ t('failures') }}</span>
        <div class="value">{{ (summary.down || 0) + (summary.partial || 0) }}</div>
        <small>{{ t('filteredMonitors') }}</small>
      </div>
      <div class="stat-card">
        <span class="eyebrow">{{ t('onlineNodes') }}</span>
        <div class="value">
          {{ online }}<small> / {{ targets.length }}</small>
        </div>
        <small>{{ date(data.time) }}</small>
      </div>
    </div>
    <div class="section-title">
      <div>
        <h2>{{ t('targetResults') }}</h2>
        <p v-if="panel" class="muted">{{ t('probeSource') }}：{{ panel.name }}</p>
      </div>
      <span class="eyebrow">{{ updated ? new Date(updated).toLocaleTimeString() : '—' }}</span>
    </div>
    <el-empty v-if="!visible.length" :description="t('noTypeMonitors')" />
    <el-table v-else :data="visible">
      <el-table-column :label="t('name')" min-width="210"
        ><template #default="{ row }">
          <router-link :to="'/monitors/' + row.id"
            ><strong>{{ row.name }}</strong></router-link
          >
          <div class="muted">
            {{ row.check.target
            }}<template v-if="['tcp', 'udp'].includes(row.check.type)">
              : {{ row.check.port }}</template
            >
          </div>
        </template></el-table-column
      >
      <el-table-column :label="t('overview')" min-width="160"
        ><template #default="{ row }">
          <HealthPill :status="row.status" /><small v-if="!row.coverage_complete">
            {{ t('incomplete') }}</small
          >
        </template></el-table-column
      >
      <el-table-column :label="t('latency')" min-width="110"
        ><template #default="{ row }">{{
          number(latest(row)?.result?.latency_ms, ' ms')
        }}</template></el-table-column
      >
      <el-table-column :label="t('updated')" min-width="175"
        ><template #default="{ row }">{{
          date(latest(row)?.last_at || 0)
        }}</template></el-table-column
      >
      <el-table-column :label="t('interval')" width="110"
        ><template #default="{ row }">{{ row.interval }}s</template></el-table-column
      >
      <el-table-column width="85"
        ><template #default="{ row }"
          ><router-link :to="'/monitors/' + row.id">{{ t('details') }} →</router-link></template
        ></el-table-column
      >
    </el-table>
  </section>
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
