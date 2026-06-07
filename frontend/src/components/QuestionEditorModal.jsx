import { useEffect, useState } from 'react';

function buildInitialState(data) {
  if (!data) {
    return {
      title: '',
      description: '',
      options: [
        { content: '', isCorrect: true },
        { content: '', isCorrect: false },
      ],
    };
  }

  return {
    title: data.title || '',
    description: data.description || '',
    options:
      data.options?.map((item) => ({
        id: item.id,
        content: item.content,
        isCorrect: item.isCorrect,
      })) || [],
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

  const updateOption = (index, patch) => {
    setForm((prev) => ({
      ...prev,
      options: prev.options.map((item, i) => (i === index ? { ...item, ...patch } : item)),
    }));
  };

  const setCorrect = (index) => {
    setForm((prev) => ({
      ...prev,
      options: prev.options.map((item, i) => ({ ...item, isCorrect: i === index })),
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

  const handleSubmit = (event) => {
    event.preventDefault();
    onSubmit({
      title: form.title,
      description: form.description,
      options: form.options.map((item) => ({
        content: item.content,
        isCorrect: item.isCorrect,
      })),
    });
  };

  return (
    <div className="fixed inset-0 z-40 flex items-center justify-center bg-slate-900/50 px-4 py-8">
      <div className="w-full max-w-2xl rounded-2xl bg-base-100 p-6 shadow-2xl">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-xl font-semibold text-slate-800">{initialData ? '编辑题目' : '新增题目'}</h3>
          <button type="button" className="btn btn-sm btn-ghost" onClick={onClose}>
            关闭
          </button>
        </div>

        <form className="space-y-4" onSubmit={handleSubmit}>
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
              onChange={(event) => setForm((prev) => ({ ...prev, description: event.target.value }))}
            />
          </label>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <p className="text-sm font-medium text-slate-700">选项</p>
              <button
                type="button"
                className="btn btn-xs btn-outline btn-primary"
                onClick={addOption}
                disabled={form.options.length >= 6}
              >
                添加选项
              </button>
            </div>

            {form.options.map((item, index) => (
              <div key={`option-${index}`} className="grid grid-cols-12 items-center gap-2 rounded-xl border border-slate-200 p-2">
                <div className="col-span-1 flex justify-center">
                  <input
                    type="radio"
                    name="correct"
                    className="radio radio-success radio-sm"
                    checked={item.isCorrect}
                    onChange={() => setCorrect(index)}
                  />
                </div>
                <div className="col-span-9">
                  <input
                    className="input input-bordered input-sm w-full"
                    value={item.content}
                    onChange={(event) => updateOption(index, { content: event.target.value })}
                    required
                  />
                </div>
                <div className="col-span-2 flex justify-end">
                  <button
                    type="button"
                    className="btn btn-xs btn-ghost text-error"
                    onClick={() => removeOption(index)}
                    disabled={form.options.length <= 2}
                  >
                    删除
                  </button>
                </div>
              </div>
            ))}
            <p className="text-xs text-slate-500">绿色单选为正确答案，必须且只能有一个正确答案。</p>
          </div>

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
