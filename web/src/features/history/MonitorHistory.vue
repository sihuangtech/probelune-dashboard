<script setup lang="ts">
import { ref, watch } from 'vue'
import { usePolling, date, number } from '../../lib/api'
import { t } from '../../lib/i18n'
import { loadUptimeRange, uptimeQuery } from './range'
import type { History, Node, Result, UptimeCell, UptimeSeries } from '../../lib/models'
import HistoryRange from './HistoryRange.vue'
import UptimeStrip from './UptimeStrip.vue'
import HistoryChart from './HistoryChart.vue'
import HealthPill from '../../components/HealthPill.vue'
import ResultCard from '../checks/ResultCard.vue'
const props = defineProps<{ monitorId: string; nodes: Node[] }>()
const range = ref(loadUptimeRange()),
  node = ref(''),
  status = ref(''),
  stage = ref(''),
  page = ref(1)
const selected = ref<UptimeCell>(),
  evidence = ref<Result>()
const {
  data: uptime,
  error: uptimeError,
  reload: reloadUptime,
} = usePolling<UptimeSeries | null>(
  () => `/monitors/${props.monitorId}/uptime?${uptimeQuery(range.value)}`,
  null,
  30000,
)
const {
  data: history,
  error,
  reload,
} = usePolling<History>(
  () => {
    const query = new URLSearchParams({
      hours: String(range.value.hours),
      node_id: node.value,
      status: status.value,
      stage: stage.value,
      limit: '50',
      offset: String((page.value - 1) * 50),
    })
    if (selected.value) {
      query.set('from', String(selected.value.at))
      query.set('until', String(Math.min(selected.value.end - 1, Math.floor(Date.now() / 1000))))
    }
    return `/monitors/${props.monitorId}/history?${query}`
  },
  { results: [], summary: [], coverage_basis: '', total: 0 },
  15000,
)
watch([node, status, stage, selected], () => {
  page.value = 1
  void reload()
})
watch(page, () => void reload())
watch(range, () => {
  selected.value = undefined
  page.value = 1
  void reload()
  void reloadUptime()
})
</script>
<template>
  <section class="panel">
    <div class="section-title">
      <h2>{{ t('uptimeHistory') }}</h2>
    </div>
    <p class="muted">{{ t('uptimeHint') }}</p>
    <HistoryRange v-model="range" /><el-alert
      v-if="uptimeError"
      :title="uptimeError"
      type="warning"
      :closable="false"
    /><UptimeStrip v-if="uptime" :series="uptime" @select="(cell) => (selected = cell)" />
  </section>
  <section class="panel">
    <div class="section-title">
      <h3>{{ t('history') }}</h3>
      <el-button v-if="selected" text @click="selected = undefined">{{
        t('clearRange')
      }}</el-button>
    </div>
    <p v-if="selected" class="mono">{{ date(selected.at) }} — {{ date(selected.end) }}</p>
    <div class="history-filters">
      <el-select v-model="node" clearable :placeholder="t('allNodes')"
        ><el-option v-for="n in nodes" :key="n.id" :value="n.id" :label="n.name" /></el-select
      ><el-select v-model="status" clearable :placeholder="t('allResults')"
        ><el-option
          v-for="value in ['success', 'failure', 'error', 'cancelled']"
          :key="value"
          :value="value"
          :label="t(value)" /></el-select
      ><el-select v-model="stage" clearable :placeholder="t('allStages')"
        ><el-option
          v-for="value in [
            'dns',
            'connect',
            'tls',
            'http',
            'assertion',
            'timeout',
            'policy',
            'executor',
            'response',
            'complete',
          ]"
          :key="value"
          :value="value"
          :label="value"
      /></el-select>
    </div>
    <el-alert v-if="error" :title="error" type="warning" :closable="false" />
    <HistoryChart
      :buckets="history.buckets || []"
      :width="history.bucket_seconds || 300"
      :nodes="nodes"
    />
    <el-table :data="history.summary"
      ><el-table-column prop="node_id" :label="t('location')"
        ><template #default="{ row }">{{
          nodes.find((n) => n.id === row.node_id)?.name || row.node_id
        }}</template></el-table-column
      ><el-table-column :label="t('availability')"
        ><template #default="{ row }">{{
          number(row.availability, '%')
        }}</template></el-table-column
      ><el-table-column :label="t('coverage')"
        ><template #default="{ row }">{{ number(row.coverage, '%') }}</template></el-table-column
      ></el-table
    >
    <p class="muted">
      <small>{{ t('historySummaryHint') }}</small>
    </p>
    <el-table :data="history.results"
      ><el-table-column :label="t('updated')" min-width="170"
        ><template #default="{ row }">{{ date(row.scheduled_at) }}</template></el-table-column
      ><el-table-column :label="t('location')"
        ><template #default="{ row }">{{
          nodes.find((n) => n.id === row.node_id)?.name || row.node_id
        }}</template></el-table-column
      ><el-table-column :label="t('overview')"
        ><template #default="{ row }"
          ><HealthPill :status="row.status" /></template></el-table-column
      ><el-table-column prop="stage" :label="t('stage')" /><el-table-column :label="t('latency')"
        ><template #default="{ row }">{{
          number(row.latency_ms, ' ms')
        }}</template></el-table-column
      ><el-table-column
        ><template #default="{ row }"
          ><el-button text @click="evidence = row">{{ t('evidence') }}</el-button></template
        ></el-table-column
      ></el-table
    >
    <el-pagination
      v-model:current-page="page"
      :page-size="50"
      :total="history.total || 0"
      layout="prev, pager, next, total"
      style="margin-top: 20px; flex-wrap: wrap"
    />
    <p v-if="!history.total" class="muted">{{ t('rawHistoryHint') }}</p>
  </section>
  <el-dialog
    :model-value="!!evidence"
    @close="evidence = undefined"
    width="700px"
    :title="t('evidence')"
    ><ResultCard v-if="evidence" :result="evidence"
  /></el-dialog>
</template>
<style scoped>
.history-filters {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 18px;
}
.history-filters .el-select {
  width: 180px;
}
</style>
