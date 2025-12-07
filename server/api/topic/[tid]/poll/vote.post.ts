import { prisma } from '~~/prisma/prisma'
import { updateUserVoteSchema } from '~/validations/topic-poll'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }

  const input = await kunParsePostBody(event, updateUserVoteSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const { poll_id, option_id_array } = input

  const poll = await prisma.topic_poll.findUnique({
    where: { id: poll_id }
  })
  if (!poll) {
    return kunError(event, '投票不存在')
  }
  if (poll.status === 'closed') {
    return kunError(event, '投票已被关闭')
  }
  if (poll.deadline && new Date(poll.deadline) < new Date()) {
    return kunError(event, '投票已过截止日期')
  }
  if (poll.type === 'single' && option_id_array.length !== 1) {
    return kunError(event, '单选投票只能选择一个选项')
  }
  if (poll.type === 'multiple') {
    if (poll.min_choice && option_id_array.length < poll.min_choice) {
      return kunError(event, `这个投票至少需要选择 ${poll.min_choice} 个选项`)
    }
    if (poll.max_choice && option_id_array.length > poll.max_choice) {
      return kunError(event, `这个投票最多只能选择 ${poll.max_choice} 个选项`)
    }
  }

  const userPreviousVotes = await prisma.topic_poll_vote.findMany({
    where: { poll_id: poll_id, user_id: userInfo.uid }
  })

  if (userPreviousVotes.length > 0 && !poll.can_change_vote) {
    return kunError(event, '该投票不允许修改结果')
  }

  await prisma.$transaction(async (prisma) => {
    await prisma.topic_poll_vote.deleteMany({
      where: {
        poll_id: poll_id,
        user_id: userInfo.uid
      }
    })
    await prisma.topic_poll_vote.createMany({
      data: option_id_array.map((option_id) => ({
        poll_id: poll_id,
        option_id,
        user_id: userInfo.uid
      }))
    })
  })

  return 'Moemoe vote successfully!'
})
