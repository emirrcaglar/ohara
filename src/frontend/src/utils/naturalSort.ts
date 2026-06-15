const naturalCollator = new Intl.Collator(undefined, {
  sensitivity: 'base',
  numeric: true,
})

export function compareNaturalText(a: string, b: string): number {
  return naturalCollator.compare(a, b)
}

export function compareByNaturalText<T>(getText: (item: T) => string) {
  return (a: T, b: T): number => compareNaturalText(getText(a), getText(b))
}
