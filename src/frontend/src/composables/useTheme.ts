import { computed, ref } from 'vue'

export type ThemeMode = 'dark' | 'light'

const STORAGE_KEY = 'ohara:theme'
const FALLBACK_THEME: ThemeMode = 'dark'

const currentTheme = ref<ThemeMode>(readInitialTheme())

function isThemeMode(value: string | null): value is ThemeMode {
  return value === 'dark' || value === 'light'
}

function readInitialTheme(): ThemeMode {
  if (typeof window === 'undefined') return FALLBACK_THEME

  try {
    const stored = window.localStorage.getItem(STORAGE_KEY)
    if (isThemeMode(stored)) return stored
  } catch {
    // Keep the default when storage is unavailable.
  }

  return FALLBACK_THEME
}

function applyTheme(theme: ThemeMode) {
  if (typeof document === 'undefined') return

  const root = document.documentElement
  root.dataset.theme = theme
  root.classList.toggle('dark', theme === 'dark')
  root.classList.toggle('light', theme === 'light')
}

export function initializeTheme() {
  applyTheme(currentTheme.value)
}

export function useTheme() {
  const isLight = computed(() => currentTheme.value === 'light')

  function setTheme(theme: ThemeMode) {
    currentTheme.value = theme
    applyTheme(theme)

    try {
      window.localStorage.setItem(STORAGE_KEY, theme)
    } catch {
      // Theme still applies for this session when storage is unavailable.
    }
  }

  function toggleTheme() {
    setTheme(isLight.value ? 'dark' : 'light')
  }

  return {
    theme: currentTheme,
    isLight,
    setTheme,
    toggleTheme,
  }
}
