import { prisma } from '~~/prisma/prisma'

export default defineTask({
  meta: {
    name: 'reset-daily',
    description:
      'Resets daily user stats like check-in and image counts at midnight.'
  },

  async run() {
    const { count } = await prisma.user.updateMany({
      data: {
        daily_check_in: 0,
        daily_image_count: 0,
        daily_toolset_upload_count: 0
      },
      where: {
        OR: [
          { daily_check_in: { not: 0 } },
          { daily_image_count: { not: 0 } },
          { daily_toolset_upload_count: { not: 0 } }
        ]
      }
    })

    return { result: `Reset stats for ${count} users.` }
  }
})
