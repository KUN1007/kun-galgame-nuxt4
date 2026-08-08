import { z } from 'zod'
import { KUN_TODO_TYPE_CONST } from '~/constants/update'

export const getTodoSchema = z.object({
  page: z.coerce.number<number>().min(1).max(9999999),
  limit: z.coerce.number<number>().min(1).max(30),
  status: z.coerce.number<number>().min(0).max(3).optional()
})

export const createTodoSchema = z.object({
  status: z.coerce.number<number>().min(0).max(10, '待办状态应该为数字'),
  type: z.enum(KUN_TODO_TYPE_CONST),
  content_en_us: z
    .string()
    .max(1000, '英语待办最多 1000 个字符')
    .optional()
    .default(''),
  content_ja_jp: z
    .string()
    .max(1000, '日语待办最多 1000 个字符')
    .optional()
    .default(''),
  content_zh_cn: z
    .string()
    .max(1000, '简体中文待办最多 1000 个字符')
    .optional()
    .default(''),
  content_zh_tw: z
    .string()
    .max(1000, '繁体中文待办最多 1000 个字符')
    .optional()
    .default('')
})

export const updateTodoSchema = createTodoSchema.merge(
  z.object({
    todo_id: z.coerce.number<number>().min(1).max(9999999)
  })
)
