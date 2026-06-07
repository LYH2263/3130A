import { useEffect, useMemo, useState } from 'react';
import { toast } from 'react-hot-toast';

import { apiRequest } from '../api/client';
import { StatCard } from '../components/StatCard';

export function StudentDashboard({ user, token, onLogout }) {
  const [loading, setLoading] = useState(true);
  const [questions, setQuestions] = useState([]);
  const [answers, setAnswers] = useState({});
  const [mistakes, setMistakes] = useState([]);
  const [attempts, setAttempts] = useState([]);
  const [submitting, setSubmitting] = useState(false);
  const [loadingQuiz, setLoadingQuiz] = useState(false);
  const [lastResult, setLastResult] = useState(null);

  const className = user.classRoom?.name || '未分班';

  const averageRate = useMemo(() => {
    if (!attempts.length) {
      return '0%';
    }
    const totalScore = attempts.reduce((sum, item) => sum + item.score, 0);
    const totalCount = attempts.reduce((sum, item) => sum + item.total, 0);
    if (!totalCount) {
      return '0%';
    }
    return `${Math.round((totalScore / totalCount) * 100)}%`;
  }, [attempts]);

  const loadStudentData = async () => {
    setLoading(true);
    try {
      const [mistakeData, attemptData] = await Promise.all([
        apiRequest('/student/mistakes', { token }),
        apiRequest('/student/attempts', { token }),
      ]);
      setMistakes(mistakeData);
      setAttempts(attemptData);
    } catch (error) {
      toast.error(error.message || '加载学生数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadStudentData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  const startQuiz = async () => {
    try {
      setLoadingQuiz(true);
      const quiz = await apiRequest('/student/questions?limit=10', { token });
      setQuestions(quiz);
      setAnswers({});
      setLastResult(null);
      toast.success('已生成新试卷，选项顺序已随机');
    } catch (error) {
      toast.error(error.message || '拉取试卷失败');
    } finally {
      setLoadingQuiz(false);
    }
  };

  const submitQuiz = async () => {
    if (!questions.length) {
      toast.error('请先开始答题');
      return;
    }

    for (const question of questions) {
      if (!answers[question.id]) {
        toast.error(`请完成题目：${question.title.slice(0, 12)}...`);
        return;
      }
    }

    try {
      setSubmitting(true);
      const payload = {
        answers: questions.map((question) => ({
          questionId: question.id,
          optionId: answers[question.id],
        })),
      };
      const result = await apiRequest('/student/submit', {
        method: 'POST',
        token,
        body: payload,
      });
      setLastResult(result);
      toast.success(`提交成功：${result.score}/${result.total}`);
      await loadStudentData();
    } catch (error) {
      toast.error(error.message || '提交失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-board px-4 py-6 md:px-8 md:py-8">
      <header className="mx-auto mb-6 flex max-w-7xl flex-col gap-3 rounded-3xl border border-white/70 bg-white/90 px-6 py-5 shadow-card md:flex-row md:items-center md:justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.25em] text-emerald-700">Student Console</p>
          <h1 className="mt-1 text-2xl font-bold text-slate-800">学生答题中心</h1>
          <p className="text-sm text-slate-600">当前班级：{className}</p>
        </div>
        <div className="flex gap-2">
          <button className="btn btn-outline btn-primary" onClick={startQuiz} disabled={loadingQuiz}>
            {loadingQuiz ? '生成中...' : '开始新一轮答题'}
          </button>
          <button className="btn btn-neutral" onClick={onLogout}>
            退出登录
          </button>
        </div>
      </header>

      {loading ? (
        <div className="mx-auto max-w-7xl">
          <div className="grid gap-4 md:grid-cols-3">
            {Array.from({ length: 3 }).map((_, idx) => (
              <div key={`student-loading-${idx}`} className="h-28 animate-pulse rounded-2xl bg-white/80" />
            ))}
          </div>
        </div>
      ) : (
        <main className="mx-auto grid max-w-7xl gap-5 lg:grid-cols-[1.2fr,0.8fr]">
          <section className="space-y-5">
            <div className="grid gap-4 md:grid-cols-3">
              <StatCard title="已完成次数" value={attempts.length} />
              <StatCard title="平均正确率" value={averageRate} />
              <StatCard title="错题数量" value={mistakes.length} hint="错题本自动更新" />
            </div>

            <article className="rounded-3xl border border-slate-200 bg-white p-5 shadow-card">
              <div className="mb-3 flex items-center justify-between">
                <h2 className="text-lg font-semibold text-slate-800">在线答题</h2>
                <button className="btn btn-sm btn-secondary" onClick={submitQuiz} disabled={submitting || !questions.length}>
                  {submitting ? '提交中...' : '提交本次答案'}
                </button>
              </div>

              {!questions.length ? (
                <p className="rounded-xl border border-dashed border-slate-300 px-4 py-8 text-center text-sm text-slate-500">
                  点击“开始新一轮答题”获取题目。每次题目选项顺序会随机打乱。
                </p>
              ) : (
                <div className="space-y-4">
                  {questions.map((question, index) => (
                    <div key={question.id} className="rounded-2xl border border-slate-200 p-4">
                      <p className="text-sm font-semibold text-slate-700">
                        {index + 1}. {question.title}
                      </p>
                      {question.description ? <p className="mt-1 text-xs text-slate-500">{question.description}</p> : null}
                      <div className="mt-3 grid gap-2">
                        {question.options.map((option) => (
                          <label
                            key={option.id}
                            className="flex cursor-pointer items-center gap-2 rounded-xl border border-slate-200 px-3 py-2 transition hover:border-teal-500"
                          >
                            <input
                              type="radio"
                              className="radio radio-primary radio-sm"
                              name={`question-${question.id}`}
                              checked={answers[question.id] === option.id}
                              onChange={() =>
                                setAnswers((prev) => ({
                                  ...prev,
                                  [question.id]: option.id,
                                }))
                              }
                            />
                            <span className="text-sm text-slate-700">{option.content}</span>
                          </label>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              )}

              {lastResult ? (
                <div className="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
                  本次成绩：{lastResult.score}/{lastResult.total}（正确率 {lastResult.rate}）
                </div>
              ) : null}
            </article>
          </section>

          <section className="space-y-5">
            <article className="rounded-3xl border border-slate-200 bg-white p-5 shadow-card">
              <h2 className="mb-3 text-lg font-semibold text-slate-800">错题本（易错题）</h2>
              <div className="max-h-[260px] overflow-auto rounded-xl border border-slate-200">
                <table className="table table-sm">
                  <thead>
                    <tr>
                      <th>题目</th>
                      <th>错次</th>
                    </tr>
                  </thead>
                  <tbody>
                    {mistakes.map((item) => (
                      <tr key={item.questionId}>
                        <td className="max-w-xs truncate" title={`${item.title}\n正确答案：${item.correctOption}`}>
                          {item.title}
                        </td>
                        <td>
                          <span className="badge badge-warning badge-outline">{item.wrongCount}</span>
                        </td>
                      </tr>
                    ))}
                    {!mistakes.length ? (
                      <tr>
                        <td colSpan={2} className="text-center text-slate-500">
                          暂无错题
                        </td>
                      </tr>
                    ) : null}
                  </tbody>
                </table>
              </div>
            </article>

            <article className="rounded-3xl border border-slate-200 bg-white p-5 shadow-card">
              <h2 className="mb-3 text-lg font-semibold text-slate-800">历史成绩</h2>
              <div className="max-h-[260px] overflow-auto rounded-xl border border-slate-200">
                <table className="table table-sm">
                  <thead>
                    <tr>
                      <th>成绩</th>
                      <th>时间</th>
                    </tr>
                  </thead>
                  <tbody>
                    {attempts.map((item) => (
                      <tr key={item.id}>
                        <td>
                          <span className="badge badge-info badge-outline">{item.score}/{item.total}</span>
                        </td>
                        <td className="text-xs text-slate-500">{new Date(item.createdAt).toLocaleString()}</td>
                      </tr>
                    ))}
                    {!attempts.length ? (
                      <tr>
                        <td colSpan={2} className="text-center text-slate-500">
                          暂无记录
                        </td>
                      </tr>
                    ) : null}
                  </tbody>
                </table>
              </div>
            </article>
          </section>
        </main>
      )}
    </div>
  );
}
