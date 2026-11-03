import type { UptimeRange } from '../../lib/models'

export function loadUptimeRange(): UptimeRange {
  try {
    const value = JSON.parse(localStorage.getItem('probelune-uptime-range') || 'null')
    if (validRange(value)) return value
  } catch {
    /* Keep a useful default when preferences are unavailable. */
  }
  return { hours: 24, bucketMinutes: 15 }
}
export function validRange(v: UptimeRange | null): v is UptimeRange {
  return (
    !!v &&
    Number.isInteger(v.hours) &&
    v.hours >= 1 &&
    v.hours <= 720 &&
    Number.isInteger(v.bucketMinutes) &&
    v.bucketMinutes >= 5 &&
    v.bucketMinutes <= 1440 &&
    v.bucketMinutes % 5 === 0 &&
    Math.ceil((v.hours * 60) / v.bucketMinutes) <= 2016
  )
}
export const uptimeQuery = (range: UptimeRange) =>
  `hours=${range.hours}&bucket_minutes=${range.bucketMinutes}`
