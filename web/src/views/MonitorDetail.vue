<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, usePolling, date, canEdit } from '../lib/api'
import type { Monitor, Node, Job } from '../lib/models'
import { t } from '../lib/i18n'
import HealthPill from '../components/HealthPill.vue'
import MonitorEditor from '../features/monitors/MonitorEditor.vue'
import MonitorHistory from '../features/history/MonitorHistory.vue'
import ResultCard from '../features/checks/ResultCard.vue'
const route = useRoute(),
  router = useRouter(),
  editor = ref(false)
const { data: monitors, reload, error } = usePolling<Monitor[]>(() => '/monitors', [])
const { data: nodes } = usePolling<Node[]>(() => '/nodes', [])
const monitor = computed(() => monitors.value.find((m) => m.id === route.params.id))
async function recheck() {
  try {
    const jobs = await api<Job[]>('/diagnostics', 'POST', { monitor_id: monitor.value?.id })
    await router.push({
      path: '/diagnostics',
      query: { batch_id: jobs[0]?.batch_id, monitor_id: monitor.value?.id },
    })
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
  <template v-if="monitor"
    ><div class="panel section-title">
      <HealthPill :status="monitor.maintenance ? 'maintenance' : monitor.status" /><span
        >{{ t('probeEvery') }} {{ monitor.interval }} {{ t('seconds') }}</span
      ><small
        >{{ t('updated') }} {{ date(Math.max(0, ...monitor.states.map((s) => s.last_at))) }}</small
      >
    </div>
    <MonitorHistory :key="monitor.id" :monitor-id="monitor.id" :nodes="nodes" />
    <div class="grid-two">
      <ResultCard
        v-for="state in monitor.states.filter((s) => s.result?.id)"
        :key="state.node_id"
        :result="state.result"
        :label="nodes.find((n) => n.id === state.node_id)?.name"
      />
    </div>
    <MonitorEditor v-model="editor" :monitor="monitor" :nodes="nodes" @saved="reload"
  /></template>
</template>
