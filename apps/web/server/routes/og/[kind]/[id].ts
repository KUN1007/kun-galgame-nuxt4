import {
  KUN_OG_CARD_KINDS,
  type KunOgCardKind
} from '../../../../shared/utils/ogCard'
import { kunOgEntityCard } from '../../../utils/kunOgCard'
import { kunOgRedirect } from '../../../utils/kunOgRedirect'

export default defineEventHandler(async (event) => {
  const kind = getRouterParam(event, 'kind') as KunOgCardKind
  const id = Number(getRouterParam(event, 'id'))
  const known =
    KUN_OG_CARD_KINDS.includes(kind) && Number.isInteger(id) && id > 0

  return kunOgRedirect(event, known ? await kunOgEntityCard(kind, id) : null)
})
