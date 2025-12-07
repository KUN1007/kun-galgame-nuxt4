import { prisma } from '~~/prisma/prisma'
import { getTodoSchema } from '~/validations/todo'
import type { UpdateTodo } from '~/types/api/update-log'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getTodoSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { page, limit } = input

  const skip = (page - 1) * limit
  const totalCount = await prisma.todo.count()

  const data = await prisma.todo.findMany({
    skip,
    take: limit,
    orderBy: { created: 'desc' }
  })

  const todos: UpdateTodo[] = data.map((todo) => ({
    id: todo.id,
    status: todo.status,
    type: todo.type,
    content: {
      'en-us': todo.content_en_us,
      'ja-jp': todo.content_ja_jp,
      'zh-cn': todo.content_zh_cn,
      'zh-tw': todo.content_zh_tw
    },
    completedTime: todo.completed_time,
    created: todo.created
  }))

  return { todos, totalCount }
})
