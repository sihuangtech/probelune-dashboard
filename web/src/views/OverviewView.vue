<script setup lang="ts">
import { usePolling, date, number, canEdit } from '../lib/api'
import type { Overview } from '../lib/models'
import { t } from '../lib/i18n'
import HealthPill from '../components/HealthPill.vue'
const { data, error, updated } = usePolling<Overview>(() => '/overview', {
  monitors: [],
  nodes: [],
  incidents: [],
  summary: {},
  online_nodes: 0,
  time: 0,
})
</script>
<template>
  <div class="page-title">
    <div>
      <p class="eyebrow">{{ t('live') }}</p>
      <h1>{{ t('overview') }}</h1>
      <p class="muted">{{ t('tagline') }}</p>
    </div>
    <router-link v-if="canEdit" to="/monitors"
      ><el-button type="primary">+ {{ t('monitors') }}</el-button></router-link
    >
  </div>
  <el-alert
    v-if="error"
    :title="t('offlineCache') + ' · ' + error"
    type="warning"
    :closable="false"
  />
  <div class="stat-grid">
    <div class="stat-card primary">
      <span class="eyebrow">{{ t('healthy') }}</span>
      <div class="value">{{ data.summary.healthy || 0 }}</div>
      <small>● {{ t('live') }}</small>
    </div>
    <div class="stat-card">
      <span class="eyebrow">{{ t('total') }}</span>
      <div class="value">{{ data.monitors.length }}</div>
      <small>{{ t('monitors') }}</small>
    </div>
    <div class="stat-card">
      <span class="eyebrow">{{ t('failures') }}</span>
      <div class="value">{{ (data.summary.down || 0) + (data.summary.partial || 0) }}</div>
      <small>{{ t('incidents') }} · {{ data.incidents.length }}</small>
    </div>
    <div class="stat-card">
      <span class="eyebrow">{{ t('onlineNodes') }}</span>
      <div class="value">
        {{ data.online_nodes }}<small> / {{ data.nodes.length }}</small>
      </div>
      <small>{{ date(data.time) }}</small>
    </div>
  </div>
  <section class="panel">
    <div class="section-title">
      <h2>{{ t('matrix') }}</h2>
      <span class="eyebrow">{{ updated ? new Date(updated).toLocaleTimeString() : '—' }}</span>
    </div>
    <div v-if="!data.monitors.length" class="empty-state">
      <div class="empty-symbol">◉</div>
      <h3>{{ t('empty') }}</h3>
      <p>{{ t('noNodes') }}</p>
      <router-link to="/nodes"
        ><el-button>{{ t('nodes') }} →</el-button></router-link
      >
    </div>
    <div v-else class="matrix">
      <table>
        <thead>
          <tr>
            <th>{{ t('monitors') }}</th>
            <th>{{ t('overview') }}</th>
            <th v-for="node in data.nodes.filter((n) => !n.revoked)" :key="node.id">
              {{ node.name }}<br /><small>{{ node.region }}</small>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in data.monitors" :key="m.id">
            <td>
              <router-link :to="'/monitors/' + m.id"
                ><strong>{{ m.name }}</strong></router-link
              ><small>{{ m.check.type.toUpperCase() }} · {{ m.interval }}s</small>
            </td>
            <td>
              <HealthPill :status="m.status" /><small v-if="!m.coverage_complete">{{
                t('incomplete')
              }}</small>
            </td>
            <td v-for="node in data.nodes.filter((n) => !n.revoked)" :key="node.id">
              <template v-if="m.node_ids.includes(node.id)"
                ><HealthPill
                  :status="m.states.find((s) => s.node_id === node.id)?.status || 'unknown'"
                /><small>{{
                  number(m.states.find((s) => s.node_id === node.id)?.result?.latency_ms, ' ms')
                }}</small></template
              ><span v-else class="muted">—</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div class="mobile-only stack">
      <router-link
        v-for="m in data.monitors"
        :key="m.id"
        :to="'/monitors/' + m.id"
        class="node-card"
        ><div class="section-title">
          <strong>{{ m.name }}</strong
          ><HealthPill :status="m.status" />
        </div>
        <p v-for="state in m.states" :key="state.node_id">
          {{ data.nodes.find((n) => n.id === state.node_id)?.name }}
          <HealthPill :status="state.status" /></p
      ></router-link>
    </div>
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
      ><span
        >{{ incident.name }}<small> · {{ date(incident.opened_at) }}</small></span
      ><HealthPill :status="incident.node_id ? 'offline' : 'failed'"
    /></router-link>
  </section>
</template>
