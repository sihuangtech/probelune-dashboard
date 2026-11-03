<script setup lang="ts">
import { computed } from 'vue'
import type { Monitor } from '../../lib/models'
import { t } from '../../lib/i18n'
import { checkTypeLabel, compareCheckTypes, useCheckCatalog } from '../checks/catalog'
import { monitorType } from './typeFilter'
import CheckIntroduction from '../checks/CheckIntroduction.vue'

const props = defineProps<{ monitors: Monitor[] }>()
const { definitions } = useCheckCatalog()
const types = computed(() =>
  [
    ...new Set([
      ...definitions.value.filter((item) => item.continuous).map((item) => item.type),
      ...props.monitors.map((monitor) => monitor.check.type),
      ...(monitorType.value ? [monitorType.value] : []),
    ]),
  ].sort(compareCheckTypes),
)
</script>

<template>
  <div class="monitor-type-filter">
    <span class="filter-label">{{ t('type') }}</span>
    <el-tabs v-model="monitorType" :aria-label="t('type')">
      <el-tab-pane name="">
        <template #label>
          <span class="type-label"
            >{{ t('allTypes') }}<small>{{ monitors.length }}</small></span
          >
        </template>
      </el-tab-pane>
      <el-tab-pane v-for="type in types" :key="type" :name="type">
        <template #label>
          <span class="type-label">
            {{ checkTypeLabel(type) }}
            <small>{{ monitors.filter((monitor) => monitor.check.type === type).length }}</small>
          </span>
        </template>
      </el-tab-pane>
    </el-tabs>
    <CheckIntroduction :type="monitorType" />
  </div>
</template>

<style scoped>
.monitor-type-filter {
  margin-bottom: 24px;
}
.filter-label {
  display: block;
  margin-bottom: 8px;
  color: var(--muted);
  font-size: 12px;
}
.type-description {
  margin: 16px 0 0;
  max-width: 76ch;
  color: var(--muted);
  font-size: 13px;
  line-height: 1.8;
}
.type-label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.type-label small {
  min-width: 22px;
  padding: 2px 7px;
  border-radius: 5px;
  background: var(--el-fill-color-light);
  color: var(--muted);
  text-align: center;
  font-variant-numeric: tabular-nums;
}
:deep(.el-tabs__header) {
  margin-bottom: 0;
}
:deep(.el-tabs__item) {
  height: 48px;
}
</style>
