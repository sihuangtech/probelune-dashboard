<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { api, isAdmin, usePolling, date, displayTimezone } from '../lib/api'
import { t } from '../lib/i18n'
import type { Settings } from '../lib/models'
import MemberSettings from '../features/settings/MemberSettings.vue'
import SessionSettings from '../features/settings/SessionSettings.vue'
const form = ref<Settings>({
    name: 'Probelune',
    public_url: '',
    timezone: 'UTC',
    raw_days: 7,
    event_days: 180,
    public_status: false,
  }),
  error = ref('')
const { data: audit } = usePolling<
  { id: string; at: number; user: string; action: string; resource: string }[]
>(() => (isAdmin.value && section.value === 'audit' ? '/audit' : ''), [])
const section = ref('instance')
const page = ref(1)
const auditRows = computed(() =>
  [...audit.value].sort((a, b) => b.at - a.at).slice((page.value - 1) * 10, page.value * 10),
)
const actionLabels: Record<string, [string, string]> = {
  'monitor.save': ['monitors', 'save'],
  'node.create': ['nodes', 'add'],
  'node.update': ['nodes', 'edit'],
  'node.credentials': ['nodes', 'credentials'],
  'incident.ack': ['incidents', 'acknowledge'],
  'incident.note': ['incidents', 'note'],
  'incident.silence': ['incidents', 'silence'],
  'maintenance.create': ['maintenance', 'add'],
  'maintenance.delete': ['maintenance', 'remove'],
  'settings.update': ['settings', 'edit'],
  'user.save': ['members', 'save'],
  'channel.save': ['channels', 'save'],
  'diagnostic.create': ['diagnostics', 'add'],
  'diagnostic.cancel': ['diagnostics', 'cancel'],
}
function actionLabel(value: string) {
  const labels = actionLabels[value]
  return labels ? labels.map(t).join(' / ') : value
}
const { data: users } = usePolling<{ id: string; username: string }[]>(
  () => (isAdmin.value && section.value === 'audit' ? '/users' : ''),
  [],
)
onMounted(async () => {
  try {
    form.value = await api<Settings>('/settings')
  } catch (e) {
    error.value = String(e)
  }
})
async function save() {
  try {
    await api('/settings', 'PUT', form.value)
    displayTimezone.value = form.value.timezone
    ElMessage.success(t('save'))
    error.value = ''
  } catch (e) {
    error.value = String(e)
  }
}
</script>
<template>
  <div class="page-title">
    <div>
      <p class="eyebrow">{{ t('settingsEyebrow') }}</p>
      <h1>{{ t('settings') }}</h1>
    </div>
  </div>
  <el-alert v-if="error" :title="error" type="error" />
  <el-tabs v-model="section">
    <el-tab-pane
      v-for="key in isAdmin ? ['instance', 'members', 'devices', 'audit'] : ['instance', 'devices']"
      :key="key"
      :name="key"
      :label="t(key)"
    />
  </el-tabs>
  <section v-if="section === 'instance'" class="panel instance-panel">
    <h2>{{ t('instance') }}</h2>
    <el-form label-position="top" :disabled="!isAdmin" @submit.prevent="save"
      ><div class="grid-two">
        <el-form-item :label="t('name')"><el-input v-model="form.name" /></el-form-item
        ><el-form-item :label="t('publicURL')"
          ><el-input
            v-model="form.public_url"
            placeholder="https://monitor.example.com" /></el-form-item
        ><el-form-item :label="t('timezone')"
          ><el-input v-model="form.timezone" placeholder="Asia/Shanghai" /></el-form-item
        ><el-form-item :label="t('rawDays')"
          ><el-input-number v-model="form.raw_days" :min="1" :max="90" /></el-form-item
        ><el-form-item :label="t('eventDays')"
          ><el-input-number v-model="form.event_days" :min="7" :max="365"
        /></el-form-item>
      </div>
      <el-checkbox v-model="form.public_status">{{ t('publicStatus') }}</el-checkbox>
      <div class="settings-actions">
        <el-button v-if="isAdmin" native-type="submit" type="primary">{{ t('save') }}</el-button
        ><a v-if="form.public_status" href="/status" target="_blank" rel="noopener"
          >{{ t('statusPage') }} ↗</a
        >
      </div></el-form
    >
  </section>
  <MemberSettings v-if="isAdmin && section === 'members'" /><SessionSettings
    v-if="section === 'devices'"
  />
  <section v-if="isAdmin && section === 'audit'" class="panel">
    <h2>{{ t('audit') }}</h2>
    <el-table :data="auditRows"
      ><el-table-column :label="t('updated')"
        ><template #default="{ row }">{{ date(row.at) }}</template></el-table-column
      ><el-table-column :label="t('username')" show-overflow-tooltip
        ><template #default="{ row }">{{
          users.find((u) => u.id === row.user)?.username || row.user
        }}</template></el-table-column
      >
      <el-table-column :label="t('auditAction')"
        ><template #default="{ row }"
          ><span :title="row.action">{{ actionLabel(row.action) }}</span></template
        ></el-table-column
      >
      <el-table-column :label="t('target')" show-overflow-tooltip
        ><template #default="{ row }">{{
          row.resource === 'instance' ? t('instance') : row.resource
        }}</template></el-table-column
      >
    </el-table>
    <el-pagination
      v-if="audit.length > 10"
      v-model:current-page="page"
      :page-size="10"
      :total="audit.length"
      layout="prev, pager, next"
      :pager-count="5"
    />
  </section>
</template>
<style scoped>
.instance-panel {
  max-width: 920px;
}
.settings-actions {
  display: flex;
  flex-direction: row-reverse;
  align-items: center;
  gap: 16px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--line);
}
.el-pagination {
  margin-top: 16px;
  justify-content: center;
}
</style>
