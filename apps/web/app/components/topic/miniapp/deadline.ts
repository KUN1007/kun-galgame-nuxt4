// KunDatePicker is date-only — its mode is 'single' | 'range' and its
// valueFormat defaults to yyyy-MM-dd — but both mini-app schemas validate the
// deadline with z.iso.datetime(). Picking a date and pressing 发布 therefore did
// nothing at all: the schema rejected "2026-08-31" and the form bailed before
// any request went out. Widen the picked day to the instant it ends.
export const deadlineFromPicker = (picked?: string) => {
  if (!picked) {
    return undefined
  }
  const [year, month, day] = picked.split('-').map(Number)
  if (!year || !month || !day) {
    return undefined
  }
  return new Date(year, month - 1, day, 23, 59, 59).toISOString()
}

// The picker cannot read back what the API stores, so an existing deadline has
// to be narrowed to the same yyyy-MM-dd shape before the modal opens.
export const deadlineToPicker = (value?: string | Date | null) => {
  if (!value) {
    return undefined
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return undefined
  }
  const pad = (part: number) => String(part).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}
