<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { api, isAdmin, usePolling, date, displayTimezone } from '../lib/api'
import { t } from '../lib/i18n'
import type { Settings } from '../lib/models'
import MemberSettings from '../components/MemberSettings.vue'
import SessionSettings from '../components/SessionSettings.vue'
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
>(() => (isAdmin.value ? '/audit' : ''), [])
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
      <p class="eyebrow">YOUR OBSERVATION NETWORK</p>
      <h1>{{ t('settings') }}</h1>
    </div>
  </div>
  <el-alert v-if="error" :title="error" type="error" />
  <section class="panel">
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
      <div class="actions">
        <el-button v-if="isAdmin" native-type="submit" type="primary">{{ t('save') }}</el-button
        ><a href="/status" target="_blank" rel="noopener">{{ t('statusPage') }} ↗</a>
      </div></el-form
    >
  </section>
  <MemberSettings v-if="isAdmin" /><SessionSettings />
  <section v-if="isAdmin" class="panel">
    <h2>{{ t('audit') }}</h2>
    <el-table :data="audit.slice().reverse().slice(0, 100)"
      ><el-table-column :label="t('updated')"
        ><template #default="{ row }">{{ date(row.at) }}</template></el-table-column
      ><el-table-column prop="user" :label="t('username')" /><el-table-column
        prop="action"
        :label="t('type')" /><el-table-column prop="resource" :label="t('target')"
    /></el-table>
  </section>
</template>
