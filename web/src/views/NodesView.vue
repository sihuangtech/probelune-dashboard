<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api, date, usePolling, isAdmin } from '../lib/api'
import type { Node } from '../lib/models'
import { t } from '../lib/i18n'
import HealthPill from '../components/HealthPill.vue'
const { data, error, reload } = usePolling<Node[]>(() => '/nodes', [])
const targets = computed(() => data.value.filter((node) => node.role !== 'panel'))
const panel = computed(() => data.value.find((node) => node.role === 'panel' && !node.revoked))
const open = ref(false),
  form = ref({ id: '', name: '', region: '', network: '', enabled: true }),
  cidrs = ref(''),
  failure = ref('')
const registration = ref<{ node: Node; registration_token: string }>()
const installMethod = ref('binary'),
  platform = ref('linux-amd64')
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
  const args = `--panel '${location.origin}' --node-id '${n.node.id}'${location.protocol === 'http:' ? ' --allow-http' : ''}`
  const token = `export PROBELUNE_REGISTER_TOKEN='${n.registration_token}'`
  if (installMethod.value === 'docker')
    return `docker build -t probelune-node:0.1.1 https://github.com/sihuangtech/probelune-node.git#v0.1.1\n${token}\ndocker run -d --name probelune-node --restart unless-stopped --cap-drop ALL --cap-add NET_RAW -e PROBELUNE_REGISTER_TOKEN -v probelune-node-data:/data probelune-node:0.1.1 ${args} --data /data\nunset PROBELUNE_REGISTER_TOKEN`
  const binary = `probelune-node-${platform.value}`
  return `(\nset -eu\ncurl -fLO https://github.com/sihuangtech/probelune-node/releases/download/v0.1.1/${binary}\ncurl -fLO https://github.com/sihuangtech/probelune-node/releases/download/v0.1.1/SHA256SUMS\nsha256sum --ignore-missing -c SHA256SUMS\nchmod +x ${binary}\n${token}\n./${binary} ${args} --data ./node-data\nunset PROBELUNE_REGISTER_TOKEN\n)`
}
async function copyCommand() {
  try {
    await navigator.clipboard.writeText(instruction())
    ElMessage.success(t('copy'))
  } catch {
    ElMessage.error(t('copyFailed'))
  }
}
</script>
<template>
  <div class="page-title">
    <div>
      <p class="eyebrow">{{ t('serversEyebrow') }}</p>
      <h1>{{ t('nodes') }}</h1>
      <p class="muted">{{ t('nodeHint') }}</p>
    </div>
    <el-button v-if="isAdmin" type="primary" @click="edit()">+ {{ t('add') }}</el-button>
  </div>
  <el-alert v-if="error" :title="error" type="error" :closable="false" />
  <p v-if="panel" class="muted">
    {{ t('probeSource') }}：{{ panel.name }} · {{ t(online(panel) ? 'online' : 'offline') }}
  </p>
  <div v-if="!targets.length" class="panel empty-state">
    <div class="empty-symbol">◎</div>
    <h3>{{ t('empty') }}</h3>
    <p>{{ t('noNodes') }}</p>
  </div>
  <div class="node-grid">
    <article v-for="node in targets" :key="node.id" class="node-card">
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
    <el-tabs v-model="installMethod"
      ><el-tab-pane name="docker" :label="t('containerInstall')" /><el-tab-pane
        name="binary"
        :label="t('binaryInstall')"
    /></el-tabs>
    <el-select v-if="installMethod === 'binary'" v-model="platform"
      ><el-option value="linux-amd64" label="Linux x86_64" /><el-option
        value="linux-arm64"
        label="Linux ARM64"
    /></el-select>
    <p>
      <a href="https://github.com/sihuangtech/probelune-node" target="_blank" rel="noopener"
        >probelune-node · README / Releases ↗</a
      >
    </p>
    <pre class="code-block">{{ instruction() }}</pre>
    <el-button @click="copyCommand">{{ t('copyCommand') }}</el-button>
    <el-alert
      :title="
        data.some((n) => n.id === registration?.node.id && online(n))
          ? t('nodeConnected')
          : t('nodeWaiting')
      "
      :type="data.some((n) => n.id === registration?.node.id && online(n)) ? 'success' : 'info'"
      :closable="false"
      style="margin-top: 16px"
    />
    <router-link to="/monitors"
      ><el-button>{{ t('monitors') }} →</el-button></router-link
    >
  </el-dialog>
</template>
