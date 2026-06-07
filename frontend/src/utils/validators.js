import { z } from 'zod';

export const loginSchema = z.object({
  username: z.string().min(1, '请输入用户名'),
  password: z.string().min(1, '请输入密码'),
});

export const registerSchema = z.object({
  username: z.string().min(3, '用户名至少3位').max(32, '用户名不能超过32位'),
  password: z.string().min(6, '密码至少6位').max(64, '密码不能超过64位'),
  classId: z.number({ invalid_type_error: '请选择班级' }).min(1, '请选择班级'),
});

export const questionSchema = z
  .object({
    title: z.string().min(2, '题干至少2个字符').max(1000, '题干不能超过1000字符'),
    description: z.string().max(2000, '描述不能超过2000字符').optional(),
    options: z
      .array(
        z.object({
          content: z.string().min(1, '选项内容不能为空').max(200, '选项不能超过200字符'),
          isCorrect: z.boolean(),
        })
      )
      .min(2, '至少2个选项')
      .max(6, '最多6个选项'),
  })
  .superRefine((value, ctx) => {
    const correctCount = value.options.filter((item) => item.isCorrect).length;
    if (correctCount !== 1) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: '必须且仅能有1个正确答案',
        path: ['options'],
      });
    }
  });
