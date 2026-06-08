import { useEffect, useMemo, useRef, useState } from 'react';
import { toast } from 'react-hot-toast';

import {
  apiRequest,
  fetchCategories,
  fetchTags,
  fetchKnowledgePoints,
  fetchQuestions,
  fetchQuestion,
  fetchTeacherLeaderboard,
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
        <span className="text-xs text-slate-400 flex-shrink-0">
          {category.questionCount ?? 0}
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
  const getInitialFilterState = () => {
    if (typeof window === 'undefined') {
      return {
        selectedCategoryId: null,
        selectedTagIds: [],
        tagMode: 'or',
        keyword: '',
        page: 1,
        pageSize: 20,
        sortBy: 'id',
        sortOrder: 'desc',
        createdFrom: '',
        createdTo: '',
        hasAnswerError: '',
        createdBy: '',
        showAdvancedFilter: false,
      };
    }
    const params = new URLSearchParams(window.location.search);
    return {
      selectedCategoryId: params.get('categoryId') ? Number(params.get('categoryId')) : null,
      selectedTagIds: params.get('tagIds')
        ? params.get('tagIds').split(',').map(Number).filter(Boolean)
        : [],
      tagMode: params.get('tagMode') || 'or',
      keyword: params.get('keyword') || '',
      page: params.get('page') ? Number(params.get('page')) : 1,
      pageSize: params.get('pageSize') ? Number(params.get('pageSize')) : 20,
      sortBy: params.get('sortBy') || 'id',
      sortOrder: params.get('sortOrder') || 'desc',
      createdFrom: params.get('createdFrom') || '',
      createdTo: params.get('createdTo') || '',
      hasAnswerError: params.get('hasAnswerError') || '',
      createdBy: params.get('createdBy') || '',
      showAdvancedFilter: params.get('advanced') === '1',
    };
  };

  const initialState = getInitialFilterState();

  const [expandedCategoryIds, setExpandedCategoryIds] = useState([]);
  const [selectedCategoryId, setSelectedCategoryId] = useState(initialState.selectedCategoryId);
  const [selectedTagIds, setSelectedTagIds] = useState(initialState.selectedTagIds);
  const [tagMode, setTagMode] = useState(initialState.tagMode);
  const [keyword, setKeyword] = useState(initialState.keyword);
  const [page, setPage] = useState(initialState.page);
  const [pageSize, setPageSize] = useState(initialState.pageSize);
  const [questionsLoading, setQuestionsLoading] = useState(false);
  const [showAdvancedFilter, setShowAdvancedFilter] = useState(initialState.showAdvancedFilter);
  const [createdFrom, setCreatedFrom] = useState(initialState.createdFrom);
  const [createdTo, setCreatedTo] = useState(initialState.createdTo);
  const [sortBy, setSortBy] = useState(initialState.sortBy);
  const [sortOrder, setSortOrder] = useState(initialState.sortOrder);
  const [hasAnswerError, setHasAnswerError] = useState(initialState.hasAnswerError);
  const [createdBy, setCreatedBy] = useState(initialState.createdBy);
  const tableContainerRef = useRef(null);
  const scrollPositionRef = useRef(0);

  const [leaderboard, setLeaderboard] = useState(null);
  const [leaderboardLoading, setLeaderboardLoading] = useState(false);
  const [leaderboardScoreType, setLeaderboardScoreType] = useState('highest');
  const [leaderboardClassId, setLeaderboardClassId] = useState('');
  const [classes, setClasses] = useState([]);
  const [leaderboardPage, setLeaderboardPage] = useState(1);
  const [leaderboardPageSize, setLeaderboardPageSize] = useState(20);
  const [leaderboardWeightedN, setLeaderboardWeightedN] = useState(5);

  const topStats = useMemo(() => stats.slice(0, 12), [stats]);

  const syncToURL = () => {
    const params = new URLSearchParams();
    if (keyword) params.set('keyword', keyword);
    if (selectedCategoryId) params.set('categoryId', String(selectedCategoryId));
    if (selectedTagIds.length > 0) params.set('tagIds', selectedTagIds.join(','));
    if (tagMode !== 'or') params.set('tagMode', tagMode);
    if (page !== 1) params.set('page', String(page));
    if (pageSize !== 20) params.set('pageSize', String(pageSize));
    if (sortBy !== 'id') params.set('sortBy', sortBy);
    if (sortOrder !== 'desc') params.set('sortOrder', sortOrder);
    if (createdFrom) params.set('createdFrom', createdFrom);
    if (createdTo) params.set('createdTo', createdTo);
    if (hasAnswerError) params.set('hasAnswerError', hasAnswerError);
    if (createdBy) params.set('createdBy', createdBy);
    if (showAdvancedFilter) params.set('advanced', '1');

    const queryString = params.toString();
    const newURL = queryString
      ? `${window.location.pathname}?${queryString}`
      : window.location.pathname;
    window.history.replaceState(null, '', newURL);
  };

  const loadDashboard = async () => {
    setLoading(true);
    try {
      const [overviewData, statData, attemptData, categoryData, tagData, kpData, classData] = await Promise.all([
        apiRequest('/teacher/overview', { token }),
        apiRequest('/teacher/class-stats', { token }),
        apiRequest('/teacher/attempts?limit=50', { token }),
        fetchCategories(token),
        fetchTags(token),
        fetchKnowledgePoints(token),
        apiRequest('/classes', { token }),
      ]);
      setOverview(overviewData);
      setStats(statData || []);
      setAttempts(attemptData || []);
      setCategories(categoryData || []);
      setTags(tagData || []);
      setKnowledgePoints(kpData || []);
      setClasses(classData || []);
      if (categoryData && categoryData.length > 0) {
        setExpandedCategoryIds(categoryData.map((c) => c.id));
      }
      loadLeaderboard();
    } catch (error) {
      toast.error(error.message || '加载教师看板失败');
    } finally {
      setLoading(false);
    }
  };

  const loadLeaderboard = async () => {
    setLeaderboardLoading(true);
    try {
      const params = {
        scoreType: leaderboardScoreType,
        limit: leaderboardPageSize,
        page: leaderboardPage,
      };
      if (leaderboardClassId) {
        params.classId = leaderboardClassId;
      }
      if (leaderboardScoreType === 'weighted') {
        params.weightedN = leaderboardWeightedN;
      }
      const data = await fetchTeacherLeaderboard(token, params);
      setLeaderboard(data);
    } catch (error) {
      console.warn('加载排行榜失败:', error);
    } finally {
      setLeaderboardLoading(false);
    }
  };

  useEffect(() => {
    if (token && !loading) {
      loadLeaderboard();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [leaderboardScoreType, leaderboardClassId, leaderboardPage, leaderboardPageSize, leaderboardWeightedN, token]);

  const loadQuestions = async () => {
    if (tableContainerRef.current) {
      scrollPositionRef.current = tableContainerRef.current.scrollTop;
    }
    setQuestionsLoading(true);
    try {
      const params = {
        keyword: keyword || undefined,
        categoryId: selectedCategoryId || undefined,
        tagIds: selectedTagIds.length > 0 ? selectedTagIds : undefined,
        tagMode,
        page,
        pageSize,
        sortBy,
        sortOrder,
      };
      if (createdBy) {
        params.createdBy = createdBy;
      }
      if (createdFrom) {
        params.createdFrom = createdFrom;
      }
      if (createdTo) {
        params.createdTo = createdTo;
      }
      if (hasAnswerError !== '') {
        params.hasAnswerError = hasAnswerError === 'true';
      }
      const result = await fetchQuestions(token, params);
      setQuestions(result.items || []);
      setTotalQuestions(result.total || 0);
    } catch (error) {
      toast.error(error.message || '加载题目失败');
    } finally {
      setQuestionsLoading(false);
      requestAnimationFrame(() => {
        if (tableContainerRef.current) {
          tableContainerRef.current.scrollTop = scrollPositionRef.current;
        }
      });
    }
  };

  useEffect(() => {
    syncToURL();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    keyword,
    selectedCategoryId,
    selectedTagIds,
    tagMode,
    page,
    pageSize,
    sortBy,
    sortOrder,
    createdFrom,
    createdTo,
    hasAnswerError,
    createdBy,
    showAdvancedFilter,
  ]);

  useEffect(() => {
    loadDashboard();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  useEffect(() => {
    if (!loading) {
      loadQuestions();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedCategoryId, selectedTagIds, tagMode, keyword, page, pageSize, loading, sortBy, sortOrder, createdFrom, createdTo, hasAnswerError, createdBy]);

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
                    setTagMode('or');
                    setCreatedFrom('');
                    setCreatedTo('');
                    setHasAnswerError('');
                    setCreatedBy('');
                    setSortBy('id');
                    setSortOrder('desc');
                    setShowAdvancedFilter(false);
                    setPage(1);
                    setPageSize(20);
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

              <div className="mb-3">
                <button
                  type="button"
                  className="btn btn-xs btn-ghost text-sky-600"
                  onClick={() => setShowAdvancedFilter((v) => !v)}
                >
                  {showAdvancedFilter ? '收起高级筛选' : '展开高级筛选'}
                  <svg
                    className={`ml-1 h-3 w-3 transition-transform ${showAdvancedFilter ? 'rotate-180' : ''}`}
                    viewBox="0 0 12 12"
                    fill="currentColor"
                  >
                    <path d="M4 2l4 4-4 4V2z" />
                  </svg>
                </button>
              </div>

              {showAdvancedFilter && (
                <div className="mb-3 rounded-xl border border-slate-200 bg-slate-50 p-3">
                  <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-4">
                    <div>
                      <label className="text-xs text-slate-500 mb-1 block">创建时间从</label>
                      <input
                        type="date"
                        className="input input-bordered input-sm w-full"
                        value={createdFrom}
                        onChange={(e) => {
                          setCreatedFrom(e.target.value);
                          setPage(1);
                        }}
                      />
                    </div>
                    <div>
                      <label className="text-xs text-slate-500 mb-1 block">创建时间至</label>
                      <input
                        type="date"
                        className="input input-bordered input-sm w-full"
                        value={createdTo}
                        onChange={(e) => {
                          setCreatedTo(e.target.value);
                          setPage(1);
                        }}
                      />
                    </div>
                    <div>
                      <label className="text-xs text-slate-500 mb-1 block">创建者ID</label>
                      <input
                        type="text"
                        className="input input-bordered input-sm w-full"
                        placeholder="输入用户ID"
                        value={createdBy}
                        onChange={(e) => {
                          setCreatedBy(e.target.value);
                          setPage(1);
                        }}
                      />
                    </div>
                    <div>
                      <label className="text-xs text-slate-500 mb-1 block">答案异常</label>
                      <select
                        className="select select-bordered select-sm w-full"
                        value={hasAnswerError}
                        onChange={(e) => {
                          setHasAnswerError(e.target.value);
                          setPage(1);
                        }}
                      >
                        <option value="">全部</option>
                        <option value="true">仅异常</option>
                        <option value="false">仅正常</option>
                      </select>
                    </div>
                  </div>

                  <div className="mt-3 flex items-center gap-3">
                    <span className="text-xs text-slate-500">排序：</span>
                    <select
                      className="select select-bordered select-xs"
                      value={sortBy}
                      onChange={(e) => {
                        setSortBy(e.target.value);
                        setPage(1);
                      }}
                    >
                      <option value="id">题目ID</option>
                      <option value="created_at">创建时间</option>
                      <option value="wrong_count">被错次数</option>
                    </select>
                    <select
                      className="select select-bordered select-xs"
                      value={sortOrder}
                      onChange={(e) => {
                        setSortOrder(e.target.value);
                        setPage(1);
                      }}
                    >
                      <option value="desc">降序</option>
                      <option value="asc">升序</option>
                    </select>
                  </div>
                </div>
              )}

              <div ref={tableContainerRef} className="max-h-[400px] overflow-auto rounded-xl border border-slate-200 relative" id="questionTableContainer">
                {questionsLoading && (
                  <div className="absolute inset-0 z-10 flex items-center justify-center bg-white/70 backdrop-blur-sm">
                    <span className="loading loading-spinner loading-md text-sky-600"></span>
                  </div>
                )}
                <table className="table table-sm">
                  <thead>
                    <tr>
                      <th>ID</th>
                      <th>题型</th>
                      <th>分类</th>
                      <th>题干</th>
                      <th>创建者</th>
                      <th>被错次数</th>
                      <th>状态</th>
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
                        <td className="text-xs text-slate-500">
                          {question.createdByName || question.createdBy || '-'}
                        </td>
                        <td>
                          <span className="badge badge-sm badge-outline">
                            {question.wrongCount ?? 0}
                          </span>
                        </td>
                        <td>
                          {question.hasAnswerError ? (
                            <span className="badge badge-sm badge-error" title="正确答案配置异常">
                              异常
                            </span>
                          ) : (
                            <span className="badge badge-sm badge-success badge-outline">
                              正常
                            </span>
                          )}
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
                    {!questions.length && !questionsLoading ? (
                      <tr>
                        <td colSpan={8} className="text-center text-slate-500 py-8">
                          当前没有题目，请先新增或上传题库。
                        </td>
                      </tr>
                    ) : null}
                  </tbody>
                </table>
              </div>

              <div className="mt-3 flex items-center justify-between">
                <div className="text-xs text-slate-500">
                  共 {totalQuestions} 题，第 {page}/{totalPages || 1} 页
                </div>
                <div className="flex gap-1 items-center">
                  <button
                    className="btn btn-xs btn-outline"
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page <= 1 || questionsLoading}
                  >
                    上一页
                  </button>
                  <span className="text-xs text-slate-500 px-1">{page}</span>
                  <button
                    className="btn btn-xs btn-outline"
                    onClick={() => setPage((p) => Math.min(totalPages || 1, p + 1))}
                    disabled={page >= totalPages || questionsLoading}
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
                    disabled={questionsLoading}
                  >
                    <option value={10}>10/页</option>
                    <option value={20}>20/页</option>
                    <option value={50}>50/页</option>
                    <option value={100}>100/页</option>
                  </select>
                </div>
              </div>
            </article>
          </section>

          <section className="rounded-3xl border border-slate-200 bg-white p-5 shadow-card">
            <div className="mb-4 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
              <div>
                <h2 className="text-lg font-semibold text-slate-800">🏆 班级排行榜</h2>
                <p className="text-xs text-slate-500 mt-1">
                  {leaderboard?.className ? leaderboard.className + '班' : '全校'} · 共 {leaderboard?.total || 0} 名学生
                </p>
              </div>
              <div className="flex flex-wrap gap-2 items-center">
                <select
                  className="select select-bordered select-sm w-40"
                  value={leaderboardClassId}
                  onChange={(e) => {
                    setLeaderboardClassId(e.target.value);
                    setLeaderboardPage(1);
                  }}
                >
                  <option value="">全校榜</option>
                  {classes.map((cls) => (
                    <option key={cls.id} value={cls.id}>
                      {cls.name}
                    </option>
                  ))}
                </select>

                <div className="tabs tabs-boxed bg-slate-100/50 tabs-sm">
                  <button
                    className={`tab ${leaderboardScoreType === 'highest' ? 'tab-active' : ''}`}
                    onClick={() => {
                      setLeaderboardScoreType('highest');
                      setLeaderboardPage(1);
                    }}
                  >
                    最高分
                  </button>
                  <button
                    className={`tab ${leaderboardScoreType === 'average' ? 'tab-active' : ''}`}
                    onClick={() => {
                      setLeaderboardScoreType('average');
                      setLeaderboardPage(1);
                    }}
                  >
                    平均正确率
                  </button>
                  <button
                    className={`tab ${leaderboardScoreType === 'weighted' ? 'tab-active' : ''}`}
                    onClick={() => {
                      setLeaderboardScoreType('weighted');
                      setLeaderboardPage(1);
                    }}
                  >
                    加权近N次
                  </button>
                </div>

                {leaderboardScoreType === 'weighted' && (
                  <select
                    className="select select-bordered select-sm w-32"
                    value={leaderboardWeightedN}
                    onChange={(e) => {
                      setLeaderboardWeightedN(Number(e.target.value));
                      setLeaderboardPage(1);
                    }}
                  >
                    <option value={3}>近3次</option>
                    <option value={5}>近5次</option>
                    <option value={10}>近10次</option>
                    <option value={20}>近20次</option>
                  </select>
                )}
              </div>
            </div>

            <div className="relative overflow-auto rounded-xl border border-slate-200">
              {leaderboardLoading && (
                <div className="absolute inset-0 z-10 flex items-center justify-center bg-white/70 backdrop-blur-sm">
                  <span className="loading loading-spinner loading-md text-sky-600"></span>
                </div>
              )}
              <table className="table table-sm">
                <thead>
                  <tr>
                    <th className="w-16">排名</th>
                    <th>学生</th>
                    <th>班级</th>
                    <th>答题次数</th>
                    <th>正确率</th>
                    <th>综合得分</th>
                  </tr>
                </thead>
                <tbody>
                  {leaderboard?.items?.length > 0 ? (
                    leaderboard.items.map((item, idx) => {
                      const isTop3 = item.hasAttempted && item.rank <= 3;
                      const medalEmoji = item.rank === 1 ? '🥇' : item.rank === 2 ? '🥈' : item.rank === 3 ? '🥉' : null;

                      return (
                        <tr
                          key={`${item.userId}-${idx}`}
                          className={`${
                            !item.hasAttempted ? 'bg-slate-50/50 opacity-70' : isTop3 ? 'bg-amber-50/30' : ''
                          }`}
                        >
                          <td>
                            {isTop3 && medalEmoji ? (
                              <span className="text-xl">{medalEmoji}</span>
                            ) : (
                              <span className={`font-mono text-sm ${
                                item.hasAttempted ? 'text-slate-600' : 'text-slate-400'
                              }`}>{item.rank}</span>
                            )}
                          </td>
                          <td className={`font-medium ${
                            item.hasAttempted ? 'text-slate-700' : 'text-slate-500'
                          }`}>{item.username}</td>
                          <td className="text-xs text-slate-500">
                            {item.className || '-'}
                          </td>
                          <td>
                            {item.hasAttempted ? (
                              <span className="badge badge-outline badge-xs">
                                {item.attemptCount} 次
                              </span>
                            ) : (
                              <span className="text-xs text-slate-400">—</span>
                            )}
                          </td>
                          <td>
                            {item.hasAttempted ? (
                              <span className="badge badge-info badge-outline badge-xs">
                                {item.correctRate}
                              </span>
                            ) : (
                              <span className="text-xs text-slate-400">—</span>
                            )}
                          </td>
                          <td>
                            <span className={`font-bold ${
                              !item.hasAttempted
                                ? 'text-slate-400'
                                : isTop3
                                ? 'text-amber-600'
                                : 'text-slate-700'
                            }`}>
                              {item.scoreDisplay}
                            </span>
                            {item.hasAttempted && (
                              <span className="text-xs text-slate-400 ml-1">分</span>
                            )}
                          </td>
                        </tr>
                      );
                    })
                  ) : (
                    <tr>
                      <td colSpan={6} className="text-center text-slate-500 py-8">
                        暂无排名数据
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {leaderboard && leaderboard.total > 0 && (
              <div className="mt-3 flex items-center justify-between">
                <div className="text-xs text-slate-500">
                  共 {leaderboard.total} 人，第 {leaderboardPage}/{Math.ceil(leaderboard.total / leaderboardPageSize) || 1} 页
                </div>
                <div className="flex gap-1 items-center">
                  <button
                    className="btn btn-xs btn-outline"
                    onClick={() => setLeaderboardPage((p) => Math.max(1, p - 1))}
                    disabled={leaderboardPage <= 1 || leaderboardLoading}
                  >
                    上一页
                  </button>
                  <span className="text-xs text-slate-500 px-1">{leaderboardPage}</span>
                  <button
                    className="btn btn-xs btn-outline"
                    onClick={() => setLeaderboardPage((p) => p + 1)}
                    disabled={leaderboardPage * leaderboardPageSize >= leaderboard.total || leaderboardLoading}
                  >
                    下一页
                  </button>
                  <select
                    className="select select-bordered select-xs w-20 ml-2"
                    value={leaderboardPageSize}
                    onChange={(e) => {
                      setLeaderboardPageSize(Number(e.target.value));
                      setLeaderboardPage(1);
                    }}
                    disabled={leaderboardLoading}
                  >
                    <option value={10}>10/页</option>
                    <option value={20}>20/页</option>
                    <option value={50}>50/页</option>
                  </select>
                </div>
              </div>
            )}
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
        categories={categories}
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
