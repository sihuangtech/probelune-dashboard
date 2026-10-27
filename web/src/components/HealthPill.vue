<script setup lang="ts">
import { t } from '../lib/i18n'
defineProps<{ status: string }>()
</script>
<template>
  <span class="health-pill" :class="status"><i />{{ t(status) }}</span>
</template>
