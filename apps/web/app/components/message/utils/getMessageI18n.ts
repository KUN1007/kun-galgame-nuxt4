const messageTemplates: Partial<Record<MessageType, string>> = {
  upvoted: ' 推了您!',
  liked: ' 点赞了您!',
  favorite: ' 收藏了您!',
  replied: ' 回复了您!',
  commented: ' 评论了您!',
  expired: ' 报告了您的资源链接已过期！',
  requested: ' 向您提出更新请求！',
  solution: '您的回复被标记为最佳答案!',
  merged: ' 合并了您的更新请求！',
  declined: ' 拒绝了您的更新请求！',
  admin: '系统消息',
  mentioned: ' 提到了您！',
  'quiz-answered': ' 回答了您的题目!',
  'lottery-won': ' 的抽奖开奖了, 您中奖了!',
  'lottery-closed': ' 的抽奖开奖了',
  'lottery-expired': ' 的抽奖兑换码已过领取期限',
  'poll-closed': ' 的投票已经截止'
}

export const getMessageI18n = (message: Message) => {
  if (message.type === 'mentioned' && message.content?.trim()) {
    return messageTemplates.replied ?? ''
  }
  return messageTemplates[message.type] ?? ''
}
