import { describe, expect, it } from 'vitest'
import { compareManga, mangaSubtitle, mangaTitle, matchesMangaSearch } from '../utils/mangaMetadata'
import type { MangaMetadata, MangaNumberRange, MangaRelease } from '../types/api'

function release(
  chapter: MangaNumberRange | null,
  fileName = 'original.cbz',
): MangaRelease & { title: string; metadata: MangaMetadata } {
  return {
    title: 'Original filename',
    seriesTitle: 'Title',
    fileName,
    parserVersion: 'test',
    metadata: {
      volume: null,
      chapter,
      group: 'Group',
      source: 'Digital',
      language: 'English',
      revision: null,
      isOneshot: false,
      edition: '',
    } satisfies MangaMetadata,
  } satisfies MangaRelease & { title: string }
}

const single = (whole: number, decimal: number | null = null): MangaNumberRange => ({
  start: { whole, decimal },
  end: null,
})

describe('manga release display and sorting', () => {
  it('sorts structured numbers and ranges using chaptr ordering, regardless of filenames', () => {
    const values = [
      release(single(10)),
      release(single(2, 10)),
      release(single(2, 5)),
      release(single(2, 0)),
      release(single(2)),
    ]
    expect(values.sort(compareManga).map((item) => mangaSubtitle(item))).toEqual([
      'Ch. 2 · Digital',
      'Ch. 2.0 · Digital',
      'Ch. 2.5 · Digital',
      'Ch. 2.10 · Digital',
      'Ch. 10 · Digital',
    ])
    const batch = release({ start: { whole: 2, decimal: null }, end: { whole: 10, decimal: null } })
    expect(compareManga(release(single(2)), batch)).toBeLessThan(0)
    expect(mangaSubtitle(batch)).toBe('Ch. 2–10 · Digital')
    const volumeFirst = release(single(100))
    volumeFirst.metadata.volume = single(1)
    const volumeSecond = release(single(1))
    volumeSecond.metadata.volume = single(2)
    expect(compareManga(volumeFirst, volumeSecond)).toBeLessThan(0)
    expect(compareManga(release(single(1), 'a.cbz'), release(single(1), 'b.cbz'))).toBeLessThan(0)
  })

  it('falls back to legacy titles and searches original filenames and parsed metadata', () => {
    const legacy = { title: 'Legacy c002', seriesTitle: '', metadata: null }
    expect(mangaTitle(legacy)).toBe('Legacy c002')
    expect(mangaSubtitle(legacy)).toBe('')
    expect(compareManga(legacy, { ...legacy, title: 'Legacy c010' })).toBeLessThan(0)
    const item = release(single(42, 5), '[Group] Title c042.5.cbz')
    expect(mangaTitle(item)).toBe('Title')
    expect(matchesMangaSearch(item, 'c042.5.cbz')).toBe(true)
    expect(matchesMangaSearch(item, 'title english group')).toBe(true)
    expect(matchesMangaSearch(item, 'missing')).toBe(false)
    expect(matchesMangaSearch(item, '   ')).toBe(true)
  })
})
