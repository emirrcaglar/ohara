const USER_PFP_MODULES = import.meta.glob<string>('../assets/user-pfp/*.png', {
  eager: true,
  query: '?url',
  import: 'default',
})

export const USER_PFPS = Object.entries(USER_PFP_MODULES)
  .sort(([a], [b]) => a.localeCompare(b))
  .map(([, src]) => src)

export const USER_PFP_OPTIONS = USER_PFPS.map((src, index) => ({
  index,
  src,
  alt: `Avatar ${index + 1}`,
}))

export function normalizePfpIndex(pfp: number | null | undefined) {
  if (!USER_PFPS.length) return 0
  if (typeof pfp !== 'number' || !Number.isFinite(pfp)) return 0

  const index = Math.trunc(pfp)
  return ((index % USER_PFPS.length) + USER_PFPS.length) % USER_PFPS.length
}

export function getUserPfpUrl(pfp: number | null | undefined) {
  return USER_PFPS[normalizePfpIndex(pfp)] ?? ''
}
