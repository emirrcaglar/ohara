import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import VaultCard from '../components/VaultCard.vue'
import type { MangaRow } from '../types/api'

describe('manga release cards', () => {
  it('shows title and chapter metadata with original filename details without opening the reader', async () => {
    const manga: MangaRow = {
      id: 1,
      catalogId: null,
      title: '[Group] Title v03 c042.5 (Digital)',
      path: '',
      seriesTitle: 'Title',
      fileName: '[Group] Title v03 c042.5 (Digital).cbz',
      parserVersion: 'test',
      pageCount: 20,
      currentPage: 0,
      fileExtension: '.cbz',
      updatedAt: null,
      metadata: {
        volume: { start: { whole: 3, decimal: null }, end: null },
        chapter: { start: { whole: 42, decimal: 5 }, end: null },
        group: 'Group',
        source: 'Digital',
        language: 'English',
        revision: null,
        isOneshot: false,
        edition: '',
      },
    }
    const wrapper = mount(VaultCard, { props: { manga } })
    expect(wrapper.get('h3').text()).toBe('Title')
    expect(wrapper.text()).toContain('Vol. 3 · Ch. 42.5 · Digital')
    expect(wrapper.get('details').text()).toContain(manga.fileName)
    expect(wrapper.get('details').text()).toContain('English')
    await wrapper.get('summary').trigger('click')
    expect(wrapper.emitted('click')).toBeUndefined()
    await wrapper.setProps({ manga: { ...manga, seriesTitle: '', metadata: null } })
    expect(wrapper.get('h3').text()).toBe(manga.title)
    expect(wrapper.find('details').exists()).toBe(false)
    wrapper.unmount()
  })
})
