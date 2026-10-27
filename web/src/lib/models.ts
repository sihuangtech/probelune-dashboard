export interface Check {
  type: string
  target: string
  port: number
  timeout: number
  method: string
  expected_status: number
  contains: string
  follow_redirects: boolean
  certificate_min_days?: number
  insecure_tls: boolean
  dns_record: string
  resolver: string
  expected_dns: string
  count: number
  size: number
  max_loss: number
  family: string
  retries: number
  headers?: Record<string, string>
}
export interface Result {
  id: string
  monitor_id: string
  job_id: string
  node_id: string
  status: string
  scheduled_at: number
  received_at: number
  started_at: number
  stage: string
  error: string
  latency_ms: number | null
  data: Record<string, unknown>
  late: boolean
  attempts: number
}
export interface State {
  node_id: string
  status: string
  last_at: number
  result: Result
}
export interface Monitor {
  id: string
  name: string
  group: string
  tags: string[]
  check: Check
  node_ids: string[]
  channel_ids: string[]
  interval: number
  fail_threshold: number
  recover_threshold: number
  min_failed_nodes: number
  enabled: boolean
  archived: boolean
  version: number
  public: boolean
  status: string
  states: State[]
  coverage_complete: boolean
  maintenance: boolean
}
export interface Node {
  id: string
  name: string
  region: string
  network: string
  enabled: boolean
  revoked: boolean
  last_seen: number
  clock_skew: number
  policy: { allow_cidrs: string[] }
  info: {
    version: string
    capabilities: string[]
    platform: string
    queue: number
    dropped: number
  }
}
export interface Incident {
  id: string
  monitor_id: string
  node_id: string
  name: string
  opened_at: number
  recovered_at: number
  ack_by: string
  silenced_until: number
  affected: string[]
  suppressed: boolean
  events: { at: number; kind: string; message: string; user: string }[]
}
export interface Job {
  id: string
  node_id: string
  check: Check
  status: string
  created_at: number
  deadline: number
}
export interface Diagnostic {
  job: Job
  result: Result
}
export interface User {
  id: string
  username: string
  role: string
  disabled: boolean
}
export interface Overview {
  monitors: Monitor[]
  nodes: Node[]
  incidents: Incident[]
  summary: Record<string, number>
  online_nodes: number
  time: number
}
export interface Bucket {
  node_id: string
  at: number
  success: number
  failed: number
  errors: number
  latency_ms: number | null
}
export interface History {
  results: Result[]
  buckets?: Bucket[]
  bucket_seconds?: number
  summary: {
    node_id: string
    availability: number | null
    coverage: number | null
    expected: number
    latency_ms: number | null
  }[]
  coverage_basis: string
}
export interface Channel {
  reminder_minutes?: number
  id: string
  name: string
  type: string
  enabled: boolean
  url: string
  secret?: string
  smtp_host: string
  smtp_port: number
  username: string
  password?: string
  from: string
  to: string
  tls_mode: string
  policy: { allow_cidrs: string[] }
}
export interface Maintenance {
  id: string
  name: string
  monitor_ids: string[]
  start: number
  end: number
}
export interface Delivery {
  id: string
  channel_id: string
  kind: string
  status: string
  attempts: number
  error: string
  created_at: number
}
export interface Settings {
  name: string
  public_url: string
  timezone: string
  raw_days: number
  event_days: number
  public_status: boolean
}
export const defaultCheck = (): Check => ({
  type: 'http',
  target: '',
  port: 443,
  timeout: 10,
  method: 'GET',
  expected_status: 200,
  contains: '',
  follow_redirects: true,
  insecure_tls: false,
  dns_record: 'A',
  resolver: '',
  expected_dns: '',
  count: 4,
  size: 32,
  max_loss: 0,
  family: 'auto',
  retries: 0,
})
export const defaultMonitor = (): Monitor => ({
  id: '',
  name: '',
  group: '',
  tags: [],
  check: defaultCheck(),
  node_ids: [],
  channel_ids: [],
  interval: 60,
  fail_threshold: 3,
  recover_threshold: 2,
  min_failed_nodes: 1,
  enabled: true,
  archived: false,
  version: 0,
  public: false,
  status: 'unknown',
  states: [],
  coverage_complete: false,
  maintenance: false,
})
