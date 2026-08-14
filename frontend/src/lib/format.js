const rtf = new Intl.RelativeTimeFormat('pt-BR', { numeric: 'auto' })

const UNITS = [
  { limit: 60, seconds: 1, unit: 'second' },
  { limit: 3600, seconds: 60, unit: 'minute' },
  { limit: 86400, seconds: 3600, unit: 'hour' },
  { limit: 604800, seconds: 86400, unit: 'day' },
  { limit: 2629800, seconds: 604800, unit: 'week' },
  { limit: 31557600, seconds: 2629800, unit: 'month' },
]

/** "há 5 minutos", "ontem", "há 2 meses". */
export function relativeTime(isoDate) {
  const date = new Date(isoDate)
  if (Number.isNaN(date.getTime())) return ''

  const elapsed = (Date.now() - date.getTime()) / 1000
  if (elapsed < 45) return 'agora mesmo'

  for (const { limit, seconds, unit } of UNITS) {
    if (elapsed < limit) return rtf.format(-Math.round(elapsed / seconds), unit)
  }
  return rtf.format(-Math.round(elapsed / 31557600), 'year')
}

/** Data completa, usada no title do elemento para quem quiser o valor exato. */
export function fullDate(isoDate) {
  const date = new Date(isoDate)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat('pt-BR', { dateStyle: 'short', timeStyle: 'short' }).format(date)
}
