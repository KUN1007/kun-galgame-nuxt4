import { prisma } from '~~/prisma/prisma'
import env from '~~/server/env/dotenv'
import sharp from 'sharp'
import { uploadImageToS3 } from '~/lib/s3/uploadImageToS3'
import {
  KUN_VISUAL_NOVEL_IMAGE_COMPRESS_QUALITY,
  KUN_VISUAL_NOVEL_IMAGE_COMPRESS_LIMIT
} from '~/config/upload'
import { checkBufferSize } from '~~/server/utils/checkBufferSize'

const resizeUserAvatar = async (name: string, avatar: Buffer, uid: number) => {
  const miniAvatar = await sharp(avatar)
    .resize(100, 100, {
      fit: 'inside',
      withoutEnlargement: true
    })
    .webp({ quality: KUN_VISUAL_NOVEL_IMAGE_COMPRESS_QUALITY })
    .toBuffer()

  if (!checkBufferSize(miniAvatar, KUN_VISUAL_NOVEL_IMAGE_COMPRESS_LIMIT)) {
    return `图片压缩后大小超过 ${KUN_VISUAL_NOVEL_IMAGE_COMPRESS_LIMIT} MB`
  }

  const miniAvatarName = `${name}-100`
  const bucketName = `avatar/user_${uid}`

  await uploadImageToS3(`${bucketName}/${name}.webp`, avatar)
  await uploadImageToS3(`${bucketName}/${miniAvatarName}.webp`, miniAvatar)
}

export default defineEventHandler(async (event) => {
  const avatarFile = await readMultipartFormData(event)
  if (!avatarFile || !Array.isArray(avatarFile)) {
    return kunError(event, '读取图片错误')
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const res = await resizeUserAvatar('avatar', avatarFile[0].data, userInfo.uid)
  if (res) {
    return kunError(event, res)
  }

  const imageLink = `${env.KUN_VISUAL_NOVEL_IMAGE_BED_URL}/avatar/user_${userInfo.uid}/avatar.webp`
  await prisma.user.update({
    where: { id: userInfo.uid },
    data: { avatar: imageLink }
  })

  await purgeCache('userAvatar', userInfo.uid)

  return imageLink
})
