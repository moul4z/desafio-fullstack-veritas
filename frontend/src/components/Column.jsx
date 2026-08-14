import { useState } from 'react'
import TaskCard from './TaskCard'
import { PlusIcon } from './Icons'

export default function Column({
  column,
  tasks,
  dragging,
  pendingIds,
  onAdd,
  onEdit,
  onDelete,
  onMove,
  onDrop,
  onDragStart,
  onDragEnd,
}) {
  const [isOver, setIsOver] = useState(false)
  const canReceive = Boolean(dragging && dragging.status !== column.id)

  const handleDragEnter = (event) => {
    if (!canReceive) return
    event.preventDefault()
    setIsOver(true)
  }

  const handleDragOver = (event) => {
    if (!canReceive) return
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
    if (!isOver) setIsOver(true)
  }

  const handleDragLeave = (event) => {
    const rect = event.currentTarget.getBoundingClientRect()
    const outside =
      event.clientX <= rect.left ||
      event.clientX >= rect.right ||
      event.clientY <= rect.top ||
      event.clientY >= rect.bottom

    if (outside) setIsOver(false)
  }

  const handleDrop = (event) => {
    event.preventDefault()

    let payload = null
    try {
      const raw = event.dataTransfer.getData('application/x-veritas-task')
      if (raw) payload = JSON.parse(raw)
    } catch {
      payload = null
    }

    const id = payload?.id || event.dataTransfer.getData('text/plain') || dragging?.id
    const sourceStatus = payload?.status || dragging?.status

    setIsOver(false)
    if (id && sourceStatus !== column.id) onDrop(id, column.id)
  }

  return (
    <section
      className={`column${isOver ? ' column--target' : ''}`}
      style={{ '--accent': column.accent }}
      aria-labelledby={`column-${column.id}`}
      onDragEnter={handleDragEnter}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <header className="column__header">
        <div className="column__heading">
          <span className="column__dot" aria-hidden="true" />
          <h2 className="column__title" id={`column-${column.id}`}>
            {column.title}
          </h2>
        </div>

        <div className="column__tools">
          <span className="column__count" aria-label={`${tasks.length} tarefas`}>
            {tasks.length}
          </span>
          <button
            type="button"
            className="column__add-icon"
            onClick={() => onAdd(column.id)}
            title={`Adicionar tarefa em ${column.title}`}
            aria-label={`Adicionar tarefa em ${column.title}`}
          >
            <PlusIcon />
          </button>
        </div>
      </header>

      <div className="column__body">
        {tasks.length === 0 ? (
          <div className="column__empty">
            <span className="column__empty-icon">+</span>
            <span>{column.empty}</span>
          </div>
        ) : (
          <ul className="column__list">
            {tasks.map((task) => (
              <li key={task.id} data-task-id={task.id}>
                <TaskCard
                  task={task}
                  isPending={pendingIds.includes(task.id)}
                  isDragging={dragging?.id === task.id}
                  onEdit={onEdit}
                  onDelete={onDelete}
                  onMove={onMove}
                  onDragStart={onDragStart}
                  onDragEnd={onDragEnd}
                />
              </li>
            ))}
          </ul>
        )}

        {isOver && (
          <div className="column__drop-hint" aria-hidden="true">
            <span>Solte aqui</span>
            <small>Mover para {column.title}</small>
          </div>
        )}
      </div>

      <button type="button" className="column__quick-add" onClick={() => onAdd(column.id)}>
        <PlusIcon />
        Adicionar tarefa
      </button>
    </section>
  )
}
