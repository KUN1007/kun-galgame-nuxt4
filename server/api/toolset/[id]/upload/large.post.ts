import { prisma } from '~~/prisma/prisma'
import { s3 } from '~/lib/s3/client'
import {
  CreateMultipartUploadCommand,
  UploadPartCommand
} from '@aws-sdk/client-s3'
import { getSignedUrl } from '@aws-sdk/s3-request-presigner'
import { generateRandomCode } from '~~/server/utils/generateRandomCode'
import {
  LARGE_FILE_CHUNK_SIZE,
  MAX_SMALL_FILE_SIZE,
  MAX_LARGE_FILE_SIZE,
  KUN_VISUAL_NOVEL_UPLOAD_TIMEOUT_LIMIT
} from '~/config/upload'
import { initToolsetUploadSchema } from '~/validations/toolset'
import { parseFileName } from '~~/server/utils/upload/parseFileName'
import { saveUploadSalt } from '~~/server/utils/upload/saveUploadSalt'
import { canUserUpload } from '~~/server/utils/upload/canUserUpload'
import type { ToolsetLargeFileUploadResponse } from '~/types/api/toolset'

const MB = 1024 * 1024

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, initToolsetUploadSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { toolsetId, filename, filesize } = input

  if (filesize <= MAX_SMALL_FILE_SIZE) {
    return kunError(
      event,
      `文件未超过 ${MAX_SMALL_FILE_SIZE / MB}MB, 请使用小文件上传接口`
    )
  }
  if (filesize > MAX_LARGE_FILE_SIZE) {
    return kunError(event, `单文件最大 ${MAX_LARGE_FILE_SIZE / (MB * 1024)}GB`)
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

  const bucket = process.env.KUN_VISUAL_NOVEL_S3_STORAGE_BUCKET_NAME!

  const createRes = await s3.send(
    new CreateMultipartUploadCommand({
      Bucket: bucket,
      Key: key,
      ContentType: 'application/octet-stream'
    })
  )
  if (!createRes.UploadId) {
    return kunError(event, '初始化分片上传失败，请稍后重试')
  }

  const partCount = Math.ceil(filesize / LARGE_FILE_CHUNK_SIZE)
  const urls: { partNumber: number; url: string }[] = []
  for (let i = 1; i <= partCount; i++) {
    const url = await getSignedUrl(
      s3,
      new UploadPartCommand({
        Bucket: bucket,
        Key: key,
        UploadId: createRes.UploadId,
        PartNumber: i
      }),
      { expiresIn: KUN_VISUAL_NOVEL_UPLOAD_TIMEOUT_LIMIT }
    )
    urls.push({ partNumber: i, url })
  }

  await saveUploadSalt(key, 'large', salt, filesize, base, ext)

  const response: ToolsetLargeFileUploadResponse = {
    key,
    salt,
    uploadId: createRes.UploadId,
    partSize: LARGE_FILE_CHUNK_SIZE,
    urls
  }

  return response
})
