import { ref, shallowRef, computed, onMounted, onUnmounted } from 'vue'
import type { User } from './models'

export const user = ref<User | null>(null)
export const displayTimezone = ref(Intl.DateTimeFormat().resolvedOptions().timeZone)
export const canEdit = computed(() => ['admin', 'operator'].includes(user.value?.role ?? ''))
export const isAdmin = computed(() => user.value?.role === 'admin')

export async function api<T>(
  path: string,
  method = 'GET',
  body?: unknown,
  headers?: HeadersInit,
): Promise<T> {
  const response = await fetch(`/api/v1${path}`, {
    method,
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json', ...headers },
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  if (response.status === 401) user.value = null
  if (!response.ok) {
    const data = await response.json().catch(() => ({ error: `HTTP ${response.status}` }))
    throw new Error(data.error || `HTTP ${response.status}`)
  }
  return response.status === 204 ? (undefined as T) : ((await response.json()) as T)
}

// 轮询避免并发重入；离开页面时解除定时器，保留最后一次成功结果。
export function usePolling<T>(path: () => string, initial: T) {
  const data = shallowRef<T>(initial)
  const error = ref('')
  const updated = ref(0)
  const loading = ref(false)
  let timer: ReturnType<typeof setInterval>
  let active = true
  async function reload() {
    if (loading.value || !path()) return
    loading.value = true
    try {
      const result = await api<T>(path())
      if (active) {
        data.value = result
        error.value = ''
        updated.value = Date.now()
      }
    } catch (e) {
      if (active) error.value = String(e)
    } finally {
      loading.value = false
    }
  }
  onMounted(() => {
    void reload()
    timer = setInterval(() => void reload(), 3000)
  })
  onUnmounted(() => {
    active = false
    clearInterval(timer)
  })
  return { data, error, updated, loading, reload }
}
export const date = (value?: number) =>
  value
    ? new Date(value * 1000).toLocaleString(undefined, { timeZone: displayTimezone.value })
    : '—'
export const clone = <T>(value: T): T => JSON.parse(JSON.stringify(value))
export const number = (value?: number | null, suffix = '') =>
  value == null ? '—' : `${value.toFixed(1)}${suffix}`

export function download(name: string, content: string, type = 'application/json') {
  const url = URL.createObjectURL(new Blob([content], { type }))
  const a = document.createElement('a')
  a.href = url
  a.download = name
  a.click()
  URL.revokeObjectURL(url)
}
