<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api, canEdit, usePolling, date } from '../lib/api'
import type { Incident } from '../lib/models'
import { t } from '../lib/i18n'
import HealthPill from '../components/HealthPill.vue'
const route = useRoute(),
  router = useRouter(),
  tab = ref('active'),
  selected = ref('')
const { data, error, reload } = usePolling<Incident[]>(() => '/incidents', [])
const visible = computed(() =>
  data.value.filter(
    (v) =>
      tab.value === 'all' || (tab.value === 'active' ? v.recovered_at === 0 : v.recovered_at > 0),
  ),
)
const detail = computed(() => data.value.find((v) => v.id === selected.value))
watch(
  () => route.query.id,
  (id) => {
    if (typeof id === 'string') selected.value = id
  },
  { immediate: true },
)
async function action(v: Incident, kind: string) {
  try {
    let note = ''
    if (kind === 'note') note = (await ElMessageBox.prompt(t('note'), v.name)).value
    await api('/incidents/' + v.id + '/actions', 'POST', { action: kind, minutes: 60, note })
    await reload()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(String(e))
  }
}
async function recheck(v: Incident) {
  try {
    await api('/diagnostics', 'POST', { monitor_id: v.monitor_id })
    await router.push('/diagnostics')
  } catch (e) {
    ElMessage.error(String(e))
  }
}
</script>
<template>
  <div class="page-title">
    <div>
      <p class="eyebrow">EVIDENCE & RESPONSE</p>
      <h1>{{ t('incidents') }}</h1>
    </div>
  </div>
  <el-alert v-if="error" :title="error" type="error" :closable="false" />
  <div class="panel">
    <el-tabs v-model="tab"
      ><el-tab-pane name="active" :label="t('active')" /><el-tab-pane
        name="recovered"
        :label="t('recovered')" /></el-tabs
    ><el-table :data="visible" @row-click="(row: Incident) => (selected = row.id)"
      ><el-table-column :label="t('name')" min-width="200"
        ><template #default="{ row }"
          ><strong>{{ row.name }}</strong
          ><small v-if="row.ack_by"> · {{ t('acknowledged') }}</small></template
        ></el-table-column
      ><el-table-column :label="t('overview')"
        ><template #default="{ row }"
          ><HealthPill
            :status="
              row.recovered_at ? 'healthy' : row.node_id ? 'offline' : 'failed'
            " /></template></el-table-column
      ><el-table-column :label="t('startTime')" min-width="170"
        ><template #default="{ row }">{{ date(row.opened_at) }}</template></el-table-column
      ><el-table-column :label="t('endTime')" min-width="170"
        ><template #default="{ row }">{{ date(row.recovered_at) }}</template></el-table-column
      ></el-table
    >
  </div>
  <el-drawer
    :model-value="!!detail"
    @close="selected = ''"
    :title="detail?.name"
    size="min(600px,100vw)"
    ><template v-if="detail"
      ><div class="actions" v-if="canEdit">
        <el-button @click="action(detail, 'ack')">{{ t('acknowledge') }}</el-button
        ><el-button @click="action(detail, 'silence')">{{ t('silence') }}</el-button
        ><el-button @click="action(detail, 'note')">{{ t('note') }}</el-button
        ><el-button v-if="detail.monitor_id" type="primary" @click="recheck(detail)">{{
          t('recheck')
        }}</el-button>
      </div>
      <p v-if="detail.silenced_until" class="muted">
        {{ t('silence') }} · {{ date(detail.silenced_until) }}
      </p>
      <el-timeline style="margin-top: 30px"
        ><el-timeline-item
          v-for="(event, index) in detail.events"
          :key="index"
          :timestamp="date(event.at)"
          ><strong>{{ event.kind }}</strong>
          <p>{{ event.message }}</p></el-timeline-item
        ></el-timeline
      ></template
    ></el-drawer
  >
</template>
