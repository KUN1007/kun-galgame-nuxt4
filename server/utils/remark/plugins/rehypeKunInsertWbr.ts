import { visit } from 'unist-util-visit'
import type { Root, Text, Element } from 'hast'

interface Options {
  minLength?: number
  chunkSize?: number
  splitChars?: string[]
  skipTags?: string[]
}

const makeWbrNode = (): Element => {
  return { type: 'element', tagName: 'wbr', properties: {}, children: [] }
}

const chunkAndPush = (
  str: string,
  size: number,
  out: Array<Text | Element>
) => {
  for (let i = 0; i < str.length; i += size) {
    out.push({ type: 'text', value: str.slice(i, i + size) })
    if (i + size < str.length) out.push(makeWbrNode())
  }
}

const escapeRegExp = (s: string) => {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

export const rehypeKunInsertWbr = (options: Options = {}) => {
  const {
    minLength = 30,
    chunkSize = 20,
    splitChars = ['_', '-', '.', '/'],
    skipTags = ['code', 'pre', 'style', 'script', 'textarea', 'input']
  } = options

  const splitRegex = new RegExp(
    `([${splitChars.map((c) => escapeRegExp(c)).join('')}])`,
    'g'
  )

  return (tree: Root) => {
    visit(tree, 'text', (node: Text, _idx, parent) => {
      if (
        !parent ||
        ((parent as Element).tagName &&
          skipTags.includes((parent as Element).tagName))
      )
        return

      const text = node.value

      if (!text || text.length < minLength) return

      const tokens = text.split(/(\s+)/)

      const newNodes: Array<Text | Element> = []

      for (const token of tokens) {
        if (!token || token.match(/^\s+$/)) {
          newNodes.push({ type: 'text', value: token })
          continue
        }

        if (token.length >= minLength && splitRegex.test(token)) {
          const parts = token.split(splitRegex)
          for (let i = 0; i < parts.length; i++) {
            const part = parts[i]
            if (!part) continue
            if (splitChars.includes(part)) {
              newNodes.push(makeWbrNode())
              newNodes.push({ type: 'text', value: part })

              newNodes.push(makeWbrNode())
            } else {
              if (part.length > chunkSize) {
                chunkAndPush(part, chunkSize, newNodes)
              } else {
                newNodes.push({ type: 'text', value: part })
              }
            }
          }
        } else if (token.length > chunkSize) {
          chunkAndPush(token, chunkSize, newNodes)
        } else {
          newNodes.push({ type: 'text', value: token })
        }
      }

      const onlyText =
        newNodes.length === 1 &&
        newNodes[0]!.type === 'text' &&
        (newNodes[0] as Text).value === text
      if (!onlyText) {
        const p = parent as Element
        if (!Array.isArray(p.children)) return

        const i = p.children.indexOf(node)
        if (i !== -1) {
          p.children.splice(i, 1, ...newNodes)
        }
      }
    })
  }
}
