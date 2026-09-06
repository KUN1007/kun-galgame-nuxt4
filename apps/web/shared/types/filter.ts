// The filter bar shared by /galgame, /gallib and the entity detail lists.

export interface FilterOption {
  value: string
  label: string
  /** Drawn right-aligned and dimmer: a count, a short example, a caveat. */
  hint?: string
}

// One entry of the bar's summary row. `key` is what @remove hands back, so it
// has to name the value and not only the dimension it came from.
export interface FilterChip {
  key: string
  label: string
  /** The dimension, drawn dimmer in front of the value. */
  prefix?: string
}
