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

export async function fetchQuestions(token, params = {}) {
  const query = new URLSearchParams();
  if (params.keyword) query.append('keyword', params.keyword);
  if (params.categoryId) query.append('categoryId', params.categoryId);
  if (params.tagIds && params.tagIds.length > 0) {
    params.tagIds.forEach((id) => query.append('tagIds', id));
  }
  if (params.tagMode) query.append('tagMode', params.tagMode);
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
