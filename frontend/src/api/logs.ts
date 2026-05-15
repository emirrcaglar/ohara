import type { LogEntry } from '../types/api'

export function createLogStream(
  onEntry: (entry: LogEntry) => void,
  onError?: (event: Event) => void
): EventSource {
  const source = new EventSource('/api/logs/stream')
  source.onmessage = (event: MessageEvent) => {
    try {
      const entry = JSON.parse(event.data) as LogEntry
      onEntry(entry)
    } catch {}
  }
  if (onError) {
    source.onerror = onError
  }
  return source
}

export async function fetchLogSnapshot(): Promise<LogEntry[]> {
  const res = await fetch('/api/logs')
  const json = await res.json()
  return (json.entries ?? []) as LogEntry[]
}
