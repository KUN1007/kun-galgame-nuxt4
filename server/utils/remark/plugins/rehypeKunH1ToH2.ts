// h1 is not require in render process

import { visit } from 'unist-util-visit'
import type { Plugin } from 'unified'
import type { Element } from 'hast'

export const rehypeKunH1ToH2: Plugin<[], Element> = () => {
  return (tree) => {
    visit(tree, 'element', (node: Element) => {
      if (node.tagName === 'h1') {
        node.tagName = 'h2'
      }
    })
  }
}
