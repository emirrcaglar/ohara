import { describe, it, expect, vi } from 'vitest'

import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import App from '../App.vue'
import appRouter from '../router'

vi.mock('../api/manga', () => ({
  fetchMangaLibrary: vi.fn().mockResolvedValue({ items: [], total: 0 }),
  getMangaPageUrl: vi.fn(),
}))

describe('App', () => {
  it('opens the reading homepage at the root and provides library navigation', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: appRouter.options.routes,
    })
    await router.push('/')
    await router.isReady()
    const wrapper = mount(App, { global: { plugins: [createPinia(), router] } })
    await flushPromises()
    expect(wrapper.get('h1').text()).toBe('Latest read chapters')
    expect(wrapper.get('nav a[href="/"]').text()).toBe('Home')
    expect(wrapper.get('nav a[href="/library"]').text()).toBe('Library')
    wrapper.unmount()
  })
})
