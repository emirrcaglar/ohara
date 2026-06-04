import type { VideoRow } from '../types/api'

function formatDuration(seconds: number) {
  if (!seconds) return ''

  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const remainingSeconds = seconds % 60

  if (hours) {
    return `${hours}:${String(minutes).padStart(2, '0')}:${String(remainingSeconds).padStart(2, '0')}`
  }
  return `${minutes}:${String(remainingSeconds).padStart(2, '0')}`
}

export function videoStats(video: VideoRow) {
  if (video.completed) return 'WATCHED'

  const parts: string[] = []
  if (video.duration) parts.push(formatDuration(video.duration))
  if (video.width && video.height) parts.push(`${video.height}P`)
  if (video.position && video.duration) {
    parts.push(`${Math.round((video.position / video.duration) * 100)}%`)
  }

  return parts.length ? parts.join(' · ') : 'READY'
}
