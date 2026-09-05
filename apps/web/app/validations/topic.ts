import { z } from 'zod'
import { KUN_TOPIC_TITLE_LENGTH_LIMIT } from '~/config/limit'
import {
  KUN_TOPIC_ACCESS_ROLE_CONST,
  KUN_TOPIC_ACCESS_ROLE_LIMIT,
  KUN_TOPIC_ACCESS_SCOPE_CONST,
  KUN_TOPIC_ACCESS_USER_LIMIT,
  KUN_TOPIC_CATEGORY_CONST,
  KUN_TOPIC_SECTION_CONST,
  TOPIC_SORT_FIELD_CONST
} from '~/constants/topic'

const SORT_ORDER_CONST = ['asc', 'desc'] as const

export const getTopicSchema = z.object({
  page: z.coerce.number<number>().min(1).max(9999999),
  limit: z.coerce.number<number>().min(1).max(30),
  sort_field: z.enum(TOPIC_SORT_FIELD_CONST),
  sort_order: z.enum(SORT_ORDER_CONST),
  category: z.enum(KUN_TOPIC_CATEGORY_CONST)
})

export const createTopicSchema = z
  .object({
    title: z
      .string()
      .min(1, { message: '话题标题最少 1 个字符' })
      .max(KUN_TOPIC_TITLE_LENGTH_LIMIT, {
        message: `话题标题最大长度为 ${KUN_TOPIC_TITLE_LENGTH_LIMIT} 个字符`
      })
      .refine((t) => t.trim().length, { message: '话题标题最少为 1 个字符' }),
    content: z
      .string()
      .min(1, { message: '话题内容最少 1 个字符' })
      .max(100007, { message: '话题标题最大长度为 100007 个字符' })
      .refine((t) => t.trim().length, { message: '话题内容最少为 1 个字符' }),
    category: z.enum(KUN_TOPIC_CATEGORY_CONST),
    section: z
      .array(z.enum(KUN_TOPIC_SECTION_CONST))
      .min(1, { message: '您至少选择一个话题的分区' })
      .max(3, { message: '您至多选择三个话题的分区' }),
    is_nsfw: z.coerce.boolean({ message: '未找到话题的 NSFW 设置' }),
    cover_images: z
      .array(
        z
          .string()
          .regex(/^\/image\/[0-9a-f]{64}$/, { message: '封面图格式不正确' })
      )
      .max(9, { message: '封面图最多 9 张' })
      .optional(),
    access_scope: z.enum(KUN_TOPIC_ACCESS_SCOPE_CONST, {
      message: '无效的话题访问范围'
    }),
    access_roles: z
      .array(z.enum(KUN_TOPIC_ACCESS_ROLE_CONST))
      .max(KUN_TOPIC_ACCESS_ROLE_LIMIT, {
        message: `最多指定 ${KUN_TOPIC_ACCESS_ROLE_LIMIT} 个角色`
      }),
    access_user_ids: z
      .array(z.number().int().positive())
      .max(KUN_TOPIC_ACCESS_USER_LIMIT, {
        message: `最多指定 ${KUN_TOPIC_ACCESS_USER_LIMIT} 位用户`
      })
  })
  .superRefine((data, ctx) => {
    if (data.access_scope === 'role' && !data.access_roles.length) {
      ctx.addIssue({
        code: 'custom',
        message: '请至少选择一个可以看到本话题的角色',
        path: ['access_roles']
      })
    }
    if (data.access_scope === 'users' && !data.access_user_ids.length) {
      ctx.addIssue({
        code: 'custom',
        message: '请至少指定一位可以看到本话题的用户',
        path: ['access_user_ids']
      })
    }
  })

export const createReplySchema = z.object({
  topic_id: z.coerce.number<number>().min(1).max(9999999),
  content: z
    .string()
    .trim()
    .min(1, { message: '回复内容不能为空' })
    .max(10007, { message: '单条回复的最大长度为 10007 个字符' })
})

export const updateReplySchema = z.object({
  reply_id: z.coerce.number<number>().min(1).max(9999999),
  content: z
    .string()
    .trim()
    .min(1, { message: '回复内容不能为空' })
    .max(10007, { message: '单条回复的最大长度为 10007 个字符' })
})
