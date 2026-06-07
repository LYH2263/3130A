import { useEffect, useState, useCallback } from 'react';
import { toast } from 'react-hot-toast';

import {
  fetchFavorites,
  fetchQuestionSets,
  fetchQuestionSet,
  createQuestionSet,
  updateQuestionSet,
  deleteQuestionSet,
  addQuestionsToSet,
  removeQuestionFromSet,
  reorderSetQuestions,
  toggleFavorite,
  fetchSetQuiz,
} from '../api/client';

import { QUESTION_TYPE_LABELS } from '../utils/validators';

function getTypeBadgeClass(type) {
  switch (type) {
    case 'single':
      return 'badge badge-info badge-outline';
    case 'multiple':
      return 'badge badge-warning badge-outline';
    case 'judge':
      return 'badge badge-success badge-outline';
    case 'blank':
      return 'badge badge-secondary badge-outline';
    default:
      return 'badge badge-ghost badge-outline';
  }
}

function CreateSetModal({ onClose, onCreated, token }) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error('请输入题集名称');
      return;
    }
    try {
      setLoading(true);
      const result = await createQuestionSet(token, {
        name: name.trim(),
        description: description.trim(),
        questionIds: [],
      });
      toast.success('题集创建成功');
      onCreated && onCreated(result);
      onClose && onClose();
    } catch (error) {
      toast.error(error.message || '创建失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="mx-4 w-full max-w-md rounded-2xl bg-white p-6 shadow-xl">
        <h3 className="text-lg font-semibold text-slate-800">创建题集</h3>
        <form onSubmit={handleSubmit} className="mt-4 space-y-4">
          <div>
            <label className="text-sm font-medium text-slate-700">题集名称</label>
            <input
              type="text"
              className="input input-bordered w-full mt-1"
              placeholder="请输入题集名称"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={128}
              autoFocus
            />
          </div>
          <div>
            <label className="text-sm font-medium text-slate-700">描述（可选）</label>
            <textarea
              className="textarea textarea-bordered w-full mt-1"
              placeholder="简单描述一下这个题集"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              maxLength={500}
              rows={3}
            />
          </div>
          <div className="flex gap-3 pt-2">
            <button
              type="button"
              className="btn btn-outline flex-1"
              onClick={onClose}
              disabled={loading}
            >
              取消
            </button>
            <button
              type="submit"
              className="btn btn-primary flex-1"
              disabled={loading}
            >
              {loading ? '创建中...' : '创建'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function AddToSetModal({ questionId, sets, onClose, onAdded, token }) {
  const [selectedSetId, setSelectedSetId] = useState('');
  const [newSetName, setNewSetName] = useState('');
  const [mode, setMode] = useState('select');
  const [loading, setLoading] = useState(false);

  const handleAddToSet = async () => {
    if (!selectedSetId) {
      toast.error('请选择题集');
      return;
    }
    try {
      setLoading(true);
      await addQuestionsToSet(token, selectedSetId, [questionId]);
      toast.success('已添加到题集');
      onAdded && onAdded();
      onClose && onClose();
    } catch (error) {
      toast.error(error.message || '添加失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateAndAdd = async () => {
    if (!newSetName.trim()) {
      toast.error('请输入题集名称');
      return;
    }
    try {
      setLoading(true);
      const newSet = await createQuestionSet(token, {
        name: newSetName.trim(),
        description: '',
        questionIds: [questionId],
      });
      toast.success('题集创建成功，已添加题目');
      onAdded && onAdded(newSet);
      onClose && onClose();
    } catch (error) {
      toast.error(error.message || '创建失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="mx-4 w-full max-w-md rounded-2xl bg-white p-6 shadow-xl">
        <h3 className="text-lg font-semibold text-slate-800">添加到题集</h3>
        
        <div className="tabs tabs-boxed mt-4 bg-slate-100/50">
          <button
            type="button"
            className={`tab ${mode === 'select' ? 'tab-active' : ''}`}
            onClick={() => setMode('select')}
          >
            选择题集
          </button>
          <button
            type="button"
            className={`tab ${mode === 'create' ? 'tab-active' : ''}`}
            onClick={() => setMode('create')}
          >
            新建题集
          </button>
        </div>

        {mode === 'select' ? (
          <div className="mt-4 space-y-3">
            {sets.length === 0 ? (
              <p className="text-sm text-slate-500 text-center py-4">
                暂无题集，切换到"新建题集"创建一个吧
              </p>
            ) : (
              <div className="max-h-60 overflow-y-auto space-y-2">
                {sets.map((set) => (
                  <label
                    key={set.id}
                    className={`flex cursor-pointer items-center gap-3 rounded-xl border p-3 transition ${
                      selectedSetId == set.id
                        ? 'border-teal-500 bg-teal-50'
                        : 'border-slate-200 hover:border-teal-300'
                    }`}
                  >
                    <input
                      type="radio"
                      name="set-select"
                      className="radio radio-primary radio-sm"
                      checked={selectedSetId == set.id}
                      onChange={() => setSelectedSetId(set.id)}
                    />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-slate-700 truncate">
                        {set.name}
                      </p>
                      <p className="text-xs text-slate-500">
                        {set.questionCount} 道题
                      </p>
                    </div>
                  </label>
                ))}
              </div>
            )}
            <div className="flex gap-3 pt-2">
              <button
                type="button"
                className="btn btn-outline flex-1"
                onClick={onClose}
                disabled={loading}
              >
                取消
              </button>
              <button
                type="button"
                className="btn btn-primary flex-1"
                onClick={handleAddToSet}
                disabled={loading || !selectedSetId}
              >
                {loading ? '添加中...' : '添加'}
              </button>
            </div>
          </div>
        ) : (
          <div className="mt-4 space-y-4">
            <div>
              <label className="text-sm font-medium text-slate-700">题集名称</label>
              <input
                type="text"
                className="input input-bordered w-full mt-1"
                placeholder="请输入题集名称"
                value={newSetName}
                onChange={(e) => setNewSetName(e.target.value)}
                maxLength={128}
                autoFocus
              />
            </div>
            <div className="flex gap-3 pt-2">
              <button
                type="button"
                className="btn btn-outline flex-1"
                onClick={onClose}
                disabled={loading}
              >
                取消
              </button>
              <button
                type="button"
                className="btn btn-primary flex-1"
                onClick={handleCreateAndAdd}
                disabled={loading || !newSetName.trim()}
              >
                {loading ? '创建中...' : '创建并添加'}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function SetDetailModal({ setId, onClose, onStartPractice, onUpdated, token }) {
  const [setDetail, setSetDetail] = useState(null);
  const [questions, setQuestions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [draggedIndex, setDraggedIndex] = useState(null);
  const [dragOverIndex, setDragOverIndex] = useState(null);

  const loadSetDetail = useCallback(async () => {
    try {
      setLoading(true);
      const detail = await fetchQuestionSet(token, setId);
      setSetDetail(detail);
      
      const quiz = await fetchSetQuiz(token, setId);
      setQuestions(quiz || []);
    } catch (error) {
      toast.error(error.message || '加载题集失败');
    } finally {
      setLoading(false);
    }
  }, [token, setId]);

  useEffect(() => {
    loadSetDetail();
  }, [loadSetDetail]);

  const handleDragStart = (e, index) => {
    setDraggedIndex(index);
    e.dataTransfer.effectAllowed = 'move';
  };

  const handleDragOver = (e, index) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    setDragOverIndex(index);
  };

  const handleDragLeave = () => {
    setDragOverIndex(null);
  };

  const handleDrop = async (e, dropIndex) => {
    e.preventDefault();
    if (draggedIndex === null || draggedIndex === dropIndex) {
      setDraggedIndex(null);
      setDragOverIndex(null);
      return;
    }

    const newQuestions = [...questions];
    const [draggedItem] = newQuestions.splice(draggedIndex, 1);
    newQuestions.splice(dropIndex, 0, draggedItem);
    setQuestions(newQuestions);
    setDraggedIndex(null);
    setDragOverIndex(null);

    const newOrder = newQuestions.map((q) => q.id);
    try {
      await reorderSetQuestions(token, setId, newOrder);
      onUpdated && onUpdated();
    } catch (error) {
      toast.error(error.message || '排序失败');
      loadSetDetail();
    }
  };

  const handleRemoveQuestion = async (questionId) => {
    if (!confirm('确定要从题集中移除这道题吗？')) return;
    try {
      await removeQuestionFromSet(token, setId, questionId);
      toast.success('已从题集移除');
      setQuestions((prev) => prev.filter((q) => q.id !== questionId));
      setSetDetail((prev) => prev ? { ...prev, questionCount: prev.questionCount - 1 } : null);
      onUpdated && onUpdated();
    } catch (error) {
      toast.error(error.message || '移除失败');
    }
  };

  const handleStartPractice = () => {
    if (questions.length === 0) {
      toast.error('题集为空，无法开始练习');
      return;
    }
    onStartPractice && onStartPractice(setId, setDetail?.name);
    onClose && onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="mx-4 w-full max-w-2xl max-h-[85vh] rounded-2xl bg-white shadow-xl flex flex-col">
        <div className="p-6 border-b border-slate-200">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0">
              <h3 className="text-lg font-semibold text-slate-800">
                {setDetail?.name || '题集详情'}
              </h3>
              {setDetail?.description && (
                <p className="text-sm text-slate-500 mt-1">{setDetail.description}</p>
              )}
              <p className="text-xs text-slate-400 mt-2">
                共 {setDetail?.questionCount || 0} 道题 · 拖拽可调整顺序
              </p>
            </div>
            <button
              type="button"
              className="btn btn-ghost btn-sm btn-square"
              onClick={onClose}
            >
              ✕
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          {loading ? (
            <div className="space-y-3">
              {Array.from({ length: 5 }).map((_, idx) => (
                <div
                  key={idx}
                  className="h-16 animate-pulse rounded-xl bg-slate-100"
                />
              ))}
            </div>
          ) : questions.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-slate-500">题集为空</p>
              <p className="text-xs text-slate-400 mt-1">
                从收藏列表添加题目到这个题集吧
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              {questions.map((q, index) => (
                <div
                  key={q.id}
                  draggable
                  onDragStart={(e) => handleDragStart(e, index)}
                  onDragOver={(e) => handleDragOver(e, index)}
                  onDragLeave={handleDragLeave}
                  onDrop={(e) => handleDrop(e, index)}
                  className={`flex items-center gap-3 rounded-xl border p-3 cursor-move transition ${
                    dragOverIndex === index
                      ? 'border-teal-500 bg-teal-50 scale-[1.02]'
                      : 'border-slate-200 bg-white hover:border-slate-300'
                  } ${draggedIndex === index ? 'opacity-50' : ''}`}
                >
                  <span className="text-slate-400 text-sm select-none">⋮⋮</span>
                  <span className="text-sm font-medium text-slate-500 w-6">
                    {index + 1}.
                  </span>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-slate-700 truncate">{q.title}</p>
                    <span className={`text-xs ${getTypeBadgeClass(q.type)} mt-1`}>
                      {QUESTION_TYPE_LABELS[q.type] || '单选题'}
                    </span>
                  </div>
                  <button
                    type="button"
                    className="btn btn-ghost btn-xs btn-square text-red-500 hover:text-red-600 hover:bg-red-50"
                    onClick={() => handleRemoveQuestion(q.id)}
                    title="从题集移除"
                  >
                    ✕
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="p-6 border-t border-slate-200">
          <div className="flex gap-3">
            <button
              type="button"
              className="btn btn-outline flex-1"
              onClick={onClose}
            >
              关闭
            </button>
            <button
              type="button"
              className="btn btn-primary flex-1"
              onClick={handleStartPractice}
              disabled={questions.length === 0}
            >
              开始专项练习
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export function FavoritesAndSets({ token, onStartSetPractice }) {
  const [activeSubTab, setActiveSubTab] = useState('favorites');
  const [favorites, setFavorites] = useState([]);
  const [sets, setSets] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [addToSetQuestionId, setAddToSetQuestionId] = useState(null);
  const [viewingSetId, setViewingSetId] = useState(null);

  const loadData = useCallback(async () => {
    try {
      setLoading(true);
      const [favData, setsData] = await Promise.all([
        fetchFavorites(token),
        fetchQuestionSets(token),
      ]);
      setFavorites(favData || []);
      setSets(setsData || []);
    } catch (error) {
      toast.error(error.message || '加载数据失败');
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleToggleFavorite = async (questionId) => {
    try {
      const result = await toggleFavorite(token, questionId);
      if (!result?.favorited) {
        setFavorites((prev) => prev.filter((f) => f.questionId !== questionId));
        toast.success('已取消收藏');
      }
    } catch (error) {
      toast.error(error.message || '操作失败');
    }
  };

  const handleSetCreated = () => {
    loadData();
  };

  const handleSetUpdated = () => {
    loadData();
  };

  const handleDeleteSet = async (setId, e) => {
    e.stopPropagation();
    if (!confirm('确定要删除这个题集吗？此操作不可恢复。')) return;
    try {
      await deleteQuestionSet(token, setId);
      toast.success('题集已删除');
      loadData();
    } catch (error) {
      toast.error(error.message || '删除失败');
    }
  };

  return (
    <div className="mx-auto max-w-7xl">
      <div className="grid gap-5 lg:grid-cols-[1.2fr,0.8fr]">
        <section className="space-y-5">
          <div className="tabs tabs-boxed bg-white/80">
            <button
              className={`tab ${activeSubTab === 'favorites' ? 'tab-active' : ''}`}
              onClick={() => setActiveSubTab('favorites')}
            >
              ⭐ 我的收藏
              <span className="ml-1 badge badge-sm badge-outline">
                {favorites.length}
              </span>
            </button>
            <button
              className={`tab ${activeSubTab === 'sets' ? 'tab-active' : ''}`}
              onClick={() => setActiveSubTab('sets')}
            >
              📚 我的题集
              <span className="ml-1 badge badge-sm badge-outline">
                {sets.length}
              </span>
            </button>
          </div>

          {activeSubTab === 'favorites' && (
            <article className="rounded-3xl border border-slate-200 bg-white p-5 shadow-card">
              <div className="mb-4 flex items-center justify-between">
                <h2 className="text-lg font-semibold text-slate-800">收藏的题目</h2>
                <span className="text-sm text-slate-500">
                  共 {favorites.length} 道
                </span>
              </div>

              {loading ? (
                <div className="space-y-3">
                  {Array.from({ length: 5 }).map((_, idx) => (
                    <div
                      key={idx}
                      className="h-16 animate-pulse rounded-xl bg-slate-100"
                    />
                  ))}
                </div>
              ) : favorites.length === 0 ? (
                <div className="rounded-xl border border-dashed border-slate-300 py-12 text-center">
                  <p className="text-slate-500">暂无收藏的题目</p>
                  <p className="mt-1 text-xs text-slate-400">
                    在答题时点击右上角 ⭐ 即可收藏题目
                  </p>
                </div>
              ) : (
                <div className="space-y-2 max-h-[500px] overflow-y-auto">
                  {favorites.map((item) => (
                    <div
                      key={item.id}
                      className="flex items-start gap-3 rounded-xl border border-slate-200 p-3 hover:border-slate-300 transition"
                    >
                      <button
                        type="button"
                        className="text-xl shrink-0 mt-0.5"
                        onClick={() => handleToggleFavorite(item.questionId)}
                        title="取消收藏"
                      >
                        <span className="text-amber-500">★</span>
                      </button>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm text-slate-700">{item.title}</p>
                        <div className="mt-1 flex items-center gap-2">
                          <span className={getTypeBadgeClass(item.type)}>
                            {QUESTION_TYPE_LABELS[item.type] || '单选题'}
                          </span>
                          <span className="text-xs text-slate-400">
                            收藏于 {item.createdAt}
                          </span>
                        </div>
                      </div>
                      <button
                        type="button"
                        className="btn btn-xs btn-outline btn-primary shrink-0"
                        onClick={() => setAddToSetQuestionId(item.questionId)}
                      >
                        加入题集
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </article>
          )}

          {activeSubTab === 'sets' && (
            <article className="rounded-3xl border border-slate-200 bg-white p-5 shadow-card">
              <div className="mb-4 flex items-center justify-between">
                <h2 className="text-lg font-semibold text-slate-800">我的题集</h2>
                <button
                  className="btn btn-sm btn-primary"
                  onClick={() => setShowCreateModal(true)}
                >
                  + 新建题集
                </button>
              </div>

              {loading ? (
                <div className="space-y-3">
                  {Array.from({ length: 3 }).map((_, idx) => (
                    <div
                      key={idx}
                      className="h-20 animate-pulse rounded-xl bg-slate-100"
                    />
                  ))}
                </div>
              ) : sets.length === 0 ? (
                <div className="rounded-xl border border-dashed border-slate-300 py-12 text-center">
                  <p className="text-slate-500">暂无题集</p>
                  <button
                    className="btn btn-sm btn-primary mt-3"
                    onClick={() => setShowCreateModal(true)}
                  >
                    创建第一个题集
                  </button>
                </div>
              ) : (
                <div className="space-y-3">
                  {sets.map((set) => (
                    <div
                      key={set.id}
                      className="flex items-center gap-4 rounded-xl border border-slate-200 p-4 hover:border-teal-400 hover:shadow-md cursor-pointer transition"
                      onClick={() => setViewingSetId(set.id)}
                    >
                      <div className="w-12 h-12 rounded-xl bg-teal-100 flex items-center justify-center text-2xl shrink-0">
                        📚
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="font-medium text-slate-700 truncate">
                          {set.name}
                        </p>
                        {set.description && (
                          <p className="text-xs text-slate-500 truncate mt-0.5">
                            {set.description}
                          </p>
                        )}
                        <p className="text-xs text-slate-400 mt-1">
                          {set.questionCount} 道题 · 创建于 {set.createdAt?.split(' ')[0]}
                        </p>
                      </div>
                      <div className="flex items-center gap-2 shrink-0">
                        <button
                          type="button"
                          className="btn btn-xs btn-primary"
                          onClick={(e) => {
                            e.stopPropagation();
                            onStartSetPractice && onStartSetPractice(set.id, set.name);
                          }}
                          disabled={set.questionCount === 0}
                        >
                          开始练习
                        </button>
                        <button
                          type="button"
                          className="btn btn-xs btn-ghost text-red-500 hover:text-red-600 hover:bg-red-50"
                          onClick={(e) => handleDeleteSet(set.id, e)}
                        >
                          删除
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </article>
          )}
        </section>

        <section className="space-y-5">
          <article className="rounded-3xl border border-slate-200 bg-white p-5 shadow-card">
            <h2 className="mb-3 text-lg font-semibold text-slate-800">使用说明</h2>
            <div className="space-y-3 text-sm text-slate-600">
              <div className="flex gap-2">
                <span>⭐</span>
                <span>答题时点击右上角星标可收藏/取消收藏题目</span>
              </div>
              <div className="flex gap-2">
                <span>📚</span>
                <span>可将收藏的题目加入自定义题集，方便专项练习</span>
              </div>
              <div className="flex gap-2">
                <span>🔀</span>
                <span>题集中的题目支持拖拽调整顺序</span>
              </div>
              <div className="flex gap-2">
                <span>🎯</span>
                <span>从题集开始的专项练习，题目顺序与题集一致</span>
              </div>
            </div>
          </article>
        </section>
      </div>

      {showCreateModal && (
        <CreateSetModal
          token={token}
          onClose={() => setShowCreateModal(false)}
          onCreated={handleSetCreated}
        />
      )}

      {addToSetQuestionId && (
        <AddToSetModal
          token={token}
          questionId={addToSetQuestionId}
          sets={sets}
          onClose={() => setAddToSetQuestionId(null)}
          onAdded={handleSetUpdated}
        />
      )}

      {viewingSetId && (
        <SetDetailModal
          token={token}
          setId={viewingSetId}
          onClose={() => setViewingSetId(null)}
          onStartPractice={onStartSetPractice}
          onUpdated={handleSetUpdated}
        />
      )}
    </div>
  );
}
