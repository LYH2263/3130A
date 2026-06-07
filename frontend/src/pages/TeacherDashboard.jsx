import { useEffect, useMemo, useState } from 'react';
import { toast } from 'react-hot-toast';

import {
  apiRequest,
  fetchCategories,
  fetchTags,
  fetchKnowledgePoints,
  fetchQuestions,
  fetchQuestion,
} from '../api/client';
import { QuestionEditorModal } from '../components/QuestionEditorModal';
import { StatCard } from '../components/StatCard';
import { questionSchema, QUESTION_TYPE_LABELS } from '../utils/validators';

function CategoryTreeNode({ category, selectedId, onSelect, onToggle, expandedIds }) {
  const hasChildren = category.children && category.children.length > 0;
  const isExpanded = expandedIds.includes(category.id);
  const isSelected = selectedId === category.id;

  return (
    <div>
      <div
        className={`flex cursor-pointer items-center gap-1 rounded-lg px-2 py-1.5 text-sm transition-colors ${
          isSelected
            ? 'bg-sky-100 text-sky-700 font-medium'
            : 'hover:bg-slate-100 text-slate-700'
        }`}
      >
        {hasChildren ? (
          <button
            type="button"
            className="flex h-5 w-5 items-center justify-center text-slate-400 hover:text-slate-600"
            onClick={(e) => {
              e.stopPropagation();
              onToggle(category.id);
            }}
          >
            <svg
              className={`h-3 w-3 transition-transform ${isExpanded ? 'rotate-90' : ''}`}
              viewBox="0 0 12 12"
              fill="currentColor"
            >
              <path d="M4 2l4 4-4 4V2z" />
            </svg>
          </button>
        ) : (
          <span className="w-5" />
        )}
        <span className="flex-1 truncate" onClick={() => onSelect(category.id)}>
          {category.name}
        </span>
      </div>
      {hasChildren && isExpanded && (
        <div className="ml-4 border-l border-slate-200 pl-2">
          {category.children.map((child) => (
            <CategoryTreeNode
              key={child.id}
              category={child}
              selectedId={selectedId}
              onSelect={onSelect}
              onToggle={onToggle}
              expandedIds={expandedIds}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function flattenCategories(categories, result = []) {
  for (const cat of categories) {
    result.push(cat);
    if (cat.children && cat.children.length > 0) {
      flattenCategories(cat.children, result);
    }
  }
  return result;
}

export function TeacherDashboard({ user, token, onLogout }) {
  const [overview, setOverview] = useState(null);
  const [questions, setQuestions] = useState([]);
  const [totalQuestions, setTotalQuestions] = useState(0);
  const [stats, setStats] = useState([]);
  const [attempts, setAttempts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [savingQuestion, setSavingQuestion] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingQuestion, setEditingQuestion] = useState(null);
  const [uploading, setUploading] = useState(false);

  const [categories, setCategories] = useState([]);
  const [tags, setTags] = useState([]);
  const [knowledgePoints, setKnowledgePoints] = useState([]);
  const [expandedCategoryIds, setExpandedCategoryIds] = useState([]);
  const [selectedCategoryId, setSelectedCategoryId] = useState(null);
  const [selectedTagIds, setSelectedTagIds] = useState([]);
  const [tagMode, setTagMode] = useState('or');
  const [keyword, setKeyword] = useState('');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);

  const topStats = useMemo(() => stats.slice(0, 12), [stats]);

  const loadDashboard = async () => {
    setLoading(true);
    try {
      const [overviewData, statData, attemptData, categoryData, tagData, kpData] = await Promise.all([
        apiRequest('/teacher/overview', { token }),
        apiRequest('/teacher/class-stats', { token }),
        apiRequest('/teacher/attempts?limit=50', { token }),
        fetchCategories(token),
        fetchTags(token),
        fetchKnowledgePoints(token),
      ]);
      setOverview(overviewData);
      setStats(statData);
      setAttempts(attemptData);
      setCategories(categoryData);
      setTags(tagData);
      setKnowledgePoints(kpData);
      if (categoryData && categoryData.length > 0) {
        setExpandedCategoryIds(categoryData.map((c) => c.id));
      }
    } catch (error) {
      toast.error(error.message || '加载教师看板失败');
    } finally {
      setLoading(false);
    }
  };

  const loadQuestions = async () => {
    try {
      const params = {
        keyword: keyword || undefined,
        categoryId: selectedCategoryId || undefined,
        tagIds: selectedTagIds.length > 0 ? selectedTagIds : undefined,
        tagMode,
        page,
        pageSize,
      };
      const result = await fetchQuestions(token, params);
      setQuestions(result.items || []);
      setTotalQuestions(result.total || 0);
    } catch (error) {
      toast.error(error.message || '加载题目失败');
    }
  };

  useEffect(() => {
    loadDashboard();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  useEffect(() => {
    if (!loading) {
      loadQuestions();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedCategoryId, selectedTagIds, tagMode, keyword, page, pageSize, loading]);

  const handleToggleCategory = (id) => {
    setExpandedCategoryIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    );
  };

  const handleSelectCategory = (id) => {
    setSelectedCategoryId((prev) => (prev === id ? null : id));
    setPage(1);
  };

  const handleToggleTag = (tagId) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((x) => x !== tagId) : [...prev, tagId]
    );
    setPage(1);
  };

  const handleSearch = (e) => {
    e.preventDefault();
    setPage(1);
    loadQuestions();
  };

  const handleSaveQuestion = async (payload) => {
    try {
      questionSchema.parse(payload);
      setSavingQuestion(true);
      if (editingQuestion) {
        await apiRequest(`/teacher/questions/${editingQuestion.id}`, {
          method: 'PUT',
          token,
          body: payload,
        });
        toast.success('题目已更新');
      } else {
        await apiRequest('/teacher/questions', {
          method: 'POST',
          token,
          body: payload,
        });
        toast.success('题目已创建');
      }
      setModalOpen(false);
      setEditingQuestion(null);
      await Promise.all([loadQuestions(), loadDashboard()]);
    } catch (error) {
      toast.error(error?.issues?.[0]?.message || error.message || '保存题目失败');
    } finally {
      setSavingQuestion(false);
    }
  };

  const handleDeleteQuestion = async (questionId) => {
    if (!window.confirm('确认删除该题目？')) {
      return;
    }
    try {
      await apiRequest(`/teacher/questions/${questionId}`, {
        method: 'DELETE',
        token,
      });
      toast.success('题目已删除');
      await Promise.all([loadQuestions(), loadDashboard()]);
    } catch (error) {
      toast.error(error.message || '删除失败');
    }
  };

  const handleUpload = async (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }
    const formData = new FormData();
    formData.append('file', file);

    try {
      setUploading(true);
      const data = await apiRequest('/teacher/questions/upload', {
        method: 'POST',
        token,
        body: formData,
        isForm: true,
      });
      toast.success(`导入成功，新增 ${data.count || 0} 题`);
      await Promise.all([loadQuestions(), loadDashboard()]);
    } catch (error) {
      toast.error(error.message || '上传失败');
    } finally {
      setUploading(false);
      event.target.value = '';
    }
  };

  const openCreateModal = () => {
    setEditingQuestion(null);
    setModalOpen(true);
  };

  const openEditModal = async (question) => {
    try {
      const fullQuestion = await fetchQuestion(token, question.id);
      setEditingQuestion(fullQuestion);
      setModalOpen(true);
    } catch (error) {
      toast.error(error.message || '加载题目详情失败');
    }
  };

  const totalPages = Math.ceil(totalQuestions / pageSize);

  return (
    <div className="min-h-screen bg-board px-4 py-6 md:px-8 md:py-8">
      <header className="mx-auto mb-6 flex max-w-7xl flex-col gap-3 rounded-3xl border border-white/70 bg-white/90 px-6 py-5 shadow-card md:flex-row md:items-center md:justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.25em] text-sky-700">Teacher Console</p>
          <h1 className="mt-1 text-2xl font-bold text-slate-800">教师机管理员面板</h1>
          <p className="text-sm text-slate-600">题库修改后学生机拉取即同步。</p>
        </div>
        <div className="flex gap-2">
          <button className="btn btn-outline btn-primary" onClick={loadDashboard}>
            刷新看板
          </button>
          <button className="btn btn-neutral" onClick={onLogout}>
            退出登录
          </button>
        </div>
      </header>

      {loading ? (
        <div className="mx-auto max-w-7xl">
          <div className="grid gap-4 md:grid-cols-4">
            {Array.from({ length: 4 }).map((_, idx) => (
              <div key={`skeleton-${idx}`} className="h-28 animate-pulse rounded-2xl bg-white/80" />
            ))}
          </div>
        </div>
      ) : (
        <main className="mx-auto grid max-w-7xl gap-5">
          <section className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <StatCard title="学生总数" value={overview?.studentCount ?? 0} />
            <StatCard title="班级数量" value={overview?.classCount ?? 0} />
            <StatCard title="题库题量" value={overview?.questionCount ?? 0} />
            <StatCard title="作答次数" value={overview?.attemptCount ?? 0} />
          </section>

          <section className="grid gap-5 lg:grid-cols-[240px_1fr]">
            <aside className="rounded-3xl border border-slate-200 bg-white p-4 shadow-card">
              <div className="mb-3 flex items-center justify-between">
                <h3 className="text-sm font-semibold text-slate-700">分类导航</h3>
                <button
                  className={`text-xs ${
                    selectedCategoryId ? 'text-sky-600 hover:text-sky-700' : 'text-slate-400'
                  }`}
                  onClick={() => {
                    setSelectedCategoryId(null);
                    setPage(1);
                  }}
                  disabled={!selectedCategoryId}
                >
                  全部
                </button>
              </div>
              <div className="space-y-1 max-h-[500px] overflow-auto pr-1">
                {categories.map((cat) => (
                  <CategoryTreeNode
                    key={cat.id}
                    category={cat}
                    selectedId={selectedCategoryId}
                    onSelect={handleSelectCategory}
                    onToggle={handleToggleCategory}
                    expandedIds={expandedCategoryIds}
                  />
                ))}
                {!categories.length ? (
                  <p className="text-xs text-slate-400 py-2">暂无分类</p>
                ) : null}
              </div>
            </aside>

            <article className="rounded-3xl border border-slate-200 bg-white p-5 shadow-card">
              <div className="mb-3 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                <h2 className="text-lg font-semibold text-slate-800">题库管理</h2>
                <div className="flex flex-wrap gap-2">
                  <label className="btn btn-outline btn-secondary">
                    {uploading ? '上传中...' : '上传 JSON 题库'}
                    <input
                      type="file"
                      className="hidden"
                      accept="application/json"
                      disabled={uploading}
                      onChange={handleUpload}
                    />
                  </label>
                  <button className="btn btn-primary" onClick={openCreateModal}>
                    新增题目
                  </button>
                </div>
              </div>

              <form onSubmit={handleSearch} className="mb-3 flex flex-wrap gap-2 items-center">
                <div className="flex-1 min-w-[200px]">
                  <input
                    type="text"
                    className="input input-bordered input-sm w-full"
                    placeholder="搜索关键词..."
                    value={keyword}
                    onChange={(e) => setKeyword(e.target.value)}
                  />
                </div>
                <button type="submit" className="btn btn-sm btn-primary">
                  搜索
                </button>
                <button
                  type="button"
                  className="btn btn-sm btn-ghost"
                  onClick={() => {
                    setKeyword('');
                    setSelectedTagIds([]);
                    setSelectedCategoryId(null);
                    setPage(1);
                  }}
                >
                  重置
                </button>
              </form>

              {tags.length > 0 && (
                <div className="mb-3">
                  <div className="mb-2 flex items-center gap-2">
                    <span className="text-xs text-slate-500">标签筛选：</span>
                    <div className="flex gap-1">
                      <button
                        className={`btn btn-xs ${
                          tagMode === 'or' ? 'btn-primary' : 'btn-ghost'
                        }`}
                        onClick={() => {
                          setTagMode('or');
                          setPage(1);
                        }}
                      >
                        或
                      </button>
                      <button
                        className={`btn btn-xs ${
                          tagMode === 'and' ? 'btn-primary' : 'btn-ghost'
                        }`}
                        onClick={() => {
                          setTagMode('and');
                          setPage(1);
                        }}
                      >
                        且
                      </button>
                    </div>
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {tags.map((tag) => (
                      <button
                        key={tag.id}
                        className={`badge badge-sm cursor-pointer transition-colors ${
                          selectedTagIds.includes(tag.id)
                            ? 'badge-primary badge-outline'
                            : 'badge-ghost hover:badge-primary/30'
                        }`}
                        onClick={() => handleToggleTag(tag.id)}
                      >
                        {tag.name}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              <div className="max-h-[400px] overflow-auto rounded-xl border border-slate-200">
                <table className="table table-sm">
                  <thead>
                    <tr>
                      <th>ID</th>
                      <th>题型</th>
                      <th>分类</th>
                      <th>题干</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {questions.map((question) => (
                      <tr key={question.id}>
                        <td className="font-mono text-xs">{question.id}</td>
                        <td>
                          <span className="badge badge-outline badge-sm">
                            {QUESTION_TYPE_LABELS[question.type] || '单选题'}
                          </span>
                        </td>
                        <td className="text-xs text-slate-500">
                          {question.categoryName || '-'}
                        </td>
                        <td className="max-w-sm truncate" title={question.title}>
                          {question.title}
                        </td>
                        <td>
                          <div className="flex gap-1">
                            <button
                              className="btn btn-xs btn-ghost"
                              onClick={() => openEditModal(question)}
                            >
                              编辑
                            </button>
                            <button
                              className="btn btn-xs btn-ghost text-error"
                              onClick={() => handleDeleteQuestion(question.id)}
                            >
                              删除
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                    {!questions.length ? (
                      <tr>
                        <td colSpan={5} className="text-center text-slate-500 py-8">
                          当前没有题目，请先新增或上传题库。
                        </td>
                      </tr>
                    ) : null}
                  </tbody>
                </table>
              </div>

              {totalPages > 1 && (
                <div className="mt-3 flex items-center justify-between">
                  <div className="text-xs text-slate-500">
                    共 {totalQuestions} 题，第 {page}/{totalPages} 页
                  </div>
                  <div className="flex gap-1">
                    <button
                      className="btn btn-xs btn-outline"
                      onClick={() => setPage((p) => Math.max(1, p - 1))}
                      disabled={page <= 1}
                    >
                      上一页
                    </button>
                    <button
                      className="btn btn-xs btn-outline"
                      onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                      disabled={page >= totalPages}
                    >
                      下一页
                    </button>
                    <select
                      className="select select-bordered select-xs w-20"
                      value={pageSize}
                      onChange={(e) => {
                        setPageSize(Number(e.target.value));
                        setPage(1);
                      }}
                    >
                      <option value={10}>10/页</option>
                      <option value={20}>20/页</option>
                      <option value={50}>50/页</option>
                      <option value={100}>100/页</option>
                    </select>
                  </div>
                </div>
              )}
            </article>
          </section>

          <section className="grid gap-5 lg:grid-cols-[1.2fr,0.8fr]">
            <article className="rounded-3xl border border-slate-200 bg-white p-5 shadow-card">
              <h2 className="mb-3 text-lg font-semibold text-slate-800">最近成绩同步</h2>
              <div className="max-h-[460px] overflow-auto rounded-xl border border-slate-200">
                <table className="table table-sm">
                  <thead>
                    <tr>
                      <th>学生</th>
                      <th>班级</th>
                      <th>成绩</th>
                      <th>时间</th>
                    </tr>
                  </thead>
                  <tbody>
                    {attempts.map((item) => (
                      <tr key={item.id}>
                        <td>{item.student}</td>
                        <td>{item.className}</td>
                        <td>
                          <span className="badge badge-outline">
                            {item.score}/{item.total}
                          </span>
                        </td>
                        <td className="text-xs text-slate-500">{item.createdAt}</td>
                      </tr>
                    ))}
                    {!attempts.length ? (
                      <tr>
                        <td colSpan={4} className="text-center text-slate-500">
                          暂无成绩记录
                        </td>
                      </tr>
                    ) : null}
                  </tbody>
                </table>
              </div>
            </article>
          </section>

          <section className="rounded-3xl border border-slate-200 bg-white p-5 shadow-card">
            <h2 className="mb-3 text-lg font-semibold text-slate-800">班级错题热区（自动统计）</h2>
            <div className="max-h-[320px] overflow-auto rounded-xl border border-slate-200">
              <table className="table table-sm">
                <thead>
                  <tr>
                    <th>班级</th>
                    <th>题目ID</th>
                    <th>题目</th>
                    <th>错误次数</th>
                  </tr>
                </thead>
                <tbody>
                  {topStats.map((item, index) => (
                    <tr key={`${item.classId}-${item.questionId}-${index}`}>
                      <td>{item.className || '-'}</td>
                      <td>{item.questionId}</td>
                      <td className="max-w-3xl truncate" title={item.question}>
                        {item.question}
                      </td>
                      <td>
                        <span className="badge badge-warning badge-outline">{item.wrongCount}</span>
                      </td>
                    </tr>
                  ))}
                  {!topStats.length ? (
                    <tr>
                      <td colSpan={4} className="text-center text-slate-500">
                        暂无错题统计数据
                      </td>
                    </tr>
                  ) : null}
                </tbody>
              </table>
            </div>
          </section>
        </main>
      )}

      <QuestionEditorModal
        open={modalOpen}
        initialData={editingQuestion}
        categories={flattenCategories(categories)}
        tags={tags}
        knowledgePoints={knowledgePoints}
        token={token}
        onClose={() => {
          setModalOpen(false);
          setEditingQuestion(null);
        }}
        onSubmit={handleSaveQuestion}
        loading={savingQuestion}
      />
    </div>
  );
}
