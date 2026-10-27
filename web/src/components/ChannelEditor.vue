<script setup lang="ts">
import { ref, watch } from 'vue'
import { api, clone } from '../lib/api'
import type { Channel } from '../lib/models'
import { t } from '../lib/i18n'
const open = defineModel<boolean>({ required: true }),
  props = defineProps<{ channel?: Channel }>(),
  emit = defineEmits<{ saved: [] }>()
const fresh = (): Channel => ({
  id: '',
  name: '',
  type: 'webhook',
  enabled: true,
  reminder_minutes: 0,
  url: '',
  smtp_host: '',
  smtp_port: 587,
  username: '',
  password: '',
  from: '',
  to: '',
  tls_mode: 'starttls',
  policy: { allow_cidrs: [] },
})
const form = ref(fresh()),
  cidrs = ref(''),
  error = ref(''),
  secret = ref('')
watch(open, (v) => {
  if (v) {
    form.value = props.channel ? clone(props.channel) : fresh()
    cidrs.value = form.value.policy.allow_cidrs?.join('\n') || ''
    error.value = ''
    secret.value = ''
  }
})
async function save() {
  try {
    form.value.policy.allow_cidrs = cidrs.value
      .split('\n')
      .map((x) => x.trim())
      .filter(Boolean)
    const value = await api<{ signing_secret: string }>(
      '/channels' + (form.value.id ? '/' + form.value.id : ''),
      form.value.id ? 'PUT' : 'POST',
      form.value,
    )
    secret.value = value.signing_secret
    emit('saved')
    if (!secret.value) open.value = false
  } catch (e) {
    error.value = String(e)
  }
}
</script>
<template>
  <el-dialog v-model="open" :title="t('channels')" width="600px"
    ><el-form label-position="top" @submit.prevent="save"
      ><el-form-item :label="t('name')"><el-input v-model="form.name" /></el-form-item
      ><el-form-item :label="t('type')"
        ><el-select v-model="form.type"
          ><el-option value="webhook" label="Webhook (HMAC-SHA256)" /><el-option
            value="smtp"
            label="SMTP" /><el-option value="push" label="FCM HTTP v1" /></el-select></el-form-item
      ><el-form-item v-if="form.type === 'webhook'" label="URL"
        ><el-input v-model="form.url" /></el-form-item
      ><template v-if="form.type === 'smtp'"
        ><div class="grid-two">
          <el-form-item label="SMTP host"><el-input v-model="form.smtp_host" /></el-form-item
          ><el-form-item :label="t('port')"
            ><el-input-number v-model="form.smtp_port"
          /></el-form-item>
        </div>
        <el-form-item label="TLS"
          ><el-select v-model="form.tls_mode"
            ><el-option value="starttls" /><el-option value="tls" /></el-select></el-form-item
        ><el-form-item :label="t('username')"><el-input v-model="form.username" /></el-form-item
        ><el-form-item :label="t('password')"
          ><el-input v-model="form.password" type="password" show-password /></el-form-item
        ><el-form-item label="From"><el-input v-model="form.from" /></el-form-item
        ><el-form-item label="To"><el-input v-model="form.to" /></el-form-item></template
      ><el-alert
        v-if="form.type === 'push'"
        title="FCM requires GOOGLE_APPLICATION_CREDENTIALS on the server and a registered mobile device."
        type="info"
        :closable="false"
      /><el-form-item v-if="form.type !== 'push'" :label="t('policy')"
        ><el-input v-model="cidrs" type="textarea" /></el-form-item
      ><el-form-item :label="t('reminderMinutes')"
        ><el-input-number v-model="form.reminder_minutes" :min="0" :max="1440" /></el-form-item
      ><el-checkbox v-model="form.enabled">{{ t('enabled') }}</el-checkbox
      ><el-alert v-if="error" :title="error" type="error" :closable="false" />
      <pre v-if="secret" class="code-block">HMAC signing secret: {{ secret }}</pre>
      <div class="actions">
        <el-button type="primary" native-type="submit">{{ t('save') }}</el-button>
      </div></el-form
    ></el-dialog
  >
</template>
