<script setup lang="ts">
import type { Result } from '../lib/models'
import { date, number } from '../lib/api'
import { t } from '../lib/i18n'
import HealthPill from './HealthPill.vue'
defineProps<{ result: Result; label?: string }>()
</script>
<template>
  <article class="result-card">
    <div class="section-title">
      <strong>{{ label || result.node_id }}</strong
      ><HealthPill :status="result.status || 'unknown'" />
    </div>
    <dl class="result-meta">
      <dt>{{ t('latency') }}</dt>
      <dd>{{ number(result.latency_ms, ' ms') }}</dd>
      <dt>{{ t('updated') }}</dt>
      <dd>{{ date(result.started_at) }}</dd>
      <dt>{{ t('stage') }}</dt>
      <dd>{{ result.stage || '—' }}</dd>
    </dl>
    <el-alert v-if="result.error" :title="result.error" type="error" :closable="false" />
    <pre v-if="result.data && Object.keys(result.data).length" class="code-block">{{
      JSON.stringify(result.data, null, 2)
    }}</pre>
  </article>
</template>
