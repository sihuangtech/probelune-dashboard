<script setup lang="ts">
import { usePolling } from '../../lib/api'
import { t } from '../../lib/i18n'
import type { UptimeSeries } from '../../lib/models'
import { loadUptimeRange, uptimeQuery } from './range'
import UptimeStrip from './UptimeStrip.vue'

const props = defineProps<{ monitorId: string }>()
const range = loadUptimeRange()
// The list shows up to 24 recent buckets at the user's chosen resolution.
const query = uptimeQuery({
  ...range,
  hours: Math.min(range.hours, Math.ceil((24 * range.bucketMinutes) / 60)),
})
const { data, error } = usePolling<UptimeSeries | null>(
  () => '/monitors/' + props.monitorId + '/uptime?' + query,
  null,
  60000,
)
</script>
<template>
  <small v-if="error" :title="error">{{ t('historyUnavailable') }}</small>
  <router-link v-else-if="data" :to="'/monitors/' + monitorId" :aria-label="t('uptimeHistory')">
    <UptimeStrip compact :series="data" />
  </router-link>
  <small v-else>…</small>
</template>
