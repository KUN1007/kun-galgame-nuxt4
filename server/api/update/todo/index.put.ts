import { prisma } from '~~/prisma/prisma'
import { updateTodoSchema } from '~/validations/todo'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限更新待办')
  }

  const input = await kunParsePutBody(event, updateTodoSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { todoId, ...rest } = input

  await prisma.todo.update({
    where: { id: todoId },
    data: { ...rest, completed_time: rest.status === 2 ? new Date() : null }
  })

  return 'Moemoe update todo successfully!'
})
