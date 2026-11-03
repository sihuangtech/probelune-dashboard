import { ref } from 'vue'
import { api } from '../../lib/api'
import type { Check, Node } from '../../lib/models'
import { t } from '../../lib/i18n'

interface Definition {
  type: string
  continuous: boolean
  group: string
}
// The API owns the check catalog. Forms and node selection share this state.
const definitions = ref<Definition[]>([])
const catalogError = ref('')
let pending: Promise<void> | undefined
const typeOrder = ['tcp', 'udp', 'ping', 'http', 'dns', 'mtr', 'traceroute']
export function checkTypeDescription(type: string) {
  return t('checkIntro_' + (typeOrder.includes(type) ? type : 'all'))
}
export function compareCheckTypes(a: string, b: string) {
  const rank = (type: string) => {
    const index = typeOrder.indexOf(type)
    return index < 0 ? typeOrder.length : index
  }
  return rank(a) - rank(b) || a.localeCompare(b)
}
export function checkTypeLabel(type: string) {
  return (
    ({ http: 'HTTP / HTTPS', ping: 'ICMP' } as Record<string, string>)[type] || type.toUpperCase()
  )
}
export function useCheckCatalog() {
  if (!definitions.value.length && !pending) {
    pending = api<Definition[]>('/check-types')
      .then((value) => {
        definitions.value = value.sort((a, b) => compareCheckTypes(a.type, b.type))
        catalogError.value = ''
      })
      .catch((error) => {
        catalogError.value = String(error)
      })
      .finally(() => {
        pending = undefined
      })
  }
  return { definitions, catalogError }
}
export function supportsCheck(node: Node, check: Check) {
  const capabilities = node.info?.capabilities || []
  return (
    capabilities.includes(check.type) &&
    (check.type !== 'http' ||
      !check.certificate_min_days ||
      capabilities.includes('certificate_expiry'))
  )
}
