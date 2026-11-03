import { ref, watch } from 'vue'

const storageKey = 'probelune-monitor-type'
// Share the same selection between the overview and monitor list.
export const monitorType = ref(localStorage.getItem(storageKey) || 'tcp')
watch(monitorType, (value) => localStorage.setItem(storageKey, value))
