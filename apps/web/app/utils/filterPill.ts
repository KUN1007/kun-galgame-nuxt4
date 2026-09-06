import { kunControlSquareClasses } from '@kungal/ui-core'

// The filter bar's pill skin. KunSelect brings the size, radius, popup and
// listbox; this is only the colouring, and FilterTrigger — the year panel's
// hand-rolled trigger, the one control KunUI cannot express — reuses it so the
// two kinds of pill in the same row cannot drift apart.
export const filterPillClass = (active: boolean) =>
  cn(
    'border bg-transparent shadow-none transition-colors [&_svg]:text-inherit',
    active
      ? 'border-primary/40 bg-primary/10 text-primary font-medium'
      : 'border-default-200 text-default-600 hover:border-default-300 hover:bg-default-100'
  )

// The square sibling (排序方向). kunControlSquareClasses is KunUI's own table of
// icon-button sides, sized to match the text controls at the same size — a
// hand-picked size-8 renders 32px against the row's 34px.
export const filterPillSquareClass = (active: boolean) =>
  cn(
    'flex shrink-0 cursor-pointer items-center justify-center rounded-full',
    kunControlSquareClasses.sm,
    filterPillClass(active)
  )
