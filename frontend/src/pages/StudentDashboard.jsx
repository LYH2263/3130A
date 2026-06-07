import { useEffect, useMemo, useRef, useState } from 'react';
import { toast } from 'react-hot-toast';

import { apiRequest, fetchMistakeReviewQuiz, submitMistakeReview, saveDraft, getDraft, clearDraft, fetchExplanations, toggleFavorite, fetchFavoriteStatus, fetchSetQuiz, fetchAttemptDetail, fetchExamConfigs, startQuiz as startQuizApi, fetchStudentLeaderboard } from '../api/client';
import { CountdownTimer } from '../components/CountdownTimer';
import { FavoritesAndSets } from './FavoritesAndSets';
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

function QuestionItem({ question, index, answer, onAnswer, favorited, onToggleFavorite }) {
  const type = question.type || 'single';
  const typeLabel = QUESTION_TYPE_LABELS[type] || '单选题';
  const [favLoading, setFavLoading] = useState(false);

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

  const handleToggleFavorite = async () => {
    if (favLoading || !onToggleFavorite) return;
    setFavLoading(true);
    try {
      await onToggleFavorite(question.id);
    } finally {
      setFavLoading(false);
    }
  };

  return (
    <div className="relative rounded-2xl border border-slate-200 p-4">
      <button
        type="button"
        className={`absolute right-3 top-3 text-xl transition-transform hover:scale-110 ${favLoading ? 'opacity-50' : ''}`}
        onClick={handleToggleFavorite}
        disabled={favLoading}
        title={favorited ? '取消收藏' : '收藏题目'}
      >
        {favorited ? (
          <span className="text-amber-500">★</span>
        ) : (
          <span className="text-slate-300 hover:text-amber-400">☆</span>
        )}
      </button>
      <div className="mb-2 flex items-center gap-2 pr-8">
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

function ExplanationPanel({ explanation, knowledgePoints, compact = false }) {
  const hasContent = explanation?.content || (knowledgePoints && knowledgePoints.length > 0);
  const refs = explanation?.references
    ? explanation.references.split('\n').filter((r) => r.trim())
    : [];

  if (!hasContent) {
    return null;
  }

  return (
    <div className={`rounded-lg bg-slate-50 border border-slate-200 ${compact ? 'p-3' : 'p-4'}`}>
      {knowledgePoints && knowledgePoints.length > 0 && (
        <div className="mb-2">
          <span className="text-xs font-medium text-slate-500">知识点：</span>
          <div className="mt-1 flex flex-wrap gap-1.5">
            {knowledgePoints.map((kp) => (
              <span key={kp.id || kp.name} className="badge badge-secondary badge-xs">
                {kp.name}
              </span>
            ))}
          </div>
        </div>
      )}
      {explanation?.content && (
        <div className="mt-2">
          <span className="text-xs font-medium text-slate-500">解析：</span>
          <div
            className="mt-1 text-sm text-slate-700 leading-relaxed"
            dangerouslySetInnerHTML={{ __html: explanation.content }}
          />
        </div>
      )}
      {refs.length > 0 && (
        <div className="mt-3">
          <span className="text-xs font-medium text-slate-500">参考链接：</span>
          <div className="mt-1 space-y-1">
            {refs.map((ref, idx) => (
              <a
                key={idx}
                href={ref}
                target="_blank"
                rel="noopener noreferrer"
                className="block text-xs text-sky-600 hover:text-sky-700 hover:underline truncate"
              >
                {ref}
              </a>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function ResultDetail({ questions, details, explanations, knowledgePointsMap }) {
  const [expandedIds, setExpandedIds] = useState([]);
  const detailMap = useMemo(() => {
    const map = {};
    details.forEach((d) => {
      map[d.questionId] = d;
    });
    return map;
  }, [details]);

  const toggleExpand = (questionId) => {
    setExpandedIds((prev) =>
      prev.includes(questionId)
        ? prev.filter((id) => id !== questionId)
        : [...prev, questionId]
    );
  };

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-semibold text-slate-700">答题详情</h3>
      {questions.map((q, idx) => {
        const detail = detailMap[q.id];
        if (!detail) return null;
        const typeLabel = QUESTION_TYPE_LABELS[detail.type] || '单选题';
        const isExpanded = expandedIds.includes(q.id);
        const explanation = explanations?.find((e) => e.questionId === q.id);
        const kps = knowledgePointsMap?.[q.id] || [];

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

            <button
              type="button"
              className="mt-2 text-xs text-sky-600 hover:text-sky-700 font-medium"
              onClick={() => toggleExpand(q.id)}
            >
              {isExpanded ? '收起解析 ▲' : '查看解析与知识点 ▼'}
            </button>

            {isExpanded && (
              <div className="mt-2">
                <ExplanationPanel
                  explanation={explanation}
                  knowledgePoints={kps}
                  compact
                />
              </div>
            )}
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
  const hasExplanation = item.explanationContent || item.explanationRefs || (item.knowledgePoints && item.knowledgePoints.length > 0);

  const explanation = hasExplanation
    ? {
        content: item.explanationContent,
        references: item.explanationRefs,
      }
    : null;

  return (
    <div className={`rounded-xl border p-3 ${isMastered ? 'bg-emerald-50/50 border-emerald-200' : 'bg-white border-slate-200'}`}>
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <p
            className="text-sm font-medium text-slate-700"
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
      {hasExplanation && (
        <div className="mt-3 pt-3 border-t border-slate-200">
          <p className="text-xs font-medium text-slate-500 mb-2">解析</p>
          <ExplanationPanel
            explanation={explanation}
            knowledgePoints={item.knowledgePoints}
            compact
          />
        </div>
      )}
    </div>
  );
}

function ReportQuestionItem({ answer, index }) {
  const typeLabel = QUESTION_TYPE_LABELS[answer.questionType] || '单选题';
  const isBlank = answer.questionType === 'blank';

  const isOptionSelected = (optId) => {
    if (answer.questionType === 'single' || answer.questionType === 'judge') {
      return answer.selectedOptionId === optId;
    }
    if (answer.questionType === 'multiple') {
      return (answer.selectedOptionIds || []).includes(optId);
    }
    return false;
  };

  const getOptionClass = (opt) => {
    const selected = isOptionSelected(opt.id);
    if (opt.isCorrect) {
      return selected
        ? 'border-emerald-400 bg-emerald-50 text-emerald-800'
        : 'border-emerald-300 bg-emerald-50/50 text-emerald-700';
    }
    if (selected && !opt.isCorrect) {
      return 'border-red-400 bg-red-50 text-red-700';
    }
    return 'border-slate-200 bg-white text-slate-600';
  };

  return (
    <div className={`rounded-xl border p-4 ${
      answer.isCorrect ? 'border-emerald-200 bg-emerald-50/30' : 'border-red-200 bg-red-50/30'
    }`}>
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-slate-700">
              {index + 1}. {answer.questionTitle}
            </span>
          </div>
          <div className="mt-1 flex items-center gap-2">
            <span className={getTypeBadgeClass(answer.questionType)}>{typeLabel}</span>
            <span className={`badge badge-xs ${answer.isCorrect ? 'badge-success' : 'badge-error'}`}>
              {answer.isCorrect ? '答对' : '答错'}
            </span>
          </div>
        </div>
        <span className="text-xs font-mono text-slate-500 shrink-0">
          {answer.score}/{answer.maxScore}分
        </span>
      </div>

      {!isBlank && answer.options && answer.options.length > 0 && (
        <div className="mt-3 grid gap-2">
          {answer.options.map((opt) => (
            <div
              key={opt.id}
              className={`flex items-center gap-2 rounded-lg border px-3 py-2 text-sm ${getOptionClass(opt)}`}
            >
              {answer.questionType === 'multiple' ? (
                <input
                  type="checkbox"
                  className="checkbox checkbox-xs"
                  checked={isOptionSelected(opt.id)}
                  readOnly
                />
              ) : (
                <input
                  type="radio"
                  className="radio radio-xs"
                  checked={isOptionSelected(opt.id)}
                  readOnly
                />
              )}
              <span className="flex-1">{opt.content}</span>
              {opt.isCorrect && (
                <span className="text-xs font-semibold text-emerald-600">✓ 正确答案</span>
              )}
              {isOptionSelected(opt.id) && !opt.isCorrect && (
                <span className="text-xs font-semibold text-red-600">✗ 你的选择</span>
              )}
            </div>
          ))}
        </div>
      )}

      {isBlank && (
        <div className="mt-3 space-y-2">
          <div className="rounded-lg border border-red-300 bg-red-50 px-3 py-2">
            <span className="text-xs font-medium text-red-600">你的答案：</span>
            <span className="text-sm text-red-700 ml-1">
              {answer.blankAnswer || '（未作答）'}
            </span>
          </div>
          <div className="rounded-lg border border-emerald-300 bg-emerald-50 px-3 py-2">
            <span className="text-xs font-medium text-emerald-600">正确答案：</span>
            <span className="text-sm text-emerald-700 ml-1">
              {(answer.correctBlankAnswers || []).join(' / ')}
            </span>
          </div>
        </div>
      )}
    </div>
  );
}

function AttemptReportModal({ report, loading, onClose, onReviewWrong }) {
  if (!report && !loading) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="flex h-[85vh] w-full max-w-3xl flex-col rounded-2xl bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-slate-200 px-6 py-4">
          <div>
            <h3 className="text-lg font-semibold text-slate-800">答题报告</h3>
            {report?.createdAt && (
              <p className="text-xs text-slate-500 mt-0.5">答题时间：{report.createdAt}</p>
            )}
          </div>
          <button
            className="btn btn-sm btn-ghost btn-circle"
            onClick={onClose}
            disabled={loading}
          >
            ✕
          </button>
        </div>

        {loading ? (
          <div className="flex flex-1 items-center justify-center">
            <div className="text-center">
              <div className="loading loading-spinner loading-lg text-primary"></div>
              <p className="mt-3 text-sm text-slate-500">加载报告中...</p>
            </div>
          </div>
        ) : report ? (
          <>
            <div className="grid grid-cols-4 gap-3 border-b border-slate-200 px-6 py-4">
              <div className="rounded-xl bg-sky-50 p-3 text-center">
                <p className="text-xs text-sky-600">总分</p>
                <p className="text-xl font-bold text-sky-700">
                  {report.score}/{report.total}
                </p>
              </div>
              <div className="rounded-xl bg-emerald-50 p-3 text-center">
                <p className="text-xs text-emerald-600">正确率</p>
                <p className="text-xl font-bold text-emerald-700">{report.rate}</p>
              </div>
              <div className="rounded-xl bg-emerald-50 p-3 text-center">
                <p className="text-xs text-emerald-600">答对</p>
                <p className="text-xl font-bold text-emerald-700">{report.correctCount}</p>
              </div>
              <div className="rounded-xl bg-red-50 p-3 text-center">
                <p className="text-xs text-red-600">答错</p>
                <p className="text-xl font-bold text-red-700">{report.wrongCount}</p>
              </div>
            </div>

            {report.wrongCount > 0 && (
              <div className="px-6 py-3 border-b border-slate-200">
                <button
                  className="btn btn-sm btn-secondary w-full"
                  onClick={onReviewWrong}
                >
                  📝 重练本次错题
                </button>
              </div>
            )}

            <div className="flex-1 overflow-auto px-6 py-4">
              <div className="space-y-3">
                <h4 className="text-sm font-semibold text-slate-700">
                  答题详情（共 {report.questionCount} 题）
                </h4>
                {report.answers.map((ans, idx) => (
                  <ReportQuestionItem key={ans.questionId} answer={ans} index={idx} />
                ))}
              </div>
            </div>

            <div className="border-t border-slate-200 px-6 py-3">
              <button className="btn btn-primary w-full" onClick={onClose}>
                关闭
              </button>
            </div>
          </>
        ) : null}
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
  const [normalQuizExplanations, setNormalQuizExplanations] = useState([]);
  const [normalQuizKPMap, setNormalQuizKPMap] = useState({});
  const [quizMode, setQuizMode] = useState('normal');
  const [mistakeReviewResult, setMistakeReviewResult] = useState(null);
  const [saveDraftStatus, setSaveDraftStatus] = useState('idle');
  const [showDraftDialog, setShowDraftDialog] = useState(false);
  const [draftData, setDraftData] = useState(null);
  const [pendingStartMode, setPendingStartMode] = useState(null);
  const [favoriteStatus, setFavoriteStatus] = useState({});
  const [activeTab, setActiveTab] = useState('quiz');
  const [showReportModal, setShowReportModal] = useState(false);
  const [reportData, setReportData] = useState(null);
  const [reportLoading, setReportLoading] = useState(false);
  const [examConfigs, setExamConfigs] = useState([]);
  const [selectedExamConfigId, setSelectedExamConfigId] = useState(null);
  const [deadline, setDeadline] = useState(null);
  const [startedAt, setStartedAt] = useState(null);
  const [allowEarlySubmit, setAllowEarlySubmit] = useState(true);
  const [isAutoSubmitting, setIsAutoSubmitting] = useState(false);
  const [leaderboard, setLeaderboard] = useState(null);
  const [leaderboardLoading, setLeaderboardLoading] = useState(false);
  const [leaderboardScoreType, setLeaderboardScoreType] = useState('highest');
  const saveDraftTimerRef = useRef(null);
  const saveDraftStatusTimerRef = useRef(null);
  const hasAutoSubmittedRef = useRef(false);

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
      const [mistakeData, attemptData, examConfigData] = await Promise.all([
        apiRequest('/student/mistakes', { token }),
        apiRequest('/student/attempts', { token }),
        fetchExamConfigs(token),
      ]);
      setMistakes(mistakeData);
      setAttempts(attemptData);
      setExamConfigs(examConfigData || []);
      const defaultConfig = (examConfigData || []).find((c) => c.isDefault);
      if (defaultConfig && !selectedExamConfigId) {
        setSelectedExamConfigId(defaultConfig.id);
      }
      loadLeaderboard();
    } catch (error) {
      toast.error(error.message || '加载学生数据失败');
    } finally {
      setLoading(false);
    }
  };

  const loadLeaderboard = async () => {
    setLeaderboardLoading(true);
    try {
      const data = await fetchStudentLeaderboard(token, {
        scoreType: leaderboardScoreType,
        limit: 20,
      });
      setLeaderboard(data);
    } catch (error) {
      console.warn('加载排行榜失败:', error);
    } finally {
      setLeaderboardLoading(false);
    }
  };

  useEffect(() => {
    if (token) {
      loadLeaderboard();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [leaderboardScoreType, token]);

  useEffect(() => {
    loadStudentData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  useEffect(() => {
    return () => {
      if (saveDraftTimerRef.current) {
        clearTimeout(saveDraftTimerRef.current);
      }
      if (saveDraftStatusTimerRef.current) {
        clearTimeout(saveDraftStatusTimerRef.current);
      }
    };
  }, []);

  const startQuiz = async () => {
    try {
      setLoadingQuiz(true);
      const draft = await getDraft(token, 'normal').catch(() => null);
      if (draft && draft.questions && draft.questions.length > 0) {
        setDraftData(draft);
        setPendingStartMode('normal');
        setShowDraftDialog(true);
        setLoadingQuiz(false);
        return;
      }
      await startFreshQuiz('normal');
    } catch (error) {
      toast.error(error.message || '拉取试卷失败');
      setLoadingQuiz(false);
    }
  };

  const loadFavoriteStatus = async (questionIds) => {
    if (!questionIds || questionIds.length === 0) return;
    try {
      const status = await fetchFavoriteStatus(token, questionIds);
      setFavoriteStatus(status || {});
    } catch (error) {
      console.warn('加载收藏状态失败:', error);
    }
  };

  const handleToggleFavorite = async (questionId) => {
    try {
      const result = await toggleFavorite(token, questionId);
      setFavoriteStatus((prev) => ({
        ...prev,
        [questionId]: result?.favorited ?? !prev[questionId],
      }));
      toast.success(result?.favorited ? '已收藏' : '已取消收藏');
    } catch (error) {
      toast.error(error.message || '操作失败');
    }
  };

  const startFreshQuiz = async (mode) => {
    try {
      setLoadingQuiz(true);
      hasAutoSubmittedRef.current = false;
      let quiz;
      let quizDeadline = null;
      let quizStartedAt = null;
      let quizAllowEarlySubmit = true;

      if (mode === 'review') {
        quiz = await fetchMistakeReviewQuiz(token, 10);
        if (quiz.length === 0) {
          toast.error('没有待复习的错题');
          setLoadingQuiz(false);
          return;
        }
        toast.success(`已生成错题重练卷，共${quiz.length}道题`);
      } else if (mode === 'set') {
        quiz = await fetchSetQuiz(token, mode);
        toast.success(`已生成题集练习卷，共${quiz.length}道题`);
      } else {
        const startData = {
          mode: 'normal',
          limit: 10,
          examConfigId: selectedExamConfigId || null,
        };
        const result = await startQuizApi(token, startData);
        quiz = result.questions;
        quizDeadline = result.deadline;
        quizStartedAt = result.startedAt;
        quizAllowEarlySubmit = result.allowEarlySubmit;
        toast.success(`已生成新试卷，共${quiz.length}道题，${result.durationMinutes}分钟限时`);
      }

      setQuestions(quiz);
      setAnswers({});
      setLastResult(null);
      setNormalQuizExplanations([]);
      setNormalQuizKPMap({});
      setMistakeReviewResult(null);
      setQuizMode(mode);
      setDeadline(quizDeadline);
      setStartedAt(quizStartedAt);
      setAllowEarlySubmit(quizAllowEarlySubmit);
      loadFavoriteStatus(quiz.map((q) => q.id));
    } catch (error) {
      toast.error(error.message || '拉取试卷失败');
    } finally {
      setLoadingQuiz(false);
    }
  };

  const startSetPractice = async (setId, setName) => {
    try {
      setLoadingQuiz(true);
      const quiz = await fetchSetQuiz(token, setId);
      if (quiz.length === 0) {
        toast.error('题集为空，无法开始练习');
        setLoadingQuiz(false);
        return;
      }
      setQuestions(quiz);
      setAnswers({});
      setLastResult(null);
      setNormalQuizExplanations([]);
      setNormalQuizKPMap({});
      setMistakeReviewResult(null);
      setQuizMode(`set-${setId}`);
      setActiveTab('quiz');
      loadFavoriteStatus(quiz.map((q) => q.id));
      toast.success(`已生成「${setName}」专项练习，共${quiz.length}道题`);
    } catch (error) {
      toast.error(error.message || '拉取题集失败');
    } finally {
      setLoadingQuiz(false);
    }
  };

  const handleViewReport = async (attemptId) => {
    try {
      setReportLoading(true);
      setShowReportModal(true);
      const report = await fetchAttemptDetail(token, attemptId);
      setReportData(report);
    } catch (error) {
      toast.error(error.message || '加载报告失败');
      setShowReportModal(false);
    } finally {
      setReportLoading(false);
    }
  };

  const handleCloseReport = () => {
    setShowReportModal(false);
    setReportData(null);
  };

  const handleReviewWrongFromReport = async () => {
    if (!reportData || reportData.wrongCount === 0) {
      toast.error('本次没有错题');
      return;
    }
    try {
      setReportLoading(true);
      const quiz = await fetchMistakeReviewQuiz(token, 20);
      const wrongIds = reportData.answers.filter(a => !a.isCorrect).map(a => a.questionId);
      const filtered = quiz.filter(q => wrongIds.includes(q.id));
      const questionsToUse = filtered.length > 0 ? filtered : quiz.slice(0, wrongIds.length);
      if (questionsToUse.length === 0) {
        toast.error('没有可重练的错题');
        return;
      }
      setQuestions(questionsToUse);
      setAnswers({});
      setLastResult(null);
      setMistakeReviewResult(null);
      setQuizMode('review');
      setActiveTab('quiz');
      setShowReportModal(false);
      setReportData(null);
      toast.success(`已生成错题重练卷，共${questionsToUse.length}道题`);
      loadFavoriteStatus(questionsToUse.map((q) => q.id));
    } catch (error) {
      toast.error(error.message || '生成错题重练卷失败');
    } finally {
      setReportLoading(false);
    }
  };

  const startMistakeReview = async () => {
    if (pendingMistakes.length === 0) {
      toast.error('没有待复习的错题');
      return;
    }
    try {
      setLoadingQuiz(true);
      const draft = await getDraft(token, 'review').catch(() => null);
      if (draft && draft.questions && draft.questions.length > 0) {
        setDraftData(draft);
        setPendingStartMode('review');
        setShowDraftDialog(true);
        setLoadingQuiz(false);
        return;
      }
      await startFreshQuiz('review');
    } catch (error) {
      toast.error(error.message || '生成错题重练卷失败');
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

  const doSaveDraft = async (questionsToSave, answersToSave, mode) => {
    try {
      setSaveDraftStatus('saving');
      await saveDraft(token, mode, questionsToSave, answersToSave);
      setSaveDraftStatus('saved');
      if (saveDraftStatusTimerRef.current) {
        clearTimeout(saveDraftStatusTimerRef.current);
      }
      saveDraftStatusTimerRef.current = setTimeout(() => {
        setSaveDraftStatus('idle');
      }, 2000);
    } catch (error) {
      setSaveDraftStatus('error');
      console.error('save draft failed', error);
    }
  };

  const debouncedSaveDraft = (questionsToSave, answersToSave, mode) => {
    if (saveDraftTimerRef.current) {
      clearTimeout(saveDraftTimerRef.current);
    }
    saveDraftTimerRef.current = setTimeout(() => {
      doSaveDraft(questionsToSave, answersToSave, mode);
    }, 800);
  };

  const resumeDraft = () => {
    if (!draftData) return;
    const restoredAnswers = {};
    if (draftData.answers) {
      Object.keys(draftData.answers).forEach((key) => {
        const numKey = Number(key);
        if (!isNaN(numKey)) {
          restoredAnswers[numKey] = draftData.answers[key];
        }
      });
    }
    setQuestions(draftData.questions);
    setAnswers(restoredAnswers);
    setQuizMode(draftData.quizMode || 'normal');
    setLastResult(null);
    setMistakeReviewResult(null);
    setShowDraftDialog(false);
    setDraftData(null);
    setPendingStartMode(null);
    toast.success('已恢复上次答题进度');
  };

  const handleDiscardAndRestart = async () => {
    const mode = pendingStartMode || draftData?.quizMode || 'normal';
    try {
      await clearDraft(token, mode);
    } catch (error) {
      console.error('discard draft failed', error);
    }
    setShowDraftDialog(false);
    setDraftData(null);
    setPendingStartMode(null);
    await startFreshQuiz(mode);
  };

  const handleCloseDraftDialog = () => {
    setShowDraftDialog(false);
    setDraftData(null);
    setPendingStartMode(null);
  };

  const handleAnswer = (questionId, answerData) => {
    setAnswers((prev) => {
      const newAnswers = {
        ...prev,
        [questionId]: { ...prev[questionId], ...answerData },
      };
      if (questions.length > 0) {
        debouncedSaveDraft(questions, newAnswers, quizMode);
      }
      return newAnswers;
    });
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

  const handleAutoSubmit = async () => {
    if (hasAutoSubmittedRef.current || isAutoSubmitting || submitting) return;
    if (!questions.length || quizMode === 'review') return;

    hasAutoSubmittedRef.current = true;
    setIsAutoSubmitting(true);
    toast('考试时间到，自动提交中...', { icon: '⏰' });

    try {
      const answersPayload = buildAnswersPayload();
      const payload = { answers: answersPayload };
      const result = await apiRequest('/student/submit', {
        method: 'POST',
        token,
        body: payload,
      });
      setLastResult(result);
      setDeadline(null);
      setStartedAt(null);

      try {
        const questionIds = questions.map((q) => q.id);
        const explanations = await fetchExplanations(token, questionIds);
        setNormalQuizExplanations(explanations || []);
        const kpMap = {};
        (explanations || []).forEach((e) => {
          if (e.questionId && e.knowledgePoints) {
            kpMap[e.questionId] = e.knowledgePoints;
          }
        });
        setNormalQuizKPMap(kpMap);
      } catch (exErr) {
        console.warn('获取解析失败:', exErr);
      }

      await loadStudentData();
      if (result.timeout) {
        toast.success(`已超时自动提交：${result.score}/${result.total}`);
      } else {
        toast.success(`提交成功：${result.score}/${result.total}`);
      }
    } catch (error) {
      toast.error(error.message || '自动提交失败');
    } finally {
      setIsAutoSubmitting(false);
    }
  };

  const submitQuiz = async () => {
    if (!questions.length) {
      toast.error('请先开始答题');
      return;
    }

    if (!allowEarlySubmit && deadline && new Date().getTime() < new Date(deadline).getTime()) {
      toast.error('当前考试不允许提前交卷');
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
        setDeadline(null);
        setStartedAt(null);

        if (result.timeout) {
          toast.success(`已超时提交：${result.score}/${result.total}`);
        } else {
          toast.success(`提交成功：${result.score}/${result.total}`);
        }

        try {
          const questionIds = questions.map((q) => q.id);
          const explanations = await fetchExplanations(token, questionIds);
          setNormalQuizExplanations(explanations || []);

          const kpMap = {};
          (explanations || []).forEach((e) => {
            if (e.questionId && e.knowledgePoints) {
              kpMap[e.questionId] = e.knowledgePoints;
            }
          });
          setNormalQuizKPMap(kpMap);
        } catch (exErr) {
          console.warn('获取解析失败:', exErr);
        }

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
      <header className="mx-auto mb-6 max-w-7xl rounded-3xl border border-white/70 bg-white/90 px-6 py-5 shadow-card">
        <div className="mb-4 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <p className="text-xs uppercase tracking-[0.25em] text-emerald-700">Student Console</p>
            <h1 className="mt-1 text-2xl font-bold text-slate-800">学生答题中心</h1>
            <p className="text-sm text-slate-600">当前班级：{className}</p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {activeTab === 'quiz' && (
              <>
                <select
                  className="select select-bordered select-sm w-48"
                  value={selectedExamConfigId || ''}
                  onChange={(e) => setSelectedExamConfigId(Number(e.target.value) || null)}
                  disabled={loadingQuiz || questions.length > 0}
                >
                  {examConfigs.map((config) => (
                    <option key={config.id} value={config.id}>
                      {config.name}
                    </option>
                  ))}
                </select>
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
              </>
            )}
            <button className="btn btn-neutral" onClick={onLogout}>
              退出登录
            </button>
          </div>
        </div>
        <div className="tabs tabs-boxed bg-slate-100/50">
          <button
            className={`tab ${activeTab === 'quiz' ? 'tab-active' : ''}`}
            onClick={() => setActiveTab('quiz')}
          >
            📝 答题中心
          </button>
          <button
            className={`tab ${activeTab === 'favorites' ? 'tab-active' : ''}`}
            onClick={() => setActiveTab('favorites')}
          >
            ⭐ 我的收藏与题集
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
      ) : activeTab === 'favorites' ? (
        <FavoritesAndSets
          token={token}
          onStartSetPractice={startSetPractice}
        />
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
              <div className="mb-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex items-center gap-3">
                  <h2 className="text-lg font-semibold text-slate-800">
                    {quizMode === 'review' ? '错题重练' : '在线答题'}
                  </h2>
                  {questions.length > 0 && (
                    <span
                      className={`text-xs transition-opacity ${
                        saveDraftStatus === 'saved'
                          ? 'text-emerald-600 opacity-100'
                          : saveDraftStatus === 'saving'
                          ? 'text-slate-400 opacity-100'
                          : saveDraftStatus === 'error'
                          ? 'text-red-500 opacity-100'
                          : 'opacity-0'
                      }`}
                    >
                      {saveDraftStatus === 'saved'
                        ? '✓ 已保存'
                        : saveDraftStatus === 'saving'
                        ? '保存中...'
                        : saveDraftStatus === 'error'
                        ? '保存失败'
                        : ''}
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-3">
                  {deadline && questions.length > 0 && !lastResult && (
                    <CountdownTimer
                      deadline={deadline}
                      onTimeout={handleAutoSubmit}
                      allowEarlySubmit={allowEarlySubmit}
                    />
                  )}
                  <button
                    className="btn btn-sm btn-secondary"
                    onClick={submitQuiz}
                    disabled={submitting || isAutoSubmitting || !questions.length || (deadline && !allowEarlySubmit && !lastResult)}
                  >
                    {submitting || isAutoSubmitting ? '提交中...' : '提交本次答案'}
                  </button>
                </div>
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
                      favorited={!!favoriteStatus[question.id]}
                      onToggleFavorite={handleToggleFavorite}
                    />
                  ))}
                </div>
              )}

              {lastResult && quizMode === 'normal' ? (
                <div className="mt-4 space-y-4">
                  <div className={`rounded-xl border px-4 py-3 text-sm ${
                    lastResult.timeout
                      ? 'border-amber-300 bg-amber-50 text-amber-800'
                      : 'border-emerald-200 bg-emerald-50 text-emerald-700'
                  }`}>
                    <div className="flex items-center gap-2">
                      <span>本次成绩：{lastResult.score}/{lastResult.total}（正确率 {lastResult.rate}）</span>
                      {lastResult.timeout && (
                        <span className="badge badge-warning badge-xs">超时提交</span>
                      )}
                    </div>
                    {lastResult.startedAt && lastResult.deadline && (
                      <div className="mt-1 text-xs opacity-80">
                        开始时间：{new Date(lastResult.startedAt).toLocaleTimeString()}，
                        截止时间：{new Date(lastResult.deadline).toLocaleTimeString()}
                      </div>
                    )}
                  </div>
                  {lastResult.details && lastResult.details.length > 0 && (
                    <ResultDetail
                      questions={questions}
                      details={lastResult.details}
                      explanations={normalQuizExplanations}
                      knowledgePointsMap={normalQuizKPMap}
                    />
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
                <h2 className="text-lg font-semibold text-slate-800">🏆 班级排行榜</h2>
                <span className="text-xs text-slate-500">
                  共 {leaderboard?.total || 0} 人
                </span>
              </div>

              <div className="mb-3 tabs tabs-boxed bg-slate-100/50 tabs-sm">
                <button
                  className={`tab ${leaderboardScoreType === 'highest' ? 'tab-active' : ''}`}
                  onClick={() => setLeaderboardScoreType('highest')}
                >
                  最高分
                </button>
                <button
                  className={`tab ${leaderboardScoreType === 'average' ? 'tab-active' : ''}`}
                  onClick={() => setLeaderboardScoreType('average')}
                >
                  平均正确率
                </button>
                <button
                  className={`tab ${leaderboardScoreType === 'weighted' ? 'tab-active' : ''}`}
                  onClick={() => setLeaderboardScoreType('weighted')}
                >
                  近5次加权
                </button>
              </div>

              {leaderboard?.currentRank && (
                <div className="mb-3 rounded-xl border-2 border-sky-300 bg-sky-50 p-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <span className="flex h-10 w-10 items-center justify-center rounded-full bg-sky-500 text-sm font-bold text-white">
                        {leaderboard.currentRank.rank}
                      </span>
                      <div>
                        <p className="text-sm font-medium text-slate-800">
                          我的排名
                        </p>
                        <p className="text-xs text-slate-500">
                          得分：{leaderboard.currentRank.scoreDisplay}
                        </p>
                      </div>
                    </div>
                    <div className="text-right">
                      {leaderboard.hasPrev ? (
                        <div className="text-xs text-amber-600">
                          距上一名差
                          <span className="ml-1 font-bold text-amber-700">
                            {leaderboard.gapToPrev.toFixed(1)}
                          </span>
                          <span className="ml-0.5">分</span>
                        </div>
                      ) : (
                        <span className="text-xs text-emerald-600 font-medium">
                          🎉 第一名！
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              )}

              <div className="max-h-[280px] space-y-1.5 overflow-auto rounded-xl">
                {leaderboardLoading ? (
                  <div className="flex items-center justify-center py-8">
                    <span className="loading loading-spinner loading-md text-sky-600"></span>
                  </div>
                ) : leaderboard?.items?.length > 0 ? (
                  leaderboard.items.map((item, idx) => {
                    const isTop3 = item.rank <= 3;
                    const medalEmoji = item.rank === 1 ? '🥇' : item.rank === 2 ? '🥈' : item.rank === 3 ? '🥉' : null;

                    return (
                      <div
                        key={`${item.userId}-${idx}`}
                        className={`flex items-center gap-3 rounded-lg px-3 py-2 transition-colors ${
                          item.isCurrentUser
                            ? 'bg-sky-100 border border-sky-300'
                            : isTop3
                            ? 'bg-amber-50/50'
                            : 'hover:bg-slate-50'
                        }`}
                      >
                        <div className="w-8 text-center">
                          {medalEmoji ? (
                            <span className="text-xl">{medalEmoji}</span>
                          ) : (
                            <span className="text-sm font-medium text-slate-500">
                              {item.rank}
                            </span>
                          )}
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className={`text-sm font-medium truncate ${
                            item.isCurrentUser ? 'text-sky-700' : 'text-slate-700'
                          }`}>
                            {item.username}
                            {item.isCurrentUser && (
                              <span className="ml-1 text-xs text-sky-500">（我）</span>
                            )}
                          </p>
                          <p className="text-xs text-slate-400">
                            答题 {item.attemptCount} 次 · 正确率 {item.correctRate}
                          </p>
                        </div>
                        <div className="text-right">
                          <span className={`text-sm font-bold ${
                            isTop3 ? 'text-amber-600' : 'text-slate-700'
                          }`}>
                            {item.scoreDisplay}
                          </span>
                          <p className="text-xs text-slate-400">分</p>
                        </div>
                      </div>
                    );
                  })
                ) : (
                  <div className="rounded-xl border border-dashed border-slate-300 py-8 text-center text-sm text-slate-500">
                    暂无排名数据
                  </div>
                )}
              </div>
            </article>

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
                      <th>操作</th>
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
                        <td>
                          <button
                            className="btn btn-xs btn-primary btn-outline"
                            onClick={() => handleViewReport(item.id)}
                          >
                            查看报告
                          </button>
                        </td>
                      </tr>
                    ))}
                    {!attempts.length ? (
                      <tr>
                        <td colSpan={3} className="text-center text-slate-500">
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

      {showDraftDialog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="mx-4 w-full max-w-md rounded-2xl bg-white p-6 shadow-xl">
            <h3 className="text-lg font-semibold text-slate-800">发现未完成的答题</h3>
            <p className="mt-2 text-sm text-slate-600">
              你有一份未完成的{ draftData?.quizMode === 'review' ? '错题重练' : '答题' }草稿，
              {draftData?.updatedAt && `最后更新于 ${draftData.updatedAt}`}
            </p>
            <p className="mt-1 text-sm text-slate-500">
              共 {draftData?.questions?.length || 0} 道题
            </p>
            <div className="mt-6 flex gap-3">
              <button
                className="btn btn-primary flex-1"
                onClick={resumeDraft}
              >
                继续上次
              </button>
              <button
                className="btn btn-outline flex-1"
                onClick={handleDiscardAndRestart}
              >
                放弃重开
              </button>
            </div>
          </div>
        </div>
      )}

      <AttemptReportModal
        report={reportData}
        loading={reportLoading}
        onClose={handleCloseReport}
        onReviewWrong={handleReviewWrongFromReport}
      />
    </div>
  );
}
