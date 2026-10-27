<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api, date, usePolling, isAdmin } from '../lib/api'
import type { Node } from '../lib/models'
import { t } from '../lib/i18n'
import HealthPill from '../components/HealthPill.vue'
const { data, error, reload } = usePolling<Node[]>(() => '/nodes', [])
const open = ref(false),
  form = ref({ id: '', name: '', region: '', network: '', enabled: true }),
  cidrs = ref(''),
  failure = ref('')
const registration = ref<{ node: Node; registration_token: string }>()
function edit(node?: Node) {
  form.value = node ? { ...node } : { id: '', name: '', region: '', network: '', enabled: true }
  cidrs.value = node?.policy.allow_cidrs?.join('\n') || ''
  failure.value = ''
  open.value = true
}
async function save() {
  try {
    const body = {
      ...form.value,
      policy: {
        allow_cidrs: cidrs.value
          .split('\n')
          .map((x) => x.trim())
          .filter(Boolean),
      },
    }
    if (form.value.id) await api('/nodes/' + form.value.id, 'PUT', body)
    else registration.value = await api('/nodes', 'POST', body)
    open.value = false
    await reload()
  } catch (e) {
    failure.value = String(e)
  }
}
async function credential(node: Node, revoke: boolean) {
  try {
    await ElMessageBox.confirm(node.name, t('removeConfirm'))
    const result = await api<{ node: Node; registration_token: string }>(
      '/nodes/' + node.id + '/credentials',
      'POST',
      { revoke },
    )
    if (!revoke) registration.value = result
    await reload()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(String(e))
  }
}
function online(node: Node) {
  return !node.revoked && node.enabled && Date.now() / 1000 - node.last_seen <= 60
}
function instruction() {
  if (!registration.value) return ''
  const n = registration.value
  return `export PROBELUNE_REGISTER_TOKEN='${n.registration_token}'\n./probelune-node --panel '${location.origin}' --node-id '${n.node.id}' --data ./node-data${location.protocol === 'http:' ? ' --allow-http' : ''}\nunset PROBELUNE_REGISTER_TOKEN`
}
</script>
<template>
  <div class="page-title">
    <div>
      <p class="eyebrow">YOUR VANTAGE POINTS</p>
      <h1>{{ t('nodes') }}</h1>
      <p class="muted">{{ t('nodeHint') }}</p>
    </div>
    <el-button v-if="isAdmin" type="primary" @click="edit()">+ {{ t('add') }}</el-button>
  </div>
  <el-alert v-if="error" :title="error" type="error" :closable="false" />
  <div v-if="!data.length" class="panel empty-state">
    <div class="empty-symbol">◎</div>
    <h3>{{ t('empty') }}</h3>
    <p>{{ t('noNodes') }}</p>
  </div>
  <div class="node-grid">
    <article v-for="node in data" :key="node.id" class="node-card">
      <div class="section-title">
        <h3>{{ node.name }}</h3>
        <HealthPill
          :status="
            node.revoked
              ? 'archived'
              : !node.enabled
                ? 'paused'
                : online(node)
                  ? 'online'
                  : 'offline'
          "
        />
      </div>
      <p class="muted">{{ node.region || '—' }} · {{ node.network || '—' }}</p>
      <dl class="result-meta">
        <dt>{{ t('updated') }}</dt>
        <dd>{{ date(node.last_seen) }}</dd>
        <dt>{{ t('version') }}</dt>
        <dd>{{ node.info.version || '—' }} · {{ node.info.platform }}</dd>
        <dt>Clock / queue</dt>
        <dd>{{ node.clock_skew }}s / {{ node.info.queue || 0 }}</dd>
      </dl>
      <div class="node-tags">
        <span v-for="capability in node.info.capabilities" :key="capability">{{
          capability.toUpperCase()
        }}</span>
      </div>
      <div v-if="isAdmin" class="actions">
        <el-button size="small" @click="edit(node)">{{ t('edit') }}</el-button
        ><el-button size="small" @click="credential(node, false)">{{ t('credentials') }}</el-button
        ><el-button size="small" type="danger" plain @click="credential(node, true)">{{
          t('revoke')
        }}</el-button>
      </div>
    </article>
  </div>
  <el-dialog v-model="open" :title="t('nodes')" width="560px"
    ><el-form label-position="top" @submit.prevent="save"
      ><el-form-item :label="t('name')"><el-input v-model="form.name" /></el-form-item>
      <div class="grid-two">
        <el-form-item :label="t('region')"><el-input v-model="form.region" /></el-form-item
        ><el-form-item :label="t('network')"><el-input v-model="form.network" /></el-form-item>
      </div>
      <el-form-item :label="t('policy')"
        ><el-input v-model="cidrs" type="textarea" placeholder="10.0.0.0/8" /></el-form-item
      ><el-checkbox v-model="form.enabled">{{ t('enabled') }}</el-checkbox
      ><el-alert v-if="failure" :title="failure" type="error" :closable="false" />
      <div class="actions">
        <el-button type="primary" native-type="submit">{{ t('save') }}</el-button>
      </div></el-form
    ></el-dialog
  >
  <el-dialog
    :model-value="!!registration"
    @close="registration = undefined"
    :title="t('install')"
    width="760px"
    ><el-alert :title="t('tokenOnce')" type="warning" :closable="false" />
    <p>
      <a href="https://github.com/sihuangtech/probelune-node" target="_blank" rel="noopener"
        >probelune-node · README / Releases ↗</a
      >
    </p>
    <pre class="code-block">{{ instruction() }}</pre>
  </el-dialog>
</template>
