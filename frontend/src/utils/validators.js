import { z } from 'zod';

export const QUESTION_TYPES = {
  SINGLE: 'single',
  MULTIPLE: 'multiple',
  JUDGE: 'judge',
  BLANK: 'blank',
};

export const QUESTION_TYPE_LABELS = {
  single: '单选题',
  multiple: '多选题',
  judge: '判断题',
  blank: '填空题',
};

export const BLANK_MATCH_MODES = {
  EXACT: 'exact',
  IGNORE_CASE: 'ignore_case',
  REGEX: 'regex',
};

export const BLANK_MATCH_MODE_LABELS = {
  exact: '精确匹配',
  ignore_case: '忽略大小写',
  regex: '正则匹配',
};

export const MULTIPLE_SCORING_MODES = {
  ALL_OR_NOTHING: 'all_or_nothing',
  PARTIAL: 'partial',
};

export const MULTIPLE_SCORING_LABELS = {
  all_or_nothing: '全对得分',
  partial: '部分得分',
};

export const loginSchema = z.object({
  username: z.string().min(1, '请输入用户名'),
  password: z.string().min(1, '请输入密码'),
});

export const registerSchema = z.object({
  username: z.string().min(3, '用户名至少3位').max(32, '用户名不能超过32位'),
  password: z.string().min(6, '密码至少6位').max(64, '密码不能超过64位'),
  classId: z.number({ invalid_type_error: '请选择班级' }).min(1, '请选择班级'),
});

export const questionOptionSchema = z.object({
  content: z.string().min(1, '选项内容不能为空').max(200, '选项不能超过200字符'),
  isCorrect: z.boolean(),
});

export const blankAnswerSchema = z.object({
  answer: z.string().min(1, '答案不能为空').max(500, '答案不能超过500字符'),
  matchMode: z.enum(['exact', 'ignore_case', 'regex']).optional(),
});

export const questionSchema = z
  .object({
    type: z.enum(['single', 'multiple', 'judge', 'blank'], {
      required_error: '请选择题型',
      invalid_type_error: '无效的题型',
    }),
    title: z.string().min(2, '题干至少2个字符').max(1000, '题干不能超过1000字符'),
    description: z.string().max(2000, '描述不能超过2000字符').optional().or(z.literal('')),
    options: z.array(questionOptionSchema).optional(),
    blankAnswers: z.array(blankAnswerSchema).optional(),
    multipleScore: z.enum(['all_or_nothing', 'partial']).optional(),
  })
  .superRefine((value, ctx) => {
    const { type } = value;

    if (type === 'single') {
      if (!value.options || value.options.length < 2 || value.options.length > 6) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: '单选题必须有2-6个选项',
          path: ['options'],
        });
        return;
      }
      const correctCount = value.options.filter((item) => item.isCorrect).length;
      if (correctCount !== 1) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: '单选题必须且仅能有1个正确答案',
          path: ['options'],
        });
      }
    }

    if (type === 'multiple') {
      if (!value.options || value.options.length < 2 || value.options.length > 6) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: '多选题必须有2-6个选项',
          path: ['options'],
        });
        return;
      }
      const correctCount = value.options.filter((item) => item.isCorrect).length;
      if (correctCount < 2) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: '多选题至少要有2个正确答案',
          path: ['options'],
        });
      }
    }

    if (type === 'judge') {
      if (!value.options || value.options.length !== 2) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: '判断题必须恰好有2个选项',
          path: ['options'],
        });
        return;
      }
      const correctCount = value.options.filter((item) => item.isCorrect).length;
      if (correctCount !== 1) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: '判断题必须且仅能有1个正确答案',
          path: ['options'],
        });
      }
    }

    if (type === 'blank') {
      if (!value.blankAnswers || value.blankAnswers.length < 1) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: '填空题至少要有1个标准答案',
          path: ['blankAnswers'],
        });
      }
    }
  });
