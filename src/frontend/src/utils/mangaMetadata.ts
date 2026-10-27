import type { MangaNumber, MangaNumberRange, MangaRelease } from '../types/api'
import { compareNaturalText } from './naturalSort'

type MangaDisplay = Partial<MangaRelease> & { title: string }

export function mangaTitle(item: MangaDisplay): string {
  return item.seriesTitle || item.title
}

export function matchesMangaSearch(item: MangaDisplay, query: string): boolean {
  const terms = query.toLocaleLowerCase().trim().split(/\s+/)
  const text = [
    item.title,
    item.seriesTitle,
    item.fileName,
    mangaSubtitle(item),
    item.metadata?.group,
    item.metadata?.language,
  ]
    .filter(Boolean)
    .join(' ')
    .toLocaleLowerCase()
  return terms.every((term) => text.includes(term))
}

function formatNumber(value: MangaNumber): string {
  return `${value.whole}${value.decimal == null ? '' : `.${value.decimal}`}`
}

function formatRange(value: MangaNumberRange): string {
  return formatNumber(value.start) + (value.end ? `–${formatNumber(value.end)}` : '')
}

export function mangaSubtitle(item: MangaDisplay): string {
  const metadata = item.metadata
  if (!metadata) return ''
  return [
    metadata.volume && `Vol. ${formatRange(metadata.volume)}`,
    metadata.chapter && `Ch. ${formatRange(metadata.chapter)}`,
    metadata.isOneshot && 'Oneshot',
    metadata.source,
    metadata.edition,
    metadata.revision != null && `Rev. ${metadata.revision}`,
  ]
    .filter(Boolean)
    .join(' · ')
}

function compareNumber(a: MangaNumber, b: MangaNumber): number {
  return a.whole - b.whole || compareOptional(a.decimal, b.decimal, (x, y) => x - y)
}

function compareOptional<T>(
  a: T | null | undefined,
  b: T | null | undefined,
  compare: (a: T, b: T) => number,
): number {
  if (a == null) return b == null ? 0 : -1
  if (b == null) return 1
  return compare(a, b)
}

function compareRange(a: MangaNumberRange, b: MangaNumberRange): number {
  return compareNumber(a.start, b.start) || compareOptional(a.end, b.end, compareNumber)
}

export function compareManga(a: MangaDisplay, b: MangaDisplay): number {
  return (
    compareNaturalText(mangaTitle(a), mangaTitle(b)) ||
    compareOptional(a.metadata?.volume, b.metadata?.volume, compareRange) ||
    compareOptional(a.metadata?.chapter, b.metadata?.chapter, compareRange) ||
    compareNaturalText(a.fileName || a.title, b.fileName || b.title)
  )
}
