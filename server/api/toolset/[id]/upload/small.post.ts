import { prisma } from '~~/prisma/prisma'
import { s3 } from '~/lib/s3/client'
import { getSignedUrl } from '@aws-sdk/s3-request-presigner'
import { generateRandomCode } from '~~/server/utils/generateRandomCode'
import {
  MAX_SMALL_FILE_SIZE,
  KUN_VISUAL_NOVEL_UPLOAD_TIMEOUT_LIMIT
} from '~/config/upload'
import { initToolsetUploadSchema } from '~/validations/toolset'
import { parseFileName } from '~~/server/utils/upload/parseFileName'
import { saveUploadSalt } from '~~/server/utils/upload/saveUploadSalt'
import { canUserUpload } from '~~/server/utils/upload/canUserUpload'
import type { ToolsetSmallFileUploadResponse } from '~/types/api/toolset'
import { PutObjectCommand } from '@aws-sdk/client-s3'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const parsed = initToolsetUploadSchema.safeParse(body)
  if (!parsed.success) {
    return kunError(event, parsed.error.message)
  }

  const { toolsetId, filename, filesize } = parsed.data

  if (filesize > MAX_SMALL_FILE_SIZE) {
    return kunError(event, '文件过大，请使用分片上传')
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '请先登录', 205)
  }

  const toolset = await prisma.galgame_toolset.findUnique({
    where: { id: toolsetId },
    select: { id: true }
  })
  if (!toolset) {
    return kunError(event, '工具不存在')
  }

  const result = await canUserUpload(userInfo.uid, filesize)
  if (typeof result === 'string') {
    return kunError(event, result)
  }

  const { base, ext } = parseFileName(filename)
  const salt = generateRandomCode(7)
  const key = `toolset/${toolsetId}/${userInfo.uid}_${base}_${salt}.${ext}`

  // Some S3 compatible providers not supported createPresignedPost
  // const post = await createPresignedPost(s3, {
  //   Bucket: process.env.KUN_VISUAL_NOVEL_S3_STORAGE_BUCKET_NAME!,
  //   Key: key,
  //   Conditions: [
  //     ['content-length-range', filesize, filesize],
  //     ['eq', '$Content-Type', 'application/octet-stream']
  //   ],
  //   Fields: { 'Content-Type': 'application/octet-stream' },
  //   Expires: 3600
  // })

  const command = new PutObjectCommand({
    Bucket: process.env.KUN_VISUAL_NOVEL_S3_STORAGE_BUCKET_NAME!,
    Key: key,
    ContentType: 'application/octet-stream'
  })

  const url = await getSignedUrl(s3, command, {
    expiresIn: KUN_VISUAL_NOVEL_UPLOAD_TIMEOUT_LIMIT
  })

  await saveUploadSalt(key, 'small', salt, filesize, base, ext)

  const response: ToolsetSmallFileUploadResponse = {
    key,
    salt,
    url
  }

  return response
})
