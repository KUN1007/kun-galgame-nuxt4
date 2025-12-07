import sharp from 'sharp'
import env from '~~/server/env/dotenv'
import { uploadImageToS3 } from '~/lib/s3/uploadImageToS3'
import {
  KUN_VISUAL_NOVEL_IMAGE_COMPRESS_QUALITY,
  KUN_VISUAL_NOVEL_IMAGE_COMPRESS_LIMIT,
  KUN_VISUAL_NOVEL_USER_DAILY_UPLOAD_IMAGE_LIMIT
} from '~/config/upload'
import { checkBufferSize } from '~~/server/utils/checkBufferSize'
import { prisma } from '~~/prisma/prisma'

const compressImage = async (name: string, image: Buffer, uid: number) => {
  const miniImage = await sharp(image)
    .resize(1920, 1080, {
      fit: 'inside',
      withoutEnlargement: true
    })
    .webp({ quality: KUN_VISUAL_NOVEL_IMAGE_COMPRESS_QUALITY })
    .toBuffer()

  if (!checkBufferSize(miniImage, KUN_VISUAL_NOVEL_IMAGE_COMPRESS_LIMIT)) {
    return `图片压缩后大小超过 ${KUN_VISUAL_NOVEL_IMAGE_COMPRESS_LIMIT} MB`
  }

  const bucketName = `topic/user_${uid}`

  await uploadImageToS3(`${bucketName}/${name}.webp`, miniImage)
}

export default defineEventHandler(async (event) => {
  const imageFile = await readMultipartFormData(event)
  if (!imageFile || !Array.isArray(imageFile)) {
    return kunError(event, '读取图片数据错误')
  }
  if (!checkBufferSize(imageFile[0].data, 10)) {
    return kunError(event, '图片大小应该小于 10MB')
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const user = await prisma.user.findUnique({
    where: { id: userInfo.uid }
  })
  if (!user) {
    return kunError(event, '未找到用户')
  }
  if (
    user.daily_image_count >= KUN_VISUAL_NOVEL_USER_DAILY_UPLOAD_IMAGE_LIMIT
  ) {
    return kunError(
      event,
      `您今日上传图片已达到 ${KUN_VISUAL_NOVEL_USER_DAILY_UPLOAD_IMAGE_LIMIT} 张上限`
    )
  }

  const newFileName = `${userInfo.name}-${Date.now()}`

  const res = await compressImage(newFileName, imageFile[0].data, userInfo.uid)
  if (res) {
    return kunError(event, res)
  }

  await prisma.user.update({
    where: { id: userInfo.uid },
    data: { daily_image_count: { increment: 1 } }
  })

  const imageLink = `${env.KUN_VISUAL_NOVEL_IMAGE_BED_URL}/topic/user_${userInfo.uid}/${newFileName}.webp`
  return imageLink
})
