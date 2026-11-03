<script setup lang="ts">
import { ref, watch } from 'vue'
import type { UptimeRange } from '../../lib/models'
import { validRange } from './range'
import { t } from '../../lib/i18n'
const model = defineModel<UptimeRange>({ required: true })
const draft = ref({ ...model.value }),
  invalid = ref(false)
watch(model, (value) => {
  draft.value = { ...value }
})
function apply() {
  invalid.value = !validRange(draft.value)
  if (invalid.value) return
  model.value = { ...draft.value }
  localStorage.setItem('probelune-uptime-range', JSON.stringify(model.value))
}
function preset(minutes: number) {
  draft.value.bucketMinutes = minutes
  apply()
}
</script>
<template>
  <div class="range-settings">
    <label
      >{{ t('rangeHours') }}<el-input-number v-model="draft.hours" :min="1" :max="720" size="small"
    /></label>
    <label
      >{{ t('bucketMinutes')
      }}<el-input-number
        v-model="draft.bucketMinutes"
        :min="5"
        :max="1440"
        :step="5"
        step-strictly
        size="small"
    /></label>
    <el-button size="small" @click="preset(15)">15 {{ t('minutes') }}</el-button>
    <el-button size="small" @click="preset(60)">1 {{ t('hour') }}</el-button>
    <el-button size="small" type="primary" @click="apply">{{ t('apply') }}</el-button>
    <small v-if="invalid" role="alert">{{ t('rangeLimit') }}</small>
  </div>
</template>
<style scoped>
.range-settings {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: end;
  margin: 16px 0;
}
label {
  display: grid;
  gap: 5px;
  font-size: 12px;
  color: var(--muted);
}
small {
  color: #ad403c;
  flex-basis: 100%;
}
</style>
