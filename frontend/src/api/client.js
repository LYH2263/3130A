const API_BASE = import.meta.env.VITE_API_BASE || '/api';

async function parseResponse(response) {
  const contentType = response.headers.get('content-type') || '';
  const isJSON = contentType.includes('application/json');
  const payload = isJSON ? await response.json() : null;

  if (!response.ok) {
    const message = payload?.message || `Request failed: ${response.status}`;
    throw new Error(message);
  }
  return payload;
}

export async function apiRequest(path, { method = 'GET', token, body, isForm = false } = {}) {
  const headers = {};
  if (!isForm) {
    headers['Content-Type'] = 'application/json';
  }
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: isForm ? body : body ? JSON.stringify(body) : undefined,
  });

  return parseResponse(response);
}

export async function fetchCategories(token) {
  return apiRequest('/teacher/categories', { token });
}

export async function createCategory(token, data) {
  return apiRequest('/teacher/categories', { method: 'POST', token, body: data });
}

export async function updateCategory(token, id, data) {
  return apiRequest(`/teacher/categories/${id}`, { method: 'PUT', token, body: data });
}

export async function deleteCategory(token, id) {
  return apiRequest(`/teacher/categories/${id}`, { method: 'DELETE', token });
}

export async function fetchTags(token) {
  return apiRequest('/teacher/tags', { token });
}

export async function createTag(token, data) {
  return apiRequest('/teacher/tags', { method: 'POST', token, body: data });
}

export async function updateTag(token, id, data) {
  return apiRequest(`/teacher/tags/${id}`, { method: 'PUT', token, body: data });
}

export async function deleteTag(token, id) {
  return apiRequest(`/teacher/tags/${id}`, { method: 'DELETE', token });
}

export async function fetchKnowledgePoints(token) {
  return apiRequest('/teacher/knowledge-points', { token });
}

export async function createKnowledgePoint(token, data) {
  return apiRequest('/teacher/knowledge-points', { method: 'POST', token, body: data });
}

export async function updateKnowledgePoint(token, id, data) {
  return apiRequest(`/teacher/knowledge-points/${id}`, { method: 'PUT', token, body: data });
}

export async function deleteKnowledgePoint(token, id) {
  return apiRequest(`/teacher/knowledge-points/${id}`, { method: 'DELETE', token });
}

export async function fetchQuestionExplanation(token, questionId) {
  return apiRequest(`/student/questions/${questionId}/explanation`, { token });
}

export async function fetchExplanations(token, questionIds = []) {
  const query = new URLSearchParams();
  questionIds.forEach((id) => query.append('questionIds', id));
  const queryString = query.toString();
  return apiRequest(`/student/explanations${queryString ? '?' + queryString : ''}`, { token });
}

export async function fetchQuestions(token, params = {}) {
  const query = new URLSearchParams();
  if (params.keyword) query.append('keyword', params.keyword);
  if (params.categoryId) query.append('categoryId', params.categoryId);
  if (params.tagIds && params.tagIds.length > 0) {
    params.tagIds.forEach((id) => query.append('tagIds', id));
  }
  if (params.tagMode) query.append('tagMode', params.tagMode);
  if (params.createdBy) query.append('createdBy', params.createdBy);
  if (params.createdFrom) query.append('createdFrom', params.createdFrom);
  if (params.createdTo) query.append('createdTo', params.createdTo);
  if (params.hasAnswerError !== undefined && params.hasAnswerError !== null) {
    query.append('hasAnswerError', params.hasAnswerError);
  }
  if (params.sortBy) query.append('sortBy', params.sortBy);
  if (params.sortOrder) query.append('sortOrder', params.sortOrder);
  if (params.page) query.append('page', params.page);
  if (params.pageSize) query.append('pageSize', params.pageSize);

  const queryString = query.toString();
  return apiRequest(`/teacher/questions${queryString ? '?' + queryString : ''}`, { token });
}

export async function fetchQuestion(token, id) {
  return apiRequest(`/teacher/questions/${id}`, { token });
}

export async function fetchMistakeReviewQuiz(token, limit = 10) {
  return apiRequest(`/student/mistake-review/quiz?limit=${limit}`, { token });
}

export async function submitMistakeReview(token, answers) {
  return apiRequest('/student/mistake-review/submit', {
    method: 'POST',
    token,
    body: { answers },
  });
}

export async function saveDraft(token, quizMode, questions, answers) {
  return apiRequest('/student/draft', {
    method: 'POST',
    token,
    body: {
      quizMode,
      questions,
      answers,
    },
  });
}

export async function getDraft(token, quizMode = 'normal') {
  return apiRequest(`/student/draft?mode=${quizMode}`, { token });
}

export async function clearDraft(token, quizMode = 'normal') {
  return apiRequest(`/student/draft?mode=${quizMode}`, {
    method: 'DELETE',
    token,
  });
}

export async function toggleFavorite(token, questionId) {
  return apiRequest(`/student/favorites/${questionId}`, {
    method: 'POST',
    token,
  });
}

export async function unfavoriteQuestion(token, questionId) {
  return apiRequest(`/student/favorites/${questionId}`, {
    method: 'DELETE',
    token,
  });
}

export async function fetchFavorites(token) {
  return apiRequest('/student/favorites', { token });
}

export async function fetchFavoriteStatus(token, questionIds = []) {
  const query = new URLSearchParams();
  questionIds.forEach((id) => query.append('questionIds', id));
  const queryString = query.toString();
  return apiRequest(`/student/favorites/status${queryString ? '?' + queryString : ''}`, { token });
}

export async function fetchFavoriteQuiz(token) {
  return apiRequest('/student/favorites/quiz', { token });
}

export async function fetchQuestionSets(token) {
  return apiRequest('/student/question-sets', { token });
}

export async function fetchQuestionSet(token, setId) {
  return apiRequest(`/student/question-sets/${setId}`, { token });
}

export async function createQuestionSet(token, data) {
  return apiRequest('/student/question-sets', {
    method: 'POST',
    token,
    body: data,
  });
}

export async function updateQuestionSet(token, setId, data) {
  return apiRequest(`/student/question-sets/${setId}`, {
    method: 'PUT',
    token,
    body: data,
  });
}

export async function deleteQuestionSet(token, setId) {
  return apiRequest(`/student/question-sets/${setId}`, {
    method: 'DELETE',
    token,
  });
}

export async function addQuestionsToSet(token, setId, questionIds) {
  return apiRequest(`/student/question-sets/${setId}/questions`, {
    method: 'POST',
    token,
    body: { questionIds },
  });
}

export async function removeQuestionFromSet(token, setId, questionId) {
  return apiRequest(`/student/question-sets/${setId}/questions/${questionId}`, {
    method: 'DELETE',
    token,
  });
}

export async function reorderSetQuestions(token, setId, questionIds) {
  return apiRequest(`/student/question-sets/${setId}/reorder`, {
    method: 'POST',
    token,
    body: { questionIds },
  });
}

export async function fetchSetQuiz(token, setId) {
  return apiRequest(`/student/question-sets/${setId}/quiz`, { token });
}

export async function fetchAttemptDetail(token, attemptId) {
  return apiRequest(`/student/attempts/${attemptId}`, { token });
}

export async function fetchExamConfigs(token) {
  return apiRequest('/student/exam-configs', { token });
}

export async function fetchTeacherExamConfigs(token) {
  return apiRequest('/teacher/exam-configs', { token });
}

export async function createExamConfig(token, data) {
  return apiRequest('/teacher/exam-configs', {
    method: 'POST',
    token,
    body: data,
  });
}

export async function updateExamConfig(token, id, data) {
  return apiRequest(`/teacher/exam-configs/${id}`, {
    method: 'PUT',
    token,
    body: data,
  });
}

export async function deleteExamConfig(token, id) {
  return apiRequest(`/teacher/exam-configs/${id}`, {
    method: 'DELETE',
    token,
  });
}

export async function startQuiz(token, data = {}) {
  return apiRequest('/student/start-quiz', {
    method: 'POST',
    token,
    body: data,
  });
}
