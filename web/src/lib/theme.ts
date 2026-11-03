import { ref, watchEffect } from 'vue'

const stored = localStorage.getItem('probelune-theme')
export const theme = ref(['light', 'dark', 'system'].includes(stored || '') ? stored! : 'system')
const media = window.matchMedia('(prefers-color-scheme: dark)')
const systemDark = ref(media.matches)
media.addEventListener('change', (event) => {
  systemDark.value = event.matches
})
watchEffect(() => {
  const dark = theme.value === 'dark' || (theme.value === 'system' && systemDark.value)
  document.documentElement.classList.toggle('dark', dark)
  document.documentElement.style.colorScheme = dark ? 'dark' : 'light'
  localStorage.setItem('probelune-theme', theme.value)
})
