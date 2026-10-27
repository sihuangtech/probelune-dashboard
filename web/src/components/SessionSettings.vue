<script setup lang="ts">
import { ref } from 'vue'
import QRCode from 'qrcode'
import { ElMessage } from 'element-plus'
import { api, date, usePolling } from '../lib/api'
import { t } from '../lib/i18n'
type Session = { id: string; device: string; expires: number }
type Device = { id: string; platform: string; created_at: number; disabled: boolean }
const { data: sessions, reload } = usePolling<Session[]>(() => '/auth/sessions', [])
const { data: devices, reload: reloadDevices } = usePolling<Device[]>(() => '/devices', [])
const qr = ref(''),
  code = ref(''),
  expires = ref(0)
function clearPair() {
  qr.value = ''
  code.value = ''
}
async function pair() {
  try {
    const p = await api<{ code: string; expires: number }>('/auth/pair', 'POST', {})
    code.value = p.code
    expires.value = p.expires
    qr.value = await QRCode.toDataURL(
      JSON.stringify({ type: 'probelune-pair', version: 1, url: location.origin, code: p.code }),
      { width: 280, margin: 2 },
    )
  } catch (e) {
    ElMessage.error(String(e))
  }
}
async function remove(path: string) {
  try {
    await api(path, 'DELETE')
    await reload()
    await reloadDevices()
  } catch (e) {
    ElMessage.error(String(e))
  }
}
</script>
<template>
  <section class="panel">
    <div class="section-title">
      <h2>{{ t('devices') }}</h2>
      <el-button @click="pair">{{ t('pair') }}</el-button>
    </div>
    <el-table :data="sessions"
      ><el-table-column prop="device" :label="t('name')" /><el-table-column :label="t('endTime')"
        ><template #default="{ row }">{{ date(row.expires) }}</template></el-table-column
      ><el-table-column width="100"
        ><template #default="{ row }"
          ><el-button text type="danger" @click="remove('/auth/sessions/' + row.id)">{{
            t('remove')
          }}</el-button></template
        ></el-table-column
      ></el-table
    ><el-table v-if="devices.length" :data="devices"
      ><el-table-column prop="platform" :label="t('type')" /><el-table-column :label="t('updated')"
        ><template #default="{ row }">{{ date(row.created_at) }}</template></el-table-column
      ><el-table-column width="100"
        ><template #default="{ row }"
          ><el-button text type="danger" @click="remove('/devices/' + row.id)">{{
            t('remove')
          }}</el-button></template
        ></el-table-column
      ></el-table
    >
  </section>
  <el-dialog :model-value="!!qr" :title="t('pair')" width="380px" @close="clearPair"
    ><img :src="qr" alt="One-time pairing QR code" width="280" height="280" />
    <p>{{ date(expires) }}</p>
    <p class="muted">{{ t('pair') }}: {{ code }}</p></el-dialog
  >
</template>
