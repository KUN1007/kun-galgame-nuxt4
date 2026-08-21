import { z } from 'zod'

const DOMAIN_RE =
  /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]?$/i

const domainField = z
  .string()
  .min(1, '网站主域名不能为空')
  .max(500, '网站主域名最多 500 个字符')
  .transform((value) =>
    value
      .trim()
      .replace(/^https?:\/\//i, '')
      .replace(/\/.*$/, '')
      .replace(/\.$/, '')
  )
  .refine((value) => DOMAIN_RE.test(value), {
    message: '无效的网站主域名 (示例: www.kungal.com)'
  })

const websiteBaseSchema = z.object({
  name: z
    .string()
    .min(1, '网站名称不能为空')
    .max(233, '网站名称最多 233 个字符'),
  url: domainField,
  description: z
    .string()
    .min(10, '网站介绍最少 10 个字符')
    .max(1000, '网站介绍最多 1000 个字符'),
  icon: z.string().max(500, '图标 URL 最多 500 个字符').optional().default(''),
  icon_image_hash: z
    .string()
    .max(128, '图标 hash 最多 128 个字符')
    .optional()
    .default(''),
  language: z.enum(['en-us', 'ja-jp', 'zh-cn', 'zh-tw']).default('zh-cn'),
  age_limit: z.enum(['all', 'r18']).default('all'),
  status: z.enum(['normal', 'unreachable', 'closed']).default('normal'),
  category_id: z.coerce.number<number>().min(1).max(9999999),
  tag_ids: z
    .array(z.coerce.number<number>().min(1).max(9999999))
    .max(20, '网站最多 20 个标签')
    .optional()
    .default([]),
  domain: z
    .array(z.string().max(100, '网站可用域名最多 100 个字符'))
    .max(10, '可用域名最多 10 个')
    .optional()
    .default([]),
  create_time: z.string().max(20, '网站创建时间描述最多 20 个字符').default('')
})

const hasIcon = (data: { icon?: string; icon_image_hash?: string }) =>
  !!(data.icon_image_hash || data.icon)
const ICON_REQUIRED = { message: '请上传网站图标' }

export const createWebsiteSchema = websiteBaseSchema.refine(
  hasIcon,
  ICON_REQUIRED
)

export const updateWebsiteSchema = websiteBaseSchema
  .extend({
    website_id: z.coerce.number<number>().min(1).max(9999999)
  })
  .refine(hasIcon, ICON_REQUIRED)

export const createWebsiteTagSchema = z.object({
  name: z.string().min(1, '标签名称不能为空').max(30, '标签名称最多 30 个字符'),
  label: z
    .string()
    .min(1, '标签显示名不能为空')
    .max(30, '标签显示名最多 30 个字符'),
  level: z.coerce
    .number()
    .int('标签等级必须是整数')
    .min(-100, '网站标签等级最小为 -100')
    .max(20, '网站标签等级最大为 20'),
  description: z.string().max(300, '网站标签描述最多 300 个字符').optional(),
  group_id: z.coerce
    .number<number>()
    .min(1)
    .max(9999999)
    .nullable()
    .default(null)
})

export const updateWebsiteTagSchema = createWebsiteTagSchema.extend({
  tag_id: z.coerce.number<number>().min(1).max(9999999)
})

export const createWebsiteCategorySchema = z.object({
  name: z
    .string()
    .min(1, '分类名称不能为空')
    .max(30, '分类名称最多 30 个字符')
    .regex(/^[a-z0-9_-]+$/, '分类名称是 URL 的一部分, 只能用小写字母/数字/-/_'),
  label: z
    .string()
    .min(1, '分类显示名不能为空')
    .max(30, '分类显示名最多 30 个字符'),
  description: z.string().max(300, '网站分类描述最多 300 个字符').optional(),
  sort_order: z.coerce.number<number>().min(0).max(9999).default(0)
})

export const updateWebsiteCategorySchema = createWebsiteCategorySchema.extend({
  category_id: z.coerce.number<number>().min(1).max(9999999)
})

export const createWebsiteTagGroupSchema = z.object({
  name: z
    .string()
    .min(1, '分组名称不能为空')
    .max(30, '分组名称最多 30 个字符')
    .regex(/^[a-z0-9_-]+$/, '分组名称只能用小写字母/数字/-/_'),
  label: z
    .string()
    .min(1, '分组显示名不能为空')
    .max(30, '分组显示名最多 30 个字符'),
  description: z.string().max(300, '分组描述最多 300 个字符').optional(),
  sort_order: z.coerce.number<number>().min(0).max(9999).default(0),
  multi_select: z.boolean().default(false)
})

export const updateWebsiteTagGroupSchema = createWebsiteTagGroupSchema.extend({
  group_id: z.coerce.number<number>().min(1).max(9999999)
})
