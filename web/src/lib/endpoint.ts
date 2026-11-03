import type { Check } from './models'

export function endpoint(check: Check) {
  if (!['tcp', 'udp'].includes(check.type)) return check.target
  const host = check.target.trim()
  return (host.includes(':') && !host.startsWith('[') ? '[' + host + ']' : host) + ':' + check.port
}
