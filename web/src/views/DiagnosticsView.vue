<script setup lang="ts">
import { endpoint } from '../lib/endpoint'
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, canEdit, usePolling, download, date } from '../lib/api'
import type { Node, Diagnostic, Job } from '../lib/models'
import { defaultCheck } from '../lib/models'
import { t } from '../lib/i18n'
import CheckFields from '../features/checks/CheckFields.vue'
import { supportsCheck } from '../features/checks/catalog'
import ResultCard from '../features/checks/ResultCard.vue'
import HealthPill from '../components/HealthPill.vue'
const check = ref(defaultCheck()),
  nodeIDs = ref<string[]>([]),
  busy = ref(false),
  failure = ref('')
const route = useRoute(),
  router = useRouter()
const query = computed(() =>
  new URLSearchParams(
    Object.entries(route.query).filter(
      ([key, value]) => ['batch_id', 'monitor_id'].includes(key) && typeof value === 'string',
    ) as [string, string][],
  ).toString(),
)
const { data: nodes } = usePolling<Node[]>(() => '/nodes', [])
const panel = computed(() => nodes.value.find((node) => node.role === 'panel' && !node.revoked))
const {
  data: jobs,
  error,
  reload,
} = usePolling<Diagnostic[]>(() => '/diagnostics?' + query.value, [])
watch(query, () => void reload())
watch(
  () => check.value.type,
  () => {
    nodeIDs.value = nodeIDs.value.filter((id) => {
      const node = nodes.value.find((n) => n.id === id)
      return !!node && supportsCheck(node, check.value)
    })
  },
)
async function submit() {
  busy.value = true
  failure.value = ''
  try {
    const created = await api<Job[]>('/diagnostics', 'POST', {
      check: check.value,
      node_ids: nodeIDs.value,
    })
    await router.replace({ path: '/diagnostics', query: { batch_id: created[0]?.batch_id } })
    await reload()
  } catch (e) {
    failure.value = String(e)
  } finally {
    busy.value = false
  }
}
async function cancel(id: string) {
  try {
    await api('/diagnostics/' + id + '/cancel', 'POST', {})
    await reload()
  } catch (e) {
    failure.value = String(e)
  }
}
function exportData(format: string) {
  // 导出只包含结果和节点 ID，不带请求头、目标 URL 或原始输出。
  const rows = jobs.value.map(({ job, result }) => ({
    id: job.id,
    node_id: job.node_id,
    type: job.check.type,
    status: job.status,
    latency_ms: result?.latency_ms,
    stage: result?.stage,
    started_at: result?.started_at,
  }))
  if (format === 'json') download('probelune-diagnostics.json', JSON.stringify(rows, null, 2))
  else
    download(
      'probelune-diagnostics.csv',
      'id,node_id,type,status,latency_ms,stage,started_at\n' +
        rows
          .map((r) =>
            Object.values(r)
              .map((v) => '"' + String(v ?? '').replaceAll('"', '""') + '"')
              .join(','),
          )
          .join('\n'),
      'text/csv',
    )
}
</script>
<template>
  <div class="page-title">
    <div>
      <p class="eyebrow">ON-DEMAND / MULTI-LOCATION</p>
      <h1>{{ t('diagnostics') }}</h1>
      <p class="muted">{{ t('diagnosisHint') }}</p>
    </div>
    <div class="actions">
      <el-button v-if="query" @click="router.replace('/diagnostics')">{{ t('allRuns') }}</el-button>
      <el-button @click="exportData('json')">JSON ↓</el-button
      ><el-button @click="exportData('csv')">CSV ↓</el-button>
    </div>
  </div>
  <el-alert v-if="error || failure" :title="error || failure" type="error" :closable="false" />
  <div class="panel" v-if="canEdit">
    <el-form label-position="top" @submit.prevent="submit"
      ><CheckFields v-model="check" diagnostic />
      <p v-if="panel" class="muted">{{ t('probeSource') }}：{{ panel.name }}</p>
      <el-form-item v-else :label="t('nodeChoice')"
        ><el-select v-model="nodeIDs" multiple class="full-width"
          ><el-option
            v-for="node in nodes.filter((n) => n.enabled && !n.revoked)"
            :key="node.id"
            :value="node.id"
            :disabled="!supportsCheck(node, check)"
            :title="!supportsCheck(node, check) ? t('unsupportedNode') : ''"
            :label="node.name" /></el-select></el-form-item
      ><el-button native-type="submit" type="primary" :loading="busy">{{
        t('start')
      }}</el-button></el-form
    >
  </div>
  <div v-if="jobs.length" class="panel">
    <div class="section-title">
      <h3>{{ query ? t('batchResults') : t('history') }}</h3>
      <router-link v-if="route.query.monitor_id" :to="'/monitors/' + route.query.monitor_id"
        >{{ t('monitors') }} →</router-link
      >
    </div>
    <el-table :data="jobs"
      ><el-table-column :label="t('location')"
        ><template #default="{ row }">{{
          nodes.find((n) => n.id === row.job.node_id)?.name || row.job.node_id
        }}</template></el-table-column
      ><el-table-column :label="t('overview')"
        ><template #default="{ row }"
          ><HealthPill :status="row.result?.status || row.job.status" /></template></el-table-column
      ><el-table-column :label="t('stage')"
        ><template #default="{ row }">{{ row.result?.stage || '—' }}</template></el-table-column
      ><el-table-column :label="t('latency')"
        ><template #default="{ row }">{{
          row.result?.latency_ms == null ? '—' : row.result.latency_ms.toFixed(1) + ' ms'
        }}</template></el-table-column
      ></el-table
    >
  </div>
  <div class="grid-two">
    <article v-for="item in jobs" :key="item.job.id" class="panel">
      <div class="section-title">
        <strong>{{ nodes.find((n) => n.id === item.job.node_id)?.name || item.job.node_id }}</strong
        ><HealthPill :status="item.job.status" />
      </div>
      <p class="mono">{{ item.job.check.type.toUpperCase() }} · {{ endpoint(item.job.check) }}</p>
      <small>{{ date(item.job.created_at) }}</small
      ><ResultCard v-if="item.result?.id" :result="item.result" :label="t('output')" /><el-button
        v-if="canEdit && ['queued', 'running'].includes(item.job.status)"
        @click="cancel(item.job.id)"
        >{{ t('cancel') }}</el-button
      >
    </article>
  </div>
  <el-empty v-if="!jobs.length" :description="t('empty')" />
</template>
