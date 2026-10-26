import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import { fetchMangaLibrary } from '../api/manga'
import type { MangaRow } from '../types/api'

vi.mock('../api/manga', () => ({
  fetchMangaLibrary: vi.fn(),
  getMangaPageUrl: (id: number, page: number) => `/api/manga/${id}/page/${page}`,
}))

const fetchHistory = vi.mocked(fetchMangaLibrary)
const chapter: MangaRow = {
  id: 7,
  catalogId: null,
  title: 'Chapter 07',
  path: '',
  pageCount: 20,
  currentPage: 4,
  fileExtension: '.cbz',
  updatedAt: '2026-10-26T12:00:00Z',
}

async function render() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: HomeView },
      { path: '/:pathMatch(.*)*', component: { template: '<div />' } },
    ],
  })
  await router.push('/')
  await router.isReady()
  return mount(HomeView, { global: { plugins: [router] } })
}

afterEach(() => vi.resetAllMocks())

describe('homepage reading history', () => {
  it('shows a loading state before the request completes', async () => {
    fetchHistory.mockReturnValue(new Promise(() => {}))
    const wrapper = await render()
    expect(wrapper.get('[role="status"]').text()).toContain('Loading reading history')
    wrapper.unmount()
  })

  it('offers the library when no chapters have been read', async () => {
    fetchHistory.mockResolvedValue({ items: [], total: 0 })
    const wrapper = await render()
    await flushPromises()
    expect(wrapper.text()).toContain('Your next chapter starts here')
    expect(wrapper.findAll('a').some((link) => link.attributes('href') === '/library')).toBe(true)
    wrapper.unmount()
  })

  it('filters unread chapters and sorts library entries by progress timestamp', async () => {
    fetchHistory.mockResolvedValue({
      items: [
        {
          ...chapter,
          id: 2,
          title: 'Earlier chapter',
          currentPage: 19,
          updatedAt: '2026-10-25T12:00:00Z',
        },
        { ...chapter, id: 3, title: 'Unread chapter', updatedAt: null },
        { ...chapter, id: 4, title: 'Empty archive', pageCount: 0 },
        { ...chapter, id: 5, title: 'Invalid timestamp', updatedAt: 'invalid' },
        chapter,
      ],
      total: 5,
    })
    const wrapper = await render()
    await flushPromises()
    expect(wrapper.get('h2').text()).toBe('Chapter 07')
    expect(wrapper.text()).toContain('Page 5 of 20')
    expect(wrapper.text()).toContain('Last page reached')
    expect(wrapper.text()).not.toContain('Unread chapter')
    expect(wrapper.text()).not.toContain('Empty archive')
    expect(wrapper.text()).not.toContain('Invalid timestamp')
    expect(wrapper.findAll('a').map((link) => link.attributes('href'))).toContain(
      '/reader?manga=7&page=4&total=20',
    )
    expect(wrapper.findAll('a').map((link) => link.attributes('href'))).toContain(
      '/reader?manga=2&page=19&total=20',
    )
    await wrapper.get('img').trigger('error')
    expect(wrapper.findAll('img')).toHaveLength(1)
    wrapper.unmount()
  })

  it('retries a failed request and includes chapters saved on page zero', async () => {
    fetchHistory
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce({ items: [{ ...chapter, currentPage: 0 }], total: 1 })
    const wrapper = await render()
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('Reading history is unavailable')
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('Page 1 of 20')
    expect(fetchHistory).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })

  it('limits history to 24 chapters and uses a stable order for equal timestamps', async () => {
    fetchHistory.mockResolvedValue({
      items: Array.from({ length: 30 }, (_, i) => ({
        ...chapter,
        id: i + 1,
        title: `Chapter ${i + 1}`,
      })),
      total: 30,
    })
    const wrapper = await render()
    await flushPromises()
    expect(wrapper.get('h2').text()).toBe('Chapter 30')
    expect(wrapper.findAll('a[href^="/reader?"]')).toHaveLength(24)
    expect(wrapper.findAll('h3').at(-1)?.text()).toBe('Chapter 7')
    wrapper.unmount()
  })
})
