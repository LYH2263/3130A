import { useEffect, useMemo, useState } from 'react';
import { toast } from 'react-hot-toast';

import { apiRequest } from '../api/client';
import { QuestionEditorModal } from '../components/QuestionEditorModal';
import { StatCard } from '../components/StatCard';
import { questionSchema } from '../utils/validators';

export function TeacherDashboard({ user, token, onLogout }) {
  const [overview, setOverview] = useState(null);
  const [questions, setQuestions] = useState([]);
  const [stats, setStats] = useState([]);
  const [attempts, setAttempts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [savingQuestion, setSavingQuestion] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingQuestion, setEditingQuestion] = useState(null);
  const [uploading, setUploading] = useState(false);

  const topStats = useMemo(() => stats.slice(0, 12), [stats]);

  const loadDashboard = async () => {
    setLoading(true);
    try {
      const [overviewData, questionData, statData, attemptData] = await Promise.all([
        apiRequest('/teacher/overview', { token }),
        apiRequest('/teacher/questions', { token }),
        apiRequest('/teacher/class-stats', { token }),
        apiRequest('/teacher/attempts?limit=50', { token }),
      ]);
      setOverview(overviewData);
      setQuestions(questionData);
      setStats(statData);
      setAttempts(attemptData);
    } catch (error) {
      toast.error(error.message || '加载教师看板失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDashboard();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

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
      await loadDashboard();
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
      await loadDashboard();
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
      await loadDashboard();
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

  const openEditModal = (question) => {
    setEditingQuestion(question);
    setModalOpen(true);
  };

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

          <section className="grid gap-5 lg:grid-cols-[1.2fr,0.8fr]">
            <article className="rounded-3xl border border-slate-200 bg-white p-5 shadow-card">
              <div className="mb-3 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
                <h2 className="text-lg font-semibold text-slate-800">题库管理</h2>
                <div className="flex flex-wrap gap-2">
                  <label className="btn btn-outline btn-secondary">
                    {uploading ? '上传中...' : '上传 JSON 题库'}
                    <input type="file" className="hidden" accept="application/json" disabled={uploading} onChange={handleUpload} />
                  </label>
                  <button className="btn btn-primary" onClick={openCreateModal}>
                    新增题目
                  </button>
                </div>
              </div>

              <div className="max-h-[460px] overflow-auto rounded-xl border border-slate-200">
                <table className="table table-sm">
                  <thead>
                    <tr>
                      <th>ID</th>
                      <th>题干</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {questions.map((question) => (
                      <tr key={question.id}>
                        <td className="font-mono text-xs">{question.id}</td>
                        <td className="max-w-sm truncate" title={question.title}>
                          {question.title}
                        </td>
                        <td>
                          <div className="flex gap-1">
                            <button className="btn btn-xs btn-ghost" onClick={() => openEditModal(question)}>
                              编辑
                            </button>
                            <button className="btn btn-xs btn-ghost text-error" onClick={() => handleDeleteQuestion(question.id)}>
                              删除
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                    {!questions.length ? (
                      <tr>
                        <td colSpan={3} className="text-center text-slate-500">
                          当前没有题目，请先新增或上传题库。
                        </td>
                      </tr>
                    ) : null}
                  </tbody>
                </table>
              </div>
            </article>

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
                          <span className="badge badge-outline">{item.score}/{item.total}</span>
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
