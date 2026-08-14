import { useCallback, useMemo, useState } from 'react'
import Column from './components/Column'
import ConfirmDialog from './components/ConfirmDialog'
import TaskDialog from './components/TaskDialog'
import Toast from './components/Toast'
import { AlertIcon, PlusIcon } from './components/Icons'
import { useTasks } from './hooks/useTasks'
import { COLUMNS } from './lib/columns'
import { BASE_URL } from './lib/api'

export default function App() {
  const {
    columns,
    tasks,
    status,
    loadError,
    pendingIds,
    notice,
    dismissNotice,
    reload,
    createTask,
    editTask,
    moveTask,
    deleteTask,
  } = useTasks()

  const [dialog, setDialog] = useState({ open: false, mode: 'create', task: null, defaultStatus: 'todo' })
  const [confirming, setConfirming] = useState(null)
  const [dragging, setDragging] = useState(null)

  const openCreate = (defaultStatus = 'todo') =>
    setDialog({ open: true, mode: 'create', task: null, defaultStatus })

  const openEdit = (task) => setDialog({ open: true, mode: 'edit', task, defaultStatus: task.status })

  const closeDialog = () => setDialog((current) => ({ ...current, open: false }))

  const handleSubmit = (input) =>
    dialog.mode === 'edit' ? editTask(dialog.task.id, input) : createTask(input)

  const handleMoveByButton = (task, targetStatus) => {
    if (!targetStatus) return
    moveTask(task.id, targetStatus, columns[targetStatus]?.length ?? 0)
  }

  const handleDragStart = useCallback((event, task) => {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', task.id)
    event.dataTransfer.setData(
      'application/x-veritas-task',
      JSON.stringify({ id: task.id, status: task.status }),
    )

    // O cartão permanece no DOM e só muda visualmente durante o arrasto.
    // Atualizar o estado aqui, sem atraso, evita perder drops muito rápidos.
    setDragging(task)
  }, [])

  const handleDragEnd = useCallback(() => setDragging(null), [])

  const handleDrop = (id, targetStatus) => {
    if (!id || dragging?.status === targetStatus) {
      setDragging(null)
      return
    }

    const targetIndex = columns[targetStatus]?.length ?? 0
    setDragging(null)
    moveTask(id, targetStatus, targetIndex)
  }

  const summary = useMemo(() => {
    const total = tasks.length
    const doing = columns.doing?.length ?? 0
    const done = columns.done?.length ?? 0
    const completion = total ? Math.round((done / total) * 100) : 0
    return { total, doing, done, completion }
  }, [tasks.length, columns.doing, columns.done])

  return (
    <div className="app">
      <header className="topbar">
        <div className="topbar__copy">
          <div className="brandline">
            <img className="brandline__logo" src="/veritas-logo.png" alt="Veritas" />
            <span className="brandline__name">Veritas</span>
            <span className="brandline__separator">/</span>
            <span>Desafio Full Stack</span>
          </div>
          <h1 className="topbar__title">Quadro de tarefas</h1>
          <p className="topbar__subtitle">Organize o trabalho e mova cada tarefa conforme ela avança.</p>
        </div>

        <button type="button" className="button button--primary button--new" onClick={() => openCreate('todo')}>
          <PlusIcon />
          Nova tarefa
        </button>
      </header>

      {status === 'ready' && (
        <section className="board-summary" aria-label="Resumo do quadro">
          <div className="summary-stats">
            <div className="summary-stat">
              <strong>{summary.total}</strong>
              <span>{summary.total === 1 ? 'tarefa' : 'tarefas'}</span>
            </div>
            <span className="summary-divider" aria-hidden="true" />
            <div className="summary-stat">
              <strong>{summary.doing}</strong>
              <span>em andamento</span>
            </div>
            <span className="summary-divider" aria-hidden="true" />
            <div className="summary-stat summary-stat--done">
              <strong>{summary.done}</strong>
              <span>concluídas</span>
            </div>
          </div>

          <div className="summary-progress">
            <div className="summary-progress__label">
              <span>Progresso</span>
              <strong>{summary.completion}%</strong>
            </div>
            <div className="summary-progress__track" aria-hidden="true">
              <span style={{ width: `${summary.completion}%` }} />
            </div>
          </div>
        </section>
      )}

      <main className="board-wrapper">
        {status === 'loading' && <BoardSkeleton />}

        {status === 'error' && (
          <div className="state state--error" role="alert">
            <AlertIcon width={22} height={22} />
            <div>
              <h2>Não foi possível carregar o quadro</h2>
              <p>{loadError}</p>
              <p className="state__hint">
                A API esperada está em <code>{BASE_URL}</code>. Suba o backend com <code>go run .</code> na pasta{' '}
                <code>/backend</code>.
              </p>
            </div>
            <button type="button" className="button button--ghost" onClick={reload}>
              Tentar de novo
            </button>
          </div>
        )}

        {status === 'ready' && (
          <div className={`board${dragging ? ' board--dragging' : ''}`}>
            {COLUMNS.map((column) => (
              <Column
                key={column.id}
                column={column}
                tasks={columns[column.id] ?? []}
                dragging={dragging}
                pendingIds={pendingIds}
                onAdd={openCreate}
                onEdit={openEdit}
                onDelete={setConfirming}
                onMove={handleMoveByButton}
                onDrop={handleDrop}
                onDragStart={handleDragStart}
                onDragEnd={handleDragEnd}
              />
            ))}
          </div>
        )}
      </main>

      <TaskDialog
        open={dialog.open}
        mode={dialog.mode}
        task={dialog.task}
        defaultStatus={dialog.defaultStatus}
        onClose={closeDialog}
        onSubmit={handleSubmit}
      />

      <ConfirmDialog
        open={Boolean(confirming)}
        title="Excluir tarefa"
        message={`"${confirming?.title ?? ''}" será removida do quadro. Esta ação não pode ser desfeita.`}
        onConfirm={() => deleteTask(confirming.id)}
        onClose={() => setConfirming(null)}
      />

      <Toast notice={notice} onDismiss={dismissNotice} />
    </div>
  )
}

function BoardSkeleton() {
  return (
    <div className="board" aria-busy="true" aria-label="Carregando o quadro">
      {COLUMNS.map((column) => (
        <section key={column.id} className="column" style={{ '--accent': column.accent }}>
          <header className="column__header">
            <span className="column__dot" aria-hidden="true" />
            <h2 className="column__title">{column.title}</h2>
          </header>
          <div className="column__body">
            <div className="skeleton skeleton--card" />
            <div className="skeleton skeleton--card" />
          </div>
        </section>
      ))}
    </div>
  )
}
