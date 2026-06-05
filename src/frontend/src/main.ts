import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './assets/main.css'
import 'primeicons/primeicons.css'
import { initializeTheme } from './composables/useTheme'

initializeTheme()

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.mount('#root')

const staleBundleReloadParam = '__ohara_reload'

void router.isReady().then(() => {
  try {
    sessionStorage.removeItem('ohara:stale-bundle-reload')

    const url = new URL(window.location.href)
    if (url.searchParams.has(staleBundleReloadParam)) {
      url.searchParams.delete(staleBundleReloadParam)
      window.history.replaceState(window.history.state, '', url.toString())
    }
  } catch {
    // Ignore storage/history failures; the recovery URL is harmless if it remains visible.
  }
})
