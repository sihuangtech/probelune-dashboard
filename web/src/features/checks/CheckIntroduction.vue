<script setup lang="ts">
import { t } from '../../lib/i18n'
import { checkTypeDescription } from './catalog'
defineProps<{ type: string }>()
const references: Record<string, [string, string]> = {
  tcp: ['RFC 9293', 'https://www.rfc-editor.org/rfc/rfc9293.html'],
  udp: ['RFC 768', 'https://datatracker.ietf.org/doc/html/rfc768'],
  ping: ['RFC 792', 'https://datatracker.ietf.org/doc/html/rfc792'],
  http: ['RFC 9110', 'https://datatracker.ietf.org/doc/html/rfc9110'],
  dns: ['RFC 1034', 'https://datatracker.ietf.org/doc/html/rfc1034'],
}
</script>
<template>
  <aside v-if="type" class="check-introduction">
    <div class="introduction-heading">
      <strong>{{ t('protocolName_' + type) }}</strong
      ><span>{{ t('protocolLayer_' + type) }}</span>
    </div>
    <p>{{ checkTypeDescription(type) }}</p>
    <a
      v-if="references[type]"
      :href="references[type]![1]"
      target="_blank"
      rel="noopener noreferrer"
      >{{ references[type]![0] }} ↗</a
    >
  </aside>
  <p v-else class="muted">{{ checkTypeDescription('') }}</p>
</template>
<style scoped>
.check-introduction {
  margin: 18px 0;
  padding: 18px 20px;
  background: var(--el-fill-color-light);
  border-left: 3px solid var(--teal);
  border-radius: 0 6px 6px 0;
}
.introduction-heading {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px 18px;
}
.introduction-heading span {
  color: var(--teal);
  font-size: 12px;
}
.check-introduction p {
  margin: 10px 0;
  max-width: 90ch;
  color: var(--muted);
  line-height: 1.85;
  font-size: 13px;
}
.check-introduction a {
  font-size: 12px;
}
</style>
