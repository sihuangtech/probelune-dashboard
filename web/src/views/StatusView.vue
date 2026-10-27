<script setup lang="ts">
import { usePolling, date } from '../lib/api'
import { t } from '../lib/i18n'
import HealthPill from '../components/HealthPill.vue'
const { data, error, updated } = usePolling<{
  name: string
  monitors: { id: string; name: string; status: string; maintenance: boolean }[]
}>(() => '/status', { name: 'Probelune', monitors: [] })
</script>
<template>
  <main class="public-status">
    <p class="eyebrow">PROBELUNE · SERVICE STATUS</p>
    <h1>{{ data.name }}</h1>
    <p class="muted">{{ t('updated') }} · {{ date(updated / 1000) }}</p>
    <el-alert v-if="error" :title="error" type="error" :closable="false" />
    <section class="panel">
      <el-empty v-if="!data.monitors.length" :description="t('empty')" />
      <div v-for="m in data.monitors" :key="m.id" class="section-title status-row">
        <strong>{{ m.name }}</strong
        ><HealthPill :status="m.maintenance ? 'maintenance' : m.status" />
      </div>
    </section>
    <p class="muted">{{ t('tagline') }}</p>
  </main>
</template>
<style scoped>
.public-status {
  max-width: 760px;
  margin: 12vh auto;
  padding: 24px;
}
.status-row {
  padding: 20px 0;
  border-bottom: 1px solid var(--border);
}
</style>
