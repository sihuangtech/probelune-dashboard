<script setup lang="ts">
import { endpoint } from '../lib/endpoint'
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api, canEdit, usePolling } from '../lib/api'
import type { Monitor, Node } from '../lib/models'
import { t } from '../lib/i18n'
import HealthPill from '../components/HealthPill.vue'
import MonitorEditor from '../features/monitors/MonitorEditor.vue'
import MonitorTypeFilter from '../features/monitors/MonitorTypeFilter.vue'
import { monitorType } from '../features/monitors/typeFilter'
const { data, error, reload } = usePolling<Monitor[]>(() => '/monitors', [])
const { data: nodes } = usePolling<Node[]>(() => '/nodes', [])
const search = ref(''),
  filter = ref(''),
  editor = ref(false),
  selected = ref<Monitor>()
const visible = computed(() =>
  data.value.filter(
    (m) =>
      (!monitorType.value || m.check.type === monitorType.value) &&
      (!filter.value || m.status === filter.value) &&
      `${m.name} ${m.check.target} ${m.group}`.toLowerCase().includes(search.value.toLowerCase()),
  ),
)
function edit(m?: Monitor, copy = false) {
  selected.value = m ? JSON.parse(JSON.stringify(m)) : undefined
  if (copy && selected.value) {
    selected.value.id = ''
    selected.value.version = 0
    selected.value.name += ' (copy)'
  }
  editor.value = true
}
async function change(m: Monitor, key: 'enabled' | 'archived') {
  try {
    if (key === 'archived') await ElMessageBox.confirm(m.name, t('removeConfirm'))
    await api('/monitors/' + m.id, 'PUT', { ...m, [key]: !m[key] })
    await reload()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(String(e))
  }
}
</script>
<template>
  <div class="page-title">
    <div>
      <p class="eyebrow">SERVICES & ENDPOINTS</p>
      <h1>{{ t('monitors') }}</h1>
    </div>
    <el-button v-if="canEdit" type="primary" @click="edit()">+ {{ t('add') }}</el-button>
  </div>
  <el-alert v-if="error" :title="error" type="error" :closable="false" />
  <div class="panel">
    <MonitorTypeFilter :monitors="data" />
    <el-button
      v-if="canEdit && monitorType && !data.some((m) => m.check.type === monitorType)"
      @click="edit()"
      >+ {{ t('add') }} {{ monitorType.toUpperCase() }}</el-button
    >
    <div class="toolbar">
      <el-input v-model="search" :placeholder="t('search')" clearable /><el-select
        v-model="filter"
        :placeholder="t('allStatuses')"
        clearable
        style="max-width: 180px"
        ><el-option
          v-for="status in [
            'healthy',
            'down',
            'partial',
            'unknown',
            'pending',
            'paused',
            'archived',
          ]"
          :key="status"
          :value="status"
          :label="t(status)"
      /></el-select>
    </div>
    <el-table class="wide-monitor-table" :data="visible" :empty-text="t('noTypeMonitors')"
      ><el-table-column :label="t('name')" min-width="200"
        ><template #default="{ row }"
          ><router-link :to="'/monitors/' + row.id"
            ><strong>{{ row.name }}</strong></router-link
          >
          <div class="muted">
            {{ endpoint(row.check) }}
          </div></template
        ></el-table-column
      ><el-table-column :label="t('monitorStatus')" width="190"
        ><template #default="{ row }"
          ><HealthPill :status="row.status" /><small v-if="!row.coverage_complete">
            {{ t('incomplete') }}</small
          ></template
        ></el-table-column
      ><el-table-column prop="group" :label="t('group')" /><el-table-column
        :label="t('interval')"
        width="120"
        ><template #default="{ row }">{{ row.interval }}s</template></el-table-column
      ><el-table-column v-if="canEdit" width="300"
        ><template #default="{ row }"
          ><div class="actions">
            <el-button size="small" @click="edit(row)">{{ t('edit') }}</el-button
            ><el-button size="small" @click="edit(row, true)">{{ t('copy') }}</el-button
            ><el-button size="small" @click="change(row, 'enabled')">{{
              t(row.enabled ? 'pause' : 'resume')
            }}</el-button
            ><el-button size="small" @click="change(row, 'archived')">{{
              t(row.archived ? 'restore' : 'archive')
            }}</el-button>
          </div></template
        ></el-table-column
      ></el-table
    >
    <div class="compact-monitor-list">
      <article v-for="m in visible" :key="m.id">
        <div class="section-title">
          <router-link :to="'/monitors/' + m.id"
            ><strong>{{ m.name }}</strong></router-link
          ><HealthPill :status="m.status" />
        </div>
        <p>{{ m.group }} · {{ endpoint(m.check) }}</p>
        <div v-if="canEdit" class="actions">
          <el-button size="small" @click="edit(m)">{{ t('edit') }}</el-button
          ><el-button size="small" @click="edit(m, true)">{{ t('copy') }}</el-button
          ><el-button size="small" @click="change(m, 'enabled')">{{
            t(m.enabled ? 'pause' : 'resume')
          }}</el-button
          ><el-button size="small" @click="change(m, 'archived')">{{
            t(m.archived ? 'restore' : 'archive')
          }}</el-button>
        </div>
      </article>
      <p v-if="!visible.length">{{ t('noTypeMonitors') }}</p>
    </div>
  </div>
  <MonitorEditor
    v-model="editor"
    :monitor="selected"
    :nodes="nodes"
    :default-type="monitorType || undefined"
    @saved="reload"
  />
</template>
