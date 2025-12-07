import { prisma } from '~~/prisma/prisma'
import { createTodoSchema } from '~/validations/todo'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限创建网站待办')
  }

  const input = await kunParsePostBody(event, createTodoSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  await prisma.todo.create({
    data: {
      user_id: userInfo.uid,
      completed_time: input.status === 2 ? new Date() : null,
      ...input
    }
  })

  return 'Moemoe create todo successfully!'
})
