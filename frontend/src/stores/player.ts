import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { AudioRow } from '../types/api'
import { getAudioStreamUrl } from '../api/audio'

export const usePlayerStore = defineStore('player', () => {
  const currentTrack = ref<AudioRow | null>(null)
  const isPlaying = ref(false)
  const currentTime = ref(0)
  const duration = ref(0)
  const volume = ref(75)
  const queue = ref<AudioRow[]>([])

  const currentTrackUrl = computed(() => {
    if (!currentTrack.value) return null
    return getAudioStreamUrl(currentTrack.value.id)
  })

  const progress = computed(() => {
    if (duration.value === 0) return 0
    return (currentTime.value / duration.value) * 100
  })

  const formattedCurrentTime = computed(() => formatTime(currentTime.value))
  const formattedDuration = computed(() => formatTime(duration.value))

  function formatTime(seconds: number): string {
    const m = Math.floor(seconds / 60)
    const s = Math.floor(seconds % 60)
    return `${m}:${s.toString().padStart(2, '0')}`
  }

  function play(track?: AudioRow) {
    if (track && track.id !== currentTrack.value?.id) {
      currentTrack.value = track
      currentTime.value = 0
      duration.value = track.duration
    }
    isPlaying.value = true
  }

  function pause() {
    isPlaying.value = false
  }

  function togglePlay() {
    if (isPlaying.value) {
      pause()
    } else {
      play()
    }
  }

  function seek(time: number) {
    currentTime.value = Math.max(0, Math.min(time, duration.value))
  }

  function setVolume(v: number) {
    volume.value = Math.max(0, Math.min(100, v))
  }

  function next() {
    if (!currentTrack.value || queue.value.length === 0) return
    const currentIndex = queue.value.findIndex(t => t.id === currentTrack.value?.id)
    const nextIndex = (currentIndex + 1) % queue.value.length
    play(queue.value[nextIndex])
  }

  function previous() {
    if (!currentTrack.value || queue.value.length === 0) return
    const currentIndex = queue.value.findIndex(t => t.id === currentTrack.value?.id)
    const prevIndex = currentIndex <= 0 ? queue.value.length - 1 : currentIndex - 1
    play(queue.value[prevIndex])
  }

  function addToQueue(track: AudioRow) {
    if (!queue.value.find(t => t.id === track.id)) {
      queue.value.push(track)
    }
  }

  function setQueue(tracks: AudioRow[]) {
    queue.value = [...tracks]
  }

  function clearQueue() {
    queue.value = []
  }

  function updateCurrentTime(time: number) {
    currentTime.value = time
  }

  function updateDuration(d: number) {
    duration.value = d
  }

  return {
    currentTrack,
    isPlaying,
    currentTime,
    duration,
    volume,
    queue,
    currentTrackUrl,
    progress,
    formattedCurrentTime,
    formattedDuration,
    play,
    pause,
    togglePlay,
    seek,
    setVolume,
    next,
    previous,
    addToQueue,
    setQueue,
    clearQueue,
    updateCurrentTime,
    updateDuration
  }
})
