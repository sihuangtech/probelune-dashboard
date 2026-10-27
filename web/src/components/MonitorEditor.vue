<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { api, isAdmin, clone } from '../lib/api'
import type { Monitor, Node, Channel } from '../lib/models'
import { defaultMonitor } from '../lib/models'
import { t } from '../lib/i18n'
import CheckFields from './CheckFields.vue'
const visible = defineModel<boolean>({ required: true })
const props = defineProps<{ monitor?: Monitor; nodes: Node[] }>()
const emit = defineEmits<{ saved: [] }>()
const model = ref(defaultMonitor()),
  headers = ref(''),
  tags = ref(''),
  error = ref(''),
  busy = ref(false)
const channels = ref<Channel[]>([])
watch(visible, async (opened) => {
  if (!opened) return
  model.value = props.monitor ? clone(props.monitor) : defaultMonitor()
  headers.value = ''
  tags.value = (model.value.tags || []).join(', ')
  error.value = ''
  if (isAdmin.value)
    channels.value = (await api<{ channel: Channel }[]>('/channels')).map((v) => v.channel)
})
async function save() {
  busy.value = true
  error.value = ''
  try {
    const m = clone(model.value)
    m.tags = tags.value
      .split(',')
      .map((x) => x.trim())
      .filter(Boolean)
    if (headers.value.trim()) m.check.headers = JSON.parse(headers.value)
    await api(`/monitors${m.id ? '/' + m.id : ''}`, m.id ? 'PUT' : 'POST', m)
    visible.value = false
    emit('saved')
    ElMessage.success(t('save'))
  } catch (e) {
    error.value = String(e)
  } finally {
    busy.value = false
  }
}
</script>
<template>
  <el-dialog
    v-model="visible"
    :title="t(model.id ? 'edit' : 'add') + ' · ' + t('monitors')"
    width="720px"
    destroy-on-close
  >
    <el-form label-position="top" @submit.prevent="save"
      ><div class="grid-two">
        <el-form-item :label="t('name')"
          ><el-input v-model="model.name" maxlength="100" /></el-form-item
        ><el-form-item :label="t('group')"><el-input v-model="model.group" /></el-form-item>
      </div>
      <CheckFields v-model="model.check" />
      <el-form-item v-if="model.check.type === 'http'" :label="t('headers')"
        ><el-input v-model="headers" type="textarea" placeholder='{"Authorization":"Bearer …"}'
      /></el-form-item>
      <el-form-item :label="t('nodeChoice')"
        ><el-select v-model="model.node_ids" multiple class="full-width"
          ><el-option
            v-for="node in nodes.filter((n) => !n.revoked)"
            :key="node.id"
            :value="node.id"
            :label="node.name" /></el-select
      ></el-form-item>
      <div class="grid-two">
        <el-form-item :label="t('interval')"
          ><el-input-number v-model="model.interval" :min="15" :max="86400" /></el-form-item
        ><el-form-item :label="t('minFailed')"
          ><el-input-number v-model="model.min_failed_nodes" :min="1" :max="50" /></el-form-item
        ><el-form-item :label="t('failThreshold')"
          ><el-input-number v-model="model.fail_threshold" :min="1" :max="20" /></el-form-item
        ><el-form-item :label="t('recoverThreshold')"
          ><el-input-number v-model="model.recover_threshold" :min="1" :max="20"
        /></el-form-item>
      </div>
      <el-form-item :label="t('tags')"
        ><el-input v-model="tags" placeholder="production, api"
      /></el-form-item>
      <el-form-item v-if="isAdmin" :label="t('channels')"
        ><el-select v-model="model.channel_ids" multiple class="full-width"
          ><el-option
            v-for="channel in channels"
            :key="channel.id"
            :value="channel.id"
            :label="channel.name" /></el-select
      ></el-form-item>
      <el-form-item
        ><el-checkbox v-model="model.enabled">{{ t('enabled') }}</el-checkbox
        ><el-checkbox v-model="model.public">{{ t('public') }}</el-checkbox></el-form-item
      >
      <el-alert v-if="error" :title="error" type="error" :closable="false" /><el-button
        type="primary"
        native-type="submit"
        :loading="busy"
        >{{ t('save') }}</el-button
      >
    </el-form></el-dialog
  >
</template>
