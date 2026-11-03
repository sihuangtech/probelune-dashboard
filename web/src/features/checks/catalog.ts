import { ref } from 'vue'
import { api } from '../../lib/api'
import type { Check, Node } from '../../lib/models'

interface Definition {
  type: string
  continuous: boolean
  group: string
}
// The API owns the check catalog. Forms and node selection share this state.
const definitions = ref<Definition[]>([])
const catalogError = ref('')
let pending: Promise<void> | undefined
export function useCheckCatalog() {
  if (!definitions.value.length && !pending) {
    pending = api<Definition[]>('/check-types')
      .then((value) => {
        definitions.value = value
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
