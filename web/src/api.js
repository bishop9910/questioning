// 极简 fetch 封装：同源 /api/v1，token 放 localStorage
const BASE = '/api/v1'

export function getToken() {
  return localStorage.getItem('q_token') || ''
}

export function setToken(t) {
  if (t) localStorage.setItem('q_token', t)
  else localStorage.removeItem('q_token')
}

export async function request(path, { method = 'GET', body, auth = true } = {}) {
  const headers = {}
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  if (auth) {
    const t = getToken()
    if (t) headers['Authorization'] = `Bearer ${t}`
  }
  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  if (res.status === 401 && auth) {
    setToken('')
    if (location.pathname !== '/login') location.href = '/login'
  }
  let data = null
  try {
    data = await res.json()
  } catch {
    data = null
  }
  if (!res.ok) {
    const err = new Error((data && data.message) || `请求失败（${res.status}）`)
    err.status = res.status
    throw err
  }
  return data
}

export const api = {
  login: (b) => request('/auth/login', { method: 'POST', body: b, auth: false }),
  register: (b) => request('/auth/register', { method: 'POST', body: b, auth: false }),
  me: () => request('/auth/me'),
  surveys: () => request('/surveys'),
  survey: (id) => request(`/surveys/${id}`),
  createSurvey: (b) => request('/surveys', { method: 'POST', body: b }),
  updateSurvey: (id, b) => request(`/surveys/${id}`, { method: 'PUT', body: b }),
  deleteSurvey: (id) => request(`/surveys/${id}`, { method: 'DELETE' }),
  setStatus: (id, status) => request(`/surveys/${id}/status`, { method: 'PUT', body: { status } }),
  saveQuestions: (id, questions) =>
    request(`/surveys/${id}/questions`, { method: 'PUT', body: { questions } }),
  publicSurvey: (id) => request(`/public/surveys/${id}`, { auth: false }),
  submit: (id, b) => request(`/public/surveys/${id}/submit`, { method: 'POST', body: b, auth: false }),
  stats: (id) => request(`/surveys/${id}/stats`),
  responses: (id, page, size) => request(`/surveys/${id}/responses?page=${page}&size=${size}`),
  exportCsv: (id) =>
    fetch(`${BASE}/surveys/${id}/export.csv`, {
      headers: { Authorization: `Bearer ${getToken()}` },
    }),
}

export function deviceOf() {
  const ua = navigator.userAgent
  if (/iPhone/.test(ua)) return 'iPhone'
  if (/iPad/.test(ua)) return 'iPad'
  if (/Android/.test(ua)) return 'Android'
  if (/Windows|Macintosh|Linux/.test(ua)) return '桌面端'
  return '手机'
}
