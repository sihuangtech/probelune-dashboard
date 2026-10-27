<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, usePolling, number, date, canEdit } from '../lib/api'
import type { Monitor, Node, History, Result } from '../lib/models'
import { t } from '../lib/i18n'
import HealthPill from '../components/HealthPill.vue'
import MonitorEditor from '../components/MonitorEditor.vue'
import HistoryChart from '../components/HistoryChart.vue'
import ResultCard from '../components/ResultCard.vue'
const route = useRoute(),
  router = useRouter(),
  hours = ref(24),
  editor = ref(false),
  evidence = ref<Result>()
const { data: monitors, reload } = usePolling<Monitor[]>(() => '/monitors', [])
const { data: nodes } = usePolling<Node[]>(() => '/nodes', [])
const {
  data: history,
  error,
  reload: reloadHistory,
} = usePolling<History>(() => `/monitors/${route.params.id}/history?hours=${hours.value}`, {
  results: [],
  summary: [],
  coverage_basis: '',
})
const monitor = computed(() => monitors.value.find((m) => m.id === route.params.id))
watch(hours, reloadHistory)
async function recheck() {
  try {
    await api('/diagnostics', 'POST', { monitor_id: monitor.value?.id })
    await router.push('/diagnostics')
  } catch (e) {
    ElMessage.error(String(e))
  }
}
</script>
<template>
  <div class="page-title">
    <div>
      <router-link to="/monitors">← {{ t('monitors') }}</router-link>
      <h1 style="margin-top: 12px">{{ monitor?.name }}</h1>
      <p class="muted mono">{{ monitor?.check.target }}</p>
    </div>
    <div class="actions" v-if="canEdit">
      <el-button @click="editor = true">{{ t('edit') }}</el-button
      ><el-button type="primary" @click="recheck">{{ t('recheck') }}</el-button>
    </div>
  </div>
  <el-alert v-if="error" :title="error" type="warning" :closable="false" />
  <div class="panel" v-if="monitor">
    <div class="section-title">
      <HealthPill :status="monitor.maintenance ? 'maintenance' : monitor.status" /><small
        v-if="!monitor.coverage_complete"
        >{{ t('incomplete') }}</small
      ><el-select v-model="hours" style="width: 130px"
        ><el-option :value="1" label="1h" /><el-option :value="24" label="24h" /><el-option
          :value="720"
          label="30d" /><el-option :value="2160" label="90d" /><el-option
          :value="8760"
          label="365d" /><el-option :value="168" label="7d"
      /></el-select>
    </div>
    <HistoryChart
      v-if="history.buckets?.length"
      :buckets="history.buckets"
      :width="history.bucket_seconds || 300"
      :nodes="nodes"
    /><el-empty v-else :description="t('noSample')" />
    <el-table :data="history.summary"
      ><el-table-column :label="t('location')"
        ><template #default="{ row }">{{
          nodes.find((n) => n.id === row.node_id)?.name || row.node_id
        }}</template></el-table-column
      ><el-table-column :label="t('availability')"
        ><template #default="{ row }">{{
          number(row.availability, '%')
        }}</template></el-table-column
      ><el-table-column :label="t('coverage')"
        ><template #default="{ row }">{{ number(row.coverage, '%') }}</template></el-table-column
      ><el-table-column :label="t('latency')"
        ><template #default="{ row }">{{
          number(row.latency_ms, ' ms')
        }}</template></el-table-column
      ></el-table
    ><small>{{ history.coverage_basis }}</small>
  </div>
  <div class="grid-two" v-if="monitor">
    <ResultCard
      v-for="state in monitor.states.filter((s) => s.result?.id)"
      :key="state.node_id"
      :result="state.result"
      :label="nodes.find((n) => n.id === state.node_id)?.name"
    />
  </div>
  <section class="panel" style="margin-top: 22px">
    <h3>{{ t('history') }}</h3>
    <el-table :data="history.results"
      ><el-table-column :label="t('updated')" min-width="170"
        ><template #default="{ row }">{{ date(row.scheduled_at) }}</template></el-table-column
      ><el-table-column :label="t('location')"
        ><template #default="{ row }">{{
          nodes.find((n) => n.id === row.node_id)?.name
        }}</template></el-table-column
      ><el-table-column :label="t('overview')"
        ><template #default="{ row }"
          ><HealthPill :status="row.status" /></template></el-table-column
      ><el-table-column :label="t('latency')"
        ><template #default="{ row }">{{
          number(row.latency_ms, ' ms')
        }}</template></el-table-column
      ><el-table-column
        ><template #default="{ row }"
          ><el-button text @click="evidence = row">{{ t('evidence') }}</el-button></template
        ></el-table-column
      ></el-table
    >
  </section>
  <MonitorEditor
    v-if="monitor"
    v-model="editor"
    :monitor="monitor"
    :nodes="nodes"
    @saved="reload"
  /><el-dialog
    :model-value="!!evidence"
    @close="evidence = undefined"
    width="700px"
    :title="t('evidence')"
    ><ResultCard v-if="evidence" :result="evidence"
  /></el-dialog>
</template>
