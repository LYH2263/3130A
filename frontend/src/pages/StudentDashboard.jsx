import { useEffect, useMemo, useState } from 'react';
import { toast } from 'react-hot-toast';

import { apiRequest, fetchMistakeReviewQuiz, submitMistakeReview } from '../api/client';
import { StatCard } from '../components/StatCard';
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

function QuestionItem({ question, index, answer, onAnswer }) {
  const type = question.type || 'single';
  const typeLabel = QUESTION_TYPE_LABELS[type] || '单选题';

  const handleSingleSelect = (optionId) => {
    onAnswer(question.id, { optionId });
  };

  const handleMultipleToggle = (optionId) => {
    const currentIds = answer?.optionIds || [];
    const newIds = currentIds.includes(optionId)
      ? currentIds.filter((id) => id !== optionId)
      : [...currentIds, optionId];
    onAnswer(question.id, { optionIds: newIds });
  };

  const handleBlankChange = (value) => {
    onAnswer(question.id, { blankAnswer: value });
  };

  return (
    <div className="rounded-2xl border border-slate-200 p-4">
      <div className="mb-2 flex items-center gap-2">
        <span className="text-sm font-semibold text-slate-700">
          {index + 1}. {question.title}
        </span>
        <span className={getTypeBadgeClass(type)}>{typeLabel}</span>
      </div>
      {question.description ? (
        <p className="mb-3 text-xs text-slate-500">{question.description}</p>
      ) : null}

      {type === 'single' || type === 'judge' ? (
        <div className="mt-3 grid gap-2">
          {question.options?.map((option) => (
            <label
              key={option.id}
              className="flex cursor-pointer items-center gap-2 rounded-xl border border-slate-200 px-3 py-2 transition hover:border-teal-500"
            >
              <input
                type="radio"
                className="radio radio-primary radio-sm"
                name={`question-${question.id}`}
                checked={answer?.optionId === option.id}
                onChange={() => handleSingleSelect(option.id)}
              />
              <span className="text-sm text-slate-700">{option.content}</span>
            </label>
          ))}
        </div>
      ) : null}

      {type === 'multiple' ? (
        <div className="mt-3 grid gap-2">
          {question.options?.map((option) => (
            <label
              key={option.id}
              className="flex cursor-pointer items-center gap-2 rounded-xl border border-slate-200 px-3 py-2 transition hover:border-teal-500"
            >
              <input
                type="checkbox"
                className="checkbox checkbox-primary checkbox-sm"
                checked={(answer?.optionIds || []).includes(option.id)}
                onChange={() => handleMultipleToggle(option.id)}
              />
              <span className="text-sm text-slate-700">{option.content}</span>
            </label>
          ))}
          <p className="text-xs text-slate-400">（多选题，可选择多个答案）</p>
        </div>
      ) : null}

      {type === 'blank' ? (
        <div className="mt-3">
          <input
            type="text"
            className="input input-bordered w-full"
            placeholder="请输入答案"
            value={answer?.blankAnswer || ''}
            onChange={(e) => handleBlankChange(e.target.value)}
          />
          <p className="mt-1 text-xs text-slate-400">（填空题，请输入答案）</p>
        </div>
      ) : null}
    </div>
  );
}

function ResultDetail({ questions, details }) {
  const detailMap = useMemo(() => {
    const map = {};
    details.forEach((d) => {
      map[d.questionId] = d;
    });
    return map;
  }, [details]);

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-semibold text-slate-700">答题详情</h3>
      {questions.map((q, idx) => {
        const detail = detailMap[q.id];
        if (!detail) return null;
        const typeLabel = QUESTION_TYPE_LABELS[detail.type] || '单选题';

        return (
          <div
            key={q.id}
            className={`rounded-xl border p-3 ${
              detail.isCorrect ? 'border-emerald-300 bg-emerald-50' : 'border-red-300 bg-red-50'
            }`}
          >
            <div className="flex items-start justify-between gap-2">
              <p className="text-sm font-medium text-slate-700">
                {idx + 1}. {q.title}
              </p>
              <div className="flex flex-col items-end gap-1">
                <span
                  className={`badge badge-xs ${
                    detail.isCorrect ? 'badge-success' : 'badge-error'
                  }`}
                >
                  {detail.isCorrect ? '正确' : '错误'}
                </span>
                <span className="text-xs font-mono text-slate-500">
                  {detail.score}/{detail.maxScore}分
                </span>
              </div>
            </div>
            <div className="mt-1">
              <span className="text-xs text-slate-400">{typeLabel}</span>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function MistakeReviewResult({ result, questions }) {
  const questionMap = useMemo(() => {
    const map = {};
    questions.forEach((q) => {
      map[q.id] = q;
    });
    return map;
  }, [questions]);

  return (
    <div className="space-y-5">
      <div className="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3">
        <div className="text-lg font-bold text-emerald-700">
          本次成绩：{result.score}/{result.total}（正确率 {result.rate}）
        </div>
        <div className="mt-1 text-sm text-emerald-600">
          新掌握 {result.newlyMastered?.length || 0} 题，仍需巩固 {result.stillNeedReview?.length || 0} 题
        </div>
      </div>

      {result.newlyMastered && result.newlyMastered.length > 0 && (
        <div className="rounded-xl border border-amber-200 bg-amber-50 p-4">
          <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold text-amber-800">
            <span className="badge badge-success badge-sm">新掌握</span>
            恭喜！本次新掌握的题目
          </h3>
          <div className="space-y-2">
            {result.newlyMastered.map((item, idx) => {
              const q = questionMap[item.questionId];
              return (
                <div
                  key={item.questionId}
                  className="rounded-lg border border-amber-200 bg-white p-3"
                >
                  <div className="flex items-start justify-between gap-2">
                    <p className="text-sm font-medium text-slate-700">
                      {idx + 1}. {q?.title || '未知题目'}
                    </p>
                    <span className="badge badge-success badge-xs">已掌握</span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {result.stillNeedReview && result.stillNeedReview.length > 0 && (
        <div className="rounded-xl border border-sky-200 bg-sky-50 p-4">
          <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold text-sky-800">
            <span className="badge badge-info badge-sm">仍需巩固</span>
            还需要继续复习的题目
          </h3>
          <div className="space-y-2">
            {result.stillNeedReview.map((item, idx) => {
              const q = questionMap[item.questionId];
              return (
                <div
                  key={item.questionId}
                  className={`rounded-lg border bg-white p-3 ${
                    item.isCorrect ? 'border-emerald-200' : 'border-red-200'
                  }`}
                >
                  <div className="flex items-start justify-between gap-2">
                    <p className="text-sm font-medium text-slate-700">
                      {idx + 1}. {q?.title || '未知题目'}
                    </p>
                    <div className="flex flex-col items-end gap-1">
                      <span
                        className={`badge badge-xs ${
                          item.isCorrect ? 'badge-success' : 'badge-error'
                        }`}
                      >
                        {item.isCorrect ? '答对' : '答错'}
                      </span>
                      <span className="text-xs text-slate-500">
                        进度：{item.reviewCount}次
                      </span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {result.details && result.details.length > 0 && (
        <ResultDetail questions={questions} details={result.details} />
      )}
    </div>
  );
}

function isAnswerProvided(question, answer) {
  const type = question.type || 'single';
  if (type === 'single' || type === 'judge') {
    return !!answer?.optionId;
  }
  if (type === 'multiple') {
    return (answer?.optionIds || []).length > 0;
  }
  if (type === 'blank') {
    return (answer?.blankAnswer || '').trim().length > 0;
  }
  return false;
}

function MistakeItem({ item, onReview }) {
  const isMastered = item.status === 'mastered';

  return (
    <div className={`rounded-xl border p-3 ${isMastered ? 'bg-emerald-50/50 border-emerald-200' : 'bg-white border-slate-200'}`}>
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <p
            className="truncate text-sm font-medium text-slate-700"
            title={`${item.title}\n正确答案：${item.correctOption}`}
          >
            {item.title}
          </p>
          <div className="mt-1 flex items-center gap-2">
            <span className={getTypeBadgeClass(item.type)}>
              {QUESTION_TYPE_LABELS[item.type] || '单选'}
            </span>
            <span className="badge badge-warning badge-outline badge-xs">
              错{item.wrongCount}次
            </span>
            {isMastered ? (
              <span className="badge badge-success badge-xs">已掌握</span>
            ) : (
              <span className="badge badge-ghost badge-xs">
                复习{item.reviewCount}次
              </span>
            )}
          </div>
        </div>
        {!isMastered && (
          <button
            className="btn btn-xs btn-primary btn-outline shrink-0"
            onClick={() => onReview && onReview(item.questionId)}
          >
            重练
          </button>
        )}
      </div>
      <div className="mt-2">
        <div className="flex items-center justify-between text-xs text-slate-500 mb-1">
          <span>掌握进度</span>
          <span>{item.masteryRate || 0}%</span>
        </div>
        <div className="h-1.5 w-full overflow-hidden rounded-full bg-slate-200">
          <div
            className={`h-full rounded-full transition-all ${
              isMastered ? 'bg-emerald-500' : 'bg-teal-500'
            }`}
            style={{ width: `${item.masteryRate || 0}%` }}
          />
        </div>
      </div>
    </div>
  );
}

export function StudentDashboard({ user, token, onLogout }) {
  const [loading, setLoading] = useState(true);
  const [questions, setQuestions] = useState([]);
  const [answers, setAnswers] = useState({});
  const [mistakes, setMistakes] = useState([]);
  const [attempts, setAttempts] = useState([]);
  const [submitting, setSubmitting] = useState(false);
  const [loadingQuiz, setLoadingQuiz] = useState(false);
  const [lastResult, setLastResult] = useState(null);
  const [quizMode, setQuizMode] = useState('normal');
  const [mistakeReviewResult, setMistakeReviewResult] = useState(null);

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

  const pendingMistakes = useMemo(() => {
    return mistakes.filter((m) => m.status !== 'mastered');
  }, [mistakes]);

  const masteredMistakes = useMemo(() => {
    return mistakes.filter((m) => m.status === 'mastered');
  }, [mistakes]);

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
      setMistakeReviewResult(null);
      setQuizMode('normal');
      toast.success('已生成新试卷，选项顺序已随机');
    } catch (error) {
      toast.error(error.message || '拉取试卷失败');
    } finally {
      setLoadingQuiz(false);
    }
  };

  const startMistakeReview = async () => {
    if (pendingMistakes.length === 0) {
      toast.error('没有待复习的错题');
      return;
    }
    try {
      setLoadingQuiz(true);
      const quiz = await fetchMistakeReviewQuiz(token, 10);
      if (quiz.length === 0) {
        toast.error('没有待复习的错题');
        return;
      }
      setQuestions(quiz);
      setAnswers({});
      setLastResult(null);
      setMistakeReviewResult(null);
      setQuizMode('review');
      toast.success(`已生成错题重练卷，共${quiz.length}道题`);
    } catch (error) {
      toast.error(error.message || '生成错题重练卷失败');
    } finally {
      setLoadingQuiz(false);
    }
  };

  const handleSingleMistakeReview = async (questionId) => {
    try {
      setLoadingQuiz(true);
      const quiz = await fetchMistakeReviewQuiz(token, 1);
      const target = quiz.find((q) => q.id === questionId);
      const questionsToUse = target ? [target] : quiz.slice(0, 1);
      if (questionsToUse.length === 0) {
        toast.error('无法加载该题目');
        return;
      }
      setQuestions(questionsToUse);
      setAnswers({});
      setLastResult(null);
      setMistakeReviewResult(null);
      setQuizMode('review');
      toast.success('开始单题重练');
    } catch (error) {
      toast.error(error.message || '加载题目失败');
    } finally {
      setLoadingQuiz(false);
    }
  };

  const handleAnswer = (questionId, answerData) => {
    setAnswers((prev) => ({
      ...prev,
      [questionId]: { ...prev[questionId], ...answerData },
    }));
  };

  const buildAnswersPayload = () => {
    return questions.map((question) => {
      const answer = answers[question.id] || {};
      const type = question.type || 'single';

      if (type === 'single' || type === 'judge') {
        return {
          questionId: question.id,
          optionId: answer.optionId,
        };
      }
      if (type === 'multiple') {
        return {
          questionId: question.id,
          optionIds: answer.optionIds || [],
        };
      }
      if (type === 'blank') {
        return {
          questionId: question.id,
          blankAnswer: answer.blankAnswer || '',
        };
      }
      return { questionId: question.id };
    });
  };

  const validateAnswers = () => {
    for (const question of questions) {
      if (!isAnswerProvided(question, answers[question.id])) {
        toast.error(`请完成第 ${questions.indexOf(question) + 1} 题：${question.title.slice(0, 12)}...`);
        return false;
      }
    }
    return true;
  };

  const submitQuiz = async () => {
    if (!questions.length) {
      toast.error('请先开始答题');
      return;
    }

    if (!validateAnswers()) {
      return;
    }

    try {
      setSubmitting(true);
      const answersPayload = buildAnswersPayload();

      if (quizMode === 'review') {
        const result = await submitMistakeReview(token, answersPayload);
        setMistakeReviewResult(result);
        toast.success(`提交成功：${result.score}/${result.total}`);
        await loadStudentData();
      } else {
        const payload = { answers: answersPayload };
        const result = await apiRequest('/student/submit', {
          method: 'POST',
          token,
          body: payload,
        });
        setLastResult(result);
        toast.success(`提交成功：${result.score}/${result.total}`);
        await loadStudentData();
      }
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
        <div className="flex flex-wrap gap-2">
          <button
            className="btn btn-outline btn-primary"
            onClick={startQuiz}
            disabled={loadingQuiz}
          >
            {loadingQuiz ? '生成中...' : '开始新一轮答题'}
          </button>
          <button
            className="btn btn-secondary"
            onClick={startMistakeReview}
            disabled={loadingQuiz || pendingMistakes.length === 0}
          >
            错题重练
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
              <div
                key={`student-loading-${idx}`}
                className="h-28 animate-pulse rounded-2xl bg-white/80"
              />
            ))}
          </div>
        </div>
      ) : (
        <main className="mx-auto grid max-w-7xl gap-5 lg:grid-cols-[1.2fr,0.8fr]">
          <section className="space-y-5">
            <div className="grid gap-4 md:grid-cols-3">
              <StatCard title="已完成次数" value={attempts.length} />
              <StatCard title="平均正确率" value={averageRate} />
              <StatCard
                title="错题数量"
                value={pendingMistakes.length}
                hint={`已掌握 ${masteredMistakes.length} 题`}
              />
            </div>

            <article className="rounded-3xl border border-slate-200 bg-white p-5 shadow-card">
              <div className="mb-3 flex items-center justify-between">
                <h2 className="text-lg font-semibold text-slate-800">
                  {quizMode === 'review' ? '错题重练' : '在线答题'}
                </h2>
                <button
                  className="btn btn-sm btn-secondary"
                  onClick={submitQuiz}
                  disabled={submitting || !questions.length}
                >
                  {submitting ? '提交中...' : '提交本次答案'}
                </button>
              </div>

              {!questions.length ? (
                <div className="space-y-3">
                  <p className="rounded-xl border border-dashed border-slate-300 px-4 py-8 text-center text-sm text-slate-500">
                    点击“开始新一轮答题”获取题目。每次题目选项顺序会随机打乱。
                  </p>
                  <p className="rounded-xl border border-dashed border-teal-300 bg-teal-50/50 px-4 py-4 text-center text-sm text-teal-600">
                    或者点击“错题重练”专项攻克易错题。连续答对2次自动标记为已掌握。
                  </p>
                </div>
              ) : (
                <div className="space-y-4">
                  {questions.map((question, index) => (
                    <QuestionItem
                      key={question.id}
                      question={question}
                      index={index}
                      answer={answers[question.id]}
                      onAnswer={handleAnswer}
                    />
                  ))}
                </div>
              )}

              {lastResult && quizMode === 'normal' ? (
                <div className="mt-4 space-y-4">
                  <div className="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
                    本次成绩：{lastResult.score}/{lastResult.total}（正确率 {lastResult.rate}）
                  </div>
                  {lastResult.details && lastResult.details.length > 0 && (
                    <ResultDetail questions={questions} details={lastResult.details} />
                  )}
                </div>
              ) : null}

              {mistakeReviewResult && quizMode === 'review' ? (
                <div className="mt-4">
                  <MistakeReviewResult result={mistakeReviewResult} questions={questions} />
                </div>
              ) : null}
            </article>
          </section>

          <section className="space-y-5">
            <article className="rounded-3xl border border-slate-200 bg-white p-5 shadow-card">
              <div className="mb-3 flex items-center justify-between">
                <h2 className="text-lg font-semibold text-slate-800">错题本</h2>
                <span className="text-xs text-slate-500">
                  待复习 {pendingMistakes.length} / 已掌握 {masteredMistakes.length}
                </span>
              </div>
              <div className="max-h-[360px] space-y-2 overflow-auto rounded-xl">
                {mistakes.map((item) => (
                  <MistakeItem
                    key={item.questionId}
                    item={item}
                    onReview={handleSingleMistakeReview}
                  />
                ))}
                {!mistakes.length ? (
                  <div className="rounded-xl border border-dashed border-slate-300 py-8 text-center text-sm text-slate-500">
                    暂无错题，继续保持！
                  </div>
                ) : null}
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
                          <span className="badge badge-info badge-outline">
                            {item.score}/{item.total}
                          </span>
                        </td>
                        <td className="text-xs text-slate-500">
                          {new Date(item.createdAt).toLocaleString()}
                        </td>
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
