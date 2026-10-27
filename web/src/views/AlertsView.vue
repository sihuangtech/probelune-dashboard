<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api, usePolling, isAdmin, canEdit, date } from '../lib/api'
import type { Channel, Maintenance, Delivery, Monitor } from '../lib/models'
import { t } from '../lib/i18n'
import ChannelEditor from '../components/ChannelEditor.vue'
import HealthPill from '../components/HealthPill.vue'
const { data: channels, reload: reloadChannels } = usePolling<
  { channel: Channel; has_secret: boolean }[]
>(() => (isAdmin.value ? '/channels' : ''), [])
const {
  data: windows,
  reload: reloadWindows,
  error,
} = usePolling<Maintenance[]>(() => '/maintenance', [])
const { data: monitors } = usePolling<Monitor[]>(() => '/monitors', [])
const { data: deliveries } = usePolling<Delivery[]>(() => (canEdit.value ? '/deliveries' : ''), [])
const channelOpen = ref(false),
  selected = ref<Channel>(),
  maintenanceOpen = ref(false),
  form = ref({ name: '', monitor_ids: [] as string[] }),
  range = ref<[Date, Date]>([new Date(), new Date(Date.now() + 3600000)])
function edit(ch?: Channel) {
  selected.value = ch
  channelOpen.value = true
}
async function test(ch: Channel) {
  try {
    await api('/channels/' + ch.id + '/test', 'POST', {})
    ElMessage.success(t('queued'))
  } catch (e) {
    ElMessage.error(String(e))
  }
}
async function saveWindow() {
  try {
    await api('/maintenance', 'POST', {
      ...form.value,
      start: Math.floor(range.value[0].getTime() / 1000),
      end: Math.floor(range.value[1].getTime() / 1000),
    })
    maintenanceOpen.value = false
    await reloadWindows()
  } catch (e) {
    ElMessage.error(String(e))
  }
}
async function remove(id: string) {
  try {
    await ElMessageBox.confirm(t('removeConfirm'))
    await api('/maintenance/' + id, 'DELETE')
    await reloadWindows()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(String(e))
  }
}
</script>
<template>
  <div class="page-title">
    <div>
      <p class="eyebrow">NOTIFICATIONS & QUIET WINDOWS</p>
      <h1>{{ t('alerts') }}</h1>
    </div>
  </div>
  <el-alert v-if="error" :title="error" type="error" :closable="false" />
  <section v-if="isAdmin" class="panel">
    <div class="section-title">
      <h2>{{ t('channels') }}</h2>
      <el-button type="primary" @click="edit()">+ {{ t('add') }}</el-button>
    </div>
    <el-table :data="channels"
      ><el-table-column prop="channel.name" :label="t('name')" /><el-table-column
        prop="channel.type"
        :label="t('type')"
      /><el-table-column
        ><template #default="{ row }"
          ><HealthPill
            :status="row.channel.enabled ? 'healthy' : 'paused'" /></template></el-table-column
      ><el-table-column
        ><template #default="{ row }"
          ><el-button size="small" @click="edit(row.channel)">{{ t('edit') }}</el-button
          ><el-button size="small" @click="test(row.channel)">{{ t('test') }}</el-button></template
        ></el-table-column
      ></el-table
    >
  </section>
  <section class="panel">
    <div class="section-title">
      <h2>{{ t('maintenance') }}</h2>
      <el-button v-if="canEdit" @click="maintenanceOpen = true">+ {{ t('add') }}</el-button>
    </div>
    <el-table :data="windows"
      ><el-table-column prop="name" :label="t('name')" /><el-table-column :label="t('startTime')"
        ><template #default="{ row }">{{ date(row.start) }}</template></el-table-column
      ><el-table-column :label="t('endTime')"
        ><template #default="{ row }">{{ date(row.end) }}</template></el-table-column
      ><el-table-column v-if="canEdit"
        ><template #default="{ row }"
          ><el-button text type="danger" @click="remove(row.id)">{{
            t('remove')
          }}</el-button></template
        ></el-table-column
      ></el-table
    >
  </section>
  <section v-if="canEdit" class="panel">
    <h2>{{ t('deliveries') }}</h2>
    <el-table :data="deliveries"
      ><el-table-column prop="kind" :label="t('type')" /><el-table-column :label="t('overview')"
        ><template #default="{ row }"
          ><HealthPill :status="row.status" /></template></el-table-column
      ><el-table-column prop="attempts" :label="t('retries')" /><el-table-column
        prop="error"
        :label="t('details')"
      /><el-table-column :label="t('updated')"
        ><template #default="{ row }">{{ date(row.created_at) }}</template></el-table-column
      ></el-table
    >
  </section>
  <ChannelEditor v-model="channelOpen" :channel="selected" @saved="reloadChannels" />
  <el-dialog v-model="maintenanceOpen" :title="t('maintenance')" width="580px"
    ><el-form label-position="top" @submit.prevent="saveWindow"
      ><el-form-item :label="t('name')"><el-input v-model="form.name" /></el-form-item
      ><el-form-item :label="t('monitors')"
        ><el-select
          v-model="form.monitor_ids"
          multiple
          :placeholder="t('allMonitors')"
          class="full-width"
          ><el-option
            v-for="m in monitors"
            :key="m.id"
            :label="m.name"
            :value="m.id" /></el-select></el-form-item
      ><el-date-picker v-model="range" type="datetimerange" class="full-width" /><el-button
        style="margin-top: 20px"
        type="primary"
        native-type="submit"
        >{{ t('save') }}</el-button
      ></el-form
    ></el-dialog
  >
</template>
