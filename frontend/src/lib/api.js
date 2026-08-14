const BASE_URL = (import.meta.env.VITE_API_URL ?? 'http://localhost:8080').replace(/\/$/, '')

/**
 * Erro de API com o status HTTP e os erros por campo devolvidos pelo backend.
 * O formulário usa `fields` para destacar o input problemático.
 */
export class ApiError extends Error {
  constructor(message, { status = 0, fields = null } = {}) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.fields = fields
    this.isOffline = status === 0
  }
}

async function request(path, { method = 'GET', body } = {}) {
  let response
  try {
    response = await fetch(`${BASE_URL}${path}`, {
      method,
      headers: body ? { 'Content-Type': 'application/json' } : undefined,
      body: body ? JSON.stringify(body) : undefined,
    })
  } catch {
    // fetch só rejeita por falha de rede: a API está fora do ar ou o CORS barrou.
    throw new ApiError('Não foi possível falar com a API. Verifique se o backend está rodando.', { status: 0 })
  }

  if (response.status === 204) return null

  const payload = await response.json().catch(() => null)

  if (!response.ok) {
    throw new ApiError(payload?.error?.message ?? `A API respondeu com o status ${response.status}.`, {
      status: response.status,
      fields: payload?.error?.fields ?? null,
    })
  }

  return payload
}

export const api = {
  health: () => request('/health'),
  listTasks: () => request('/tasks'),
  createTask: (data) => request('/tasks', { method: 'POST', body: data }),
  updateTask: (id, data) => request(`/tasks/${id}`, { method: 'PUT', body: data }),
  deleteTask: (id) => request(`/tasks/${id}`, { method: 'DELETE' }),
}

export { BASE_URL }
