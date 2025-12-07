import {
  MAX_LARGE_FILE_SIZE,
  MOEMOEPOINT_SINGLE_MB_DIVISOR,
  USER_DAILY_UPLOAD_LIMIT
} from '~/config/upload'
import { prisma } from '~~/prisma/prisma'

const MB = 1024 * 1024

export const canUserUpload = async (user_id: number, filesize: number) => {
  const user = await prisma.user.findUnique({
    where: { id: user_id },
    select: { role: true, moemoepoint: true, daily_toolset_upload_count: true }
  })
  if (!user) {
    return '未找到该用户'
  }

  const isAdmin = user.role > 1
  const dailyUsed = user.daily_toolset_upload_count
  const dailyRemaining = isAdmin
    ? MAX_LARGE_FILE_SIZE
    : Math.max(0, USER_DAILY_UPLOAD_LIMIT - dailyUsed)
  const moemoepointDeltaForDailyFileSize =
    dailyRemaining + user.moemoepoint * MB

  const moemoepointMaxSingleFile =
    Math.floor(user.moemoepoint / MOEMOEPOINT_SINGLE_MB_DIVISOR) * MB
  const maxSingleFileForUser = isAdmin
    ? MAX_LARGE_FILE_SIZE
    : Math.min(
        Math.max(USER_DAILY_UPLOAD_LIMIT, moemoepointMaxSingleFile),
        MAX_LARGE_FILE_SIZE
      )

  if (filesize > moemoepointDeltaForDailyFileSize) {
    return `超出当日可用上传额度, 剩余 ${(moemoepointDeltaForDailyFileSize / MB).toFixed(2)} MB`
  }
  if (filesize > maxSingleFileForUser) {
    const limitMB = (maxSingleFileForUser / MB).toFixed(2)
    return `单文件大小超过限制, 最大 ${limitMB} MB`
  }

  return user.daily_toolset_upload_count + filesize
}
