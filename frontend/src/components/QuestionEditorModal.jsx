import { useEffect, useState } from 'react';
import {
  QUESTION_TYPES,
  QUESTION_TYPE_LABELS,
  BLANK_MATCH_MODES,
  BLANK_MATCH_MODE_LABELS,
  MULTIPLE_SCORING_MODES,
  MULTIPLE_SCORING_LABELS,
} from '../utils/validators';

function getDefaultState(type) {
  switch (type) {
    case QUESTION_TYPES.SINGLE:
      return {
        type,
        title: '',
        description: '',
        options: [
          { content: '', isCorrect: true },
          { content: '', isCorrect: false },
        ],
        blankAnswers: [],
        multipleScore: MULTIPLE_SCORING_MODES.ALL_OR_NOTHING,
      };
    case QUESTION_TYPES.MULTIPLE:
      return {
        type,
        title: '',
        description: '',
        options: [
          { content: '', isCorrect: true },
          { content: '', isCorrect: true },
          { content: '', isCorrect: false },
        ],
        blankAnswers: [],
        multipleScore: MULTIPLE_SCORING_MODES.ALL_OR_NOTHING,
      };
    case QUESTION_TYPES.JUDGE:
      return {
        type,
        title: '',
        description: '',
        options: [
          { content: '正确', isCorrect: true },
          { content: '错误', isCorrect: false },
        ],
        blankAnswers: [],
        multipleScore: MULTIPLE_SCORING_MODES.ALL_OR_NOTHING,
      };
    case QUESTION_TYPES.BLANK:
      return {
        type,
        title: '',
        description: '',
        options: [],
        blankAnswers: [{ answer: '', matchMode: BLANK_MATCH_MODES.EXACT }],
        multipleScore: MULTIPLE_SCORING_MODES.ALL_OR_NOTHING,
      };
    default:
      return {
        type: QUESTION_TYPES.SINGLE,
        title: '',
        description: '',
        options: [
          { content: '', isCorrect: true },
          { content: '', isCorrect: false },
        ],
        blankAnswers: [],
        multipleScore: MULTIPLE_SCORING_MODES.ALL_OR_NOTHING,
      };
  }
}

function buildInitialState(data) {
  if (!data) {
    return getDefaultState(QUESTION_TYPES.SINGLE);
  }

  const type = data.type || QUESTION_TYPES.SINGLE;
  return {
    type,
    title: data.title || '',
    description: data.description || '',
    options:
      data.options?.map((item) => ({
        id: item.id,
        content: item.content,
        isCorrect: item.isCorrect,
      })) || [],
    blankAnswers:
      data.blankAnswers?.map((item) => ({
        id: item.id,
        answer: item.answer,
        matchMode: item.matchMode || BLANK_MATCH_MODES.EXACT,
      })) || [],
    multipleScore: data.multipleScore || MULTIPLE_SCORING_MODES.ALL_OR_NOTHING,
  };
}

export function QuestionEditorModal({ open, initialData, onClose, onSubmit, loading }) {
  const [form, setForm] = useState(buildInitialState(initialData));

  useEffect(() => {
    if (open) {
      setForm(buildInitialState(initialData));
    }
  }, [open, initialData]);

  if (!open) {
    return null;
  }

  const handleTypeChange = (newType) => {
    if (newType === form.type) return;

    const defaultState = getDefaultState(newType);
    setForm((prev) => ({
      ...defaultState,
      title: prev.title,
      description: prev.description,
    }));
  };

  const updateOption = (index, patch) => {
    setForm((prev) => ({
      ...prev,
      options: prev.options.map((item, i) => (i === index ? { ...item, ...patch } : item)),
    }));
  };

  const setSingleCorrect = (index) => {
    setForm((prev) => ({
      ...prev,
      options: prev.options.map((item, i) => ({ ...item, isCorrect: i === index })),
    }));
  };

  const toggleMultipleCorrect = (index) => {
    setForm((prev) => ({
      ...prev,
      options: prev.options.map((item, i) =>
        i === index ? { ...item, isCorrect: !item.isCorrect } : item
      ),
    }));
  };

  const addOption = () => {
    setForm((prev) => ({
      ...prev,
      options: [...prev.options, { content: '', isCorrect: false }],
    }));
  };

  const removeOption = (index) => {
    setForm((prev) => ({
      ...prev,
      options: prev.options.filter((_, i) => i !== index),
    }));
  };

  const updateBlankAnswer = (index, patch) => {
    setForm((prev) => ({
      ...prev,
      blankAnswers: prev.blankAnswers.map((item, i) =>
        i === index ? { ...item, ...patch } : item
      ),
    }));
  };

  const addBlankAnswer = () => {
    setForm((prev) => ({
      ...prev,
      blankAnswers: [...prev.blankAnswers, { answer: '', matchMode: BLANK_MATCH_MODES.EXACT }],
    }));
  };

  const removeBlankAnswer = (index) => {
    setForm((prev) => ({
      ...prev,
      blankAnswers: prev.blankAnswers.filter((_, i) => i !== index),
    }));
  };

  const handleSubmit = (event) => {
    event.preventDefault();

    const payload = {
      type: form.type,
      title: form.title,
      description: form.description,
      options: form.options.map((item) => ({
        content: item.content,
        isCorrect: item.isCorrect,
      })),
      blankAnswers: form.blankAnswers.map((item) => ({
        answer: item.answer,
        matchMode: item.matchMode,
      })),
      multipleScore: form.multipleScore,
    };

    onSubmit(payload);
  };

  const isOptionType =
    form.type === QUESTION_TYPES.SINGLE ||
    form.type === QUESTION_TYPES.MULTIPLE ||
    form.type === QUESTION_TYPES.JUDGE;

  const canAddOption =
    (form.type === QUESTION_TYPES.SINGLE || form.type === QUESTION_TYPES.MULTIPLE) &&
    form.options.length < 6;

  const canRemoveOption =
    (form.type === QUESTION_TYPES.SINGLE || form.type === QUESTION_TYPES.MULTIPLE) &&
    form.options.length > 2;

  return (
    <div className="fixed inset-0 z-40 flex items-center justify-center bg-slate-900/50 px-4 py-8">
      <div className="w-full max-w-2xl rounded-2xl bg-base-100 p-6 shadow-2xl">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-xl font-semibold text-slate-800">
            {initialData ? '编辑题目' : '新增题目'}
          </h3>
          <button type="button" className="btn btn-sm btn-ghost" onClick={onClose}>
            关闭
          </button>
        </div>

        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="form-control">
            <span className="label-text mb-1 text-sm font-medium">题型</span>
            <div className="flex flex-wrap gap-2">
              {Object.entries(QUESTION_TYPE_LABELS).map(([value, label]) => (
                <button
                  key={value}
                  type="button"
                  className={`btn btn-sm ${form.type === value ? 'btn-primary' : 'btn-outline'}`}
                  onClick={() => handleTypeChange(value)}
                >
                  {label}
                </button>
              ))}
            </div>
          </div>

          <label className="form-control">
            <span className="label-text mb-1 text-sm font-medium">题干</span>
            <textarea
              className="textarea textarea-bordered min-h-24"
              value={form.title}
              onChange={(event) => setForm((prev) => ({ ...prev, title: event.target.value }))}
              required
            />
          </label>

          <label className="form-control">
            <span className="label-text mb-1 text-sm font-medium">知识点说明（可选）</span>
            <input
              className="input input-bordered"
              value={form.description}
              onChange={(event) =>
                setForm((prev) => ({ ...prev, description: event.target.value }))
              }
            />
          </label>

          {form.type === QUESTION_TYPES.MULTIPLE && (
            <label className="form-control">
              <span className="label-text mb-1 text-sm font-medium">评分方式</span>
              <select
                className="select select-bordered"
                value={form.multipleScore}
                onChange={(event) =>
                  setForm((prev) => ({ ...prev, multipleScore: event.target.value }))
                }
              >
                {Object.entries(MULTIPLE_SCORING_LABELS).map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
              <p className="mt-1 text-xs text-slate-500">
                全对得分：只有全部选对才得分；部分得分：选对部分按比例给分，选错不得分。
              </p>
            </label>
          )}

          {isOptionType && (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium text-slate-700">
                  {form.type === QUESTION_TYPES.JUDGE ? '判断选项' : '选项'}
                </p>
                {canAddOption && (
                  <button
                    type="button"
                    className="btn btn-xs btn-outline btn-primary"
                    onClick={addOption}
                  >
                    添加选项
                  </button>
                )}
              </div>

              {form.options.map((item, index) => (
                <div
                  key={`option-${index}`}
                  className="grid grid-cols-12 items-center gap-2 rounded-xl border border-slate-200 p-2"
                >
                  <div className="col-span-1 flex justify-center">
                    {form.type === QUESTION_TYPES.MULTIPLE ? (
                      <input
                        type="checkbox"
                        className="checkbox checkbox-success checkbox-sm"
                        checked={item.isCorrect}
                        onChange={() => toggleMultipleCorrect(index)}
                      />
                    ) : (
                      <input
                        type="radio"
                        name="correct"
                        className="radio radio-success radio-sm"
                        checked={item.isCorrect}
                        onChange={() => setSingleCorrect(index)}
                      />
                    )}
                  </div>
                  <div className="col-span-9">
                    <input
                      className="input input-bordered input-sm w-full"
                      value={item.content}
                      onChange={(event) =>
                        updateOption(index, { content: event.target.value })
                      }
                      disabled={form.type === QUESTION_TYPES.JUDGE}
                      required
                    />
                  </div>
                  <div className="col-span-2 flex justify-end">
                    {canRemoveOption && (
                      <button
                        type="button"
                        className="btn btn-xs btn-ghost text-error"
                        onClick={() => removeOption(index)}
                      >
                        删除
                      </button>
                    )}
                  </div>
                </div>
              ))}
              {form.type === QUESTION_TYPES.SINGLE && (
                <p className="text-xs text-slate-500">
                  绿色单选为正确答案，必须且只能有一个正确答案。
                </p>
              )}
              {form.type === QUESTION_TYPES.MULTIPLE && (
                <p className="text-xs text-slate-500">
                  绿色勾选为正确答案，至少需要2个正确答案。
                </p>
              )}
              {form.type === QUESTION_TYPES.JUDGE && (
                <p className="text-xs text-slate-500">
                  判断题固定为两个选项，绿色单选为正确答案。
                </p>
              )}
            </div>
          )}

          {form.type === QUESTION_TYPES.BLANK && (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium text-slate-700">标准答案</p>
                <button
                  type="button"
                  className="btn btn-xs btn-outline btn-primary"
                  onClick={addBlankAnswer}
                >
                  添加答案
                </button>
              </div>

              {form.blankAnswers.map((item, index) => (
                <div
                  key={`blank-${index}`}
                  className="grid grid-cols-12 items-center gap-2 rounded-xl border border-slate-200 p-2"
                >
                  <div className="col-span-7">
                    <input
                      className="input input-bordered input-sm w-full"
                      placeholder="标准答案"
                      value={item.answer}
                      onChange={(event) =>
                        updateBlankAnswer(index, { answer: event.target.value })
                      }
                      required
                    />
                  </div>
                  <div className="col-span-3">
                    <select
                      className="select select-bordered select-sm w-full"
                      value={item.matchMode}
                      onChange={(event) =>
                        updateBlankAnswer(index, { matchMode: event.target.value })
                      }
                    >
                      {Object.entries(BLANK_MATCH_MODE_LABELS).map(([value, label]) => (
                        <option key={value} value={value}>
                          {label}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="col-span-2 flex justify-end">
                    <button
                      type="button"
                      className="btn btn-xs btn-ghost text-error"
                      onClick={() => removeBlankAnswer(index)}
                      disabled={form.blankAnswers.length <= 1}
                    >
                      删除
                    </button>
                  </div>
                </div>
              ))}
              <p className="text-xs text-slate-500">
                匹配到任一答案即算正确。系统会自动去除前后空格并将全角字符转为半角。
              </p>
            </div>
          )}

          <div className="flex justify-end gap-2">
            <button type="button" className="btn btn-ghost" onClick={onClose}>
              取消
            </button>
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? '保存中...' : '保存题目'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
