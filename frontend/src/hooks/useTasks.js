import { useCallback, useEffect, useMemo, useState } from 'react'
import { api, ApiError } from '../lib/api'
import { COLUMN_IDS } from '../lib/columns'

const sortBoard = (tasks) =>
  [...tasks].sort(
    (a, b) => COLUMN_IDS.indexOf(a.status) - COLUMN_IDS.indexOf(b.status) || a.position - b.position,
  )

/**
 * Centraliza o estado das tarefas e as chamadas à API.
 * Os componentes recebem apenas dados prontos e funções de ação.
 */
export function useTasks() {
  const [tasks, setTasks] = useState([])
  const [status, setStatus] = useState('loading') // loading | ready | error
  const [loadError, setLoadError] = useState(null)
  const [pendingIds, setPendingIds] = useState([])
  const [notice, setNotice] = useState(null)

  const markPending = useCallback((id, isPending) => {
    setPendingIds((current) =>
      isPending ? [...new Set([...current, id])] : current.filter((pendingId) => pendingId !== id),
    )
  }, [])

  const load = useCallback(async ({ showLoading = true } = {}) => {
    if (showLoading) setStatus('loading')
    setLoadError(null)

    try {
      const data = await api.listTasks()
      setTasks(sortBoard(data ?? []))
      setStatus('ready')
      return true
    } catch (error) {
      const message = error instanceof ApiError ? error.message : 'Erro inesperado ao carregar o quadro.'
      setLoadError(message)
      if (showLoading) setStatus('error')
      return false
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const createTask = useCallback(async (input) => {
    const created = await api.createTask(input)
    setTasks((current) => sortBoard([...current, created]))
    setNotice({ tone: 'success', message: 'Tarefa criada.' })
    return created
  }, [])

  const editTask = useCallback(
    async (id, input) => {
      markPending(id, true)
      try {
        const updated = await api.updateTask(id, input)
        setTasks((current) => sortBoard(current.map((task) => (task.id === id ? updated : task))))
        setNotice({ tone: 'success', message: 'Tarefa atualizada.' })
        return updated
      } finally {
        markPending(id, false)
      }
    },
    [markPending],
  )

  const moveTask = useCallback(
    async (id, targetStatus, targetIndex) => {
      markPending(id, true)
      try {
        await api.updateTask(id, { status: targetStatus, position: targetIndex })
        // Recarrega o quadro porque mover uma tarefa também pode alterar a
        // posição das demais tarefas da coluna.
        const refreshed = await load({ showLoading: false })
        setNotice(
          refreshed
            ? { tone: 'success', message: 'Tarefa movida.' }
            : { tone: 'error', message: 'A tarefa foi salva, mas o quadro não pôde ser recarregado.' },
        )
      } catch (error) {
        setNotice({
          tone: 'error',
          message: error instanceof ApiError ? error.message : 'Não foi possível mover a tarefa.',
        })
      } finally {
        markPending(id, false)
      }
    },
    [load, markPending],
  )

  const deleteTask = useCallback(
    async (id) => {
      markPending(id, true)
      try {
        await api.deleteTask(id)
        setTasks((current) => current.filter((task) => task.id !== id))
        setNotice({ tone: 'success', message: 'Tarefa excluída.' })
      } catch (error) {
        setNotice({
          tone: 'error',
          message: error instanceof ApiError ? error.message : 'Não foi possível excluir a tarefa.',
        })
      } finally {
        markPending(id, false)
      }
    },
    [markPending],
  )

  const columns = useMemo(() => {
    const grouped = Object.fromEntries(COLUMN_IDS.map((id) => [id, []]))
    for (const task of sortBoard(tasks)) {
      if (grouped[task.status]) grouped[task.status].push(task)
    }
    return grouped
  }, [tasks])

  return {
    tasks,
    columns,
    status,
    loadError,
    pendingIds,
    notice,
    dismissNotice: () => setNotice(null),
    reload: () => load(),
    createTask,
    editTask,
    moveTask,
    deleteTask,
  }
}
