import { COLUMN_IDS, getColumn } from '../lib/columns'
import { fullDate, relativeTime } from '../lib/format'
import { ArrowLeftIcon, ArrowRightIcon, GripIcon, PencilIcon, TrashIcon } from './Icons'

export default function TaskCard({ task, isPending, isDragging, onEdit, onDelete, onMove, onDragStart, onDragEnd }) {
  const columnIndex = COLUMN_IDS.indexOf(task.status)
  const previous = COLUMN_IDS[columnIndex - 1]
  const next = COLUMN_IDS[columnIndex + 1]
  const column = getColumn(task.status)

  const className = ['card', isDragging && 'card--dragging', isPending && 'card--pending']
    .filter(Boolean)
    .join(' ')

  return (
    <article
      className={className}
      style={{ '--card-accent': column.accent }}
      draggable={!isPending}
      onDragStart={(event) => onDragStart(event, task)}
      onDragEnd={onDragEnd}
      aria-busy={isPending || undefined}
    >
      <div className="card__header">
        <h3 className="card__title">{task.title}</h3>
        <span className="card__grip" aria-hidden="true" title="Arraste para mover">
          <GripIcon />
        </span>
      </div>

      {task.description && <p className="card__description">{task.description}</p>}

      <footer className="card__footer">
        <div className="card__meta">
          <span className="card__status-dot" aria-hidden="true" />
          <time dateTime={task.updatedAt} title={`Atualizada em ${fullDate(task.updatedAt)}`}>
            {relativeTime(task.updatedAt)}
          </time>
        </div>

        <div className="card__actions">
          <button
            type="button"
            className="icon-button"
            onClick={() => onMove(task, previous)}
            disabled={!previous || isPending}
            title={previous ? `Mover para ${getColumn(previous).title}` : 'Já está na primeira coluna'}
            aria-label={previous ? `Mover "${task.title}" para ${getColumn(previous).title}` : 'Já está na primeira coluna'}
          >
            <ArrowLeftIcon />
          </button>
          <button
            type="button"
            className="icon-button"
            onClick={() => onMove(task, next)}
            disabled={!next || isPending}
            title={next ? `Mover para ${getColumn(next).title}` : 'Já está na última coluna'}
            aria-label={next ? `Mover "${task.title}" para ${getColumn(next).title}` : 'Já está na última coluna'}
          >
            <ArrowRightIcon />
          </button>
          <span className="card__divider" aria-hidden="true" />
          <button
            type="button"
            className="icon-button"
            onClick={() => onEdit(task)}
            disabled={isPending}
            title="Editar tarefa"
            aria-label={`Editar "${task.title}"`}
          >
            <PencilIcon />
          </button>
          <button
            type="button"
            className="icon-button icon-button--danger"
            onClick={() => onDelete(task)}
            disabled={isPending}
            title="Excluir tarefa"
            aria-label={`Excluir "${task.title}"`}
          >
            <TrashIcon />
          </button>
        </div>
      </footer>
    </article>
  )
}
