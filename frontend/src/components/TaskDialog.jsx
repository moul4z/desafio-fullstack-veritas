import { useEffect, useRef, useState } from 'react'
import { COLUMNS } from '../lib/columns'
import { CloseIcon } from './Icons'

const MAX_TITLE = 120
const MAX_DESCRIPTION = 1000
const emptyForm = { title: '', description: '', status: 'todo' }

/**
 * Formulário de criação e edição.
 *
 * Usa o elemento <dialog> nativo: com showModal() já vêm de graça o foco preso
 * dentro do modal, o fechamento com Esc e o fundo inerte — coisas que uma div
 * com position: fixed exigiria reimplementar à mão.
 */
export default function TaskDialog({ open, mode, task, defaultStatus, onClose, onSubmit }) {
  const dialogRef = useRef(null)
  const titleRef = useRef(null)
  const [form, setForm] = useState(emptyForm)
  const [errors, setErrors] = useState({})
  const [generalError, setGeneralError] = useState(null)
  const [saving, setSaving] = useState(false)

  const isEditing = mode === 'edit'

  // Sincroniza o formulário com a tarefa escolhida sempre que o modal abre.
  useEffect(() => {
    if (!open) return
    setForm(
      isEditing && task
        ? { title: task.title, description: task.description ?? '', status: task.status }
        : { ...emptyForm, status: defaultStatus ?? 'todo' },
    )
    setErrors({})
    setGeneralError(null)
    setSaving(false)
  }, [open, isEditing, task, defaultStatus])

  // Abre e fecha o <dialog> de acordo com o estado do React.
  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) return

    if (open && !dialog.open) {
      dialog.showModal()
      // Deixa o campo pronto para digitar assim que o modal aparece.
      requestAnimationFrame(() => titleRef.current?.focus())
    } else if (!open && dialog.open) {
      dialog.close()
    }
  }, [open])

  const update = (field) => (event) => {
    setForm((current) => ({ ...current, [field]: event.target.value }))
    setErrors((current) => ({ ...current, [field]: undefined }))
  }

  const handleSubmit = async (event) => {
    event.preventDefault()

    const title = form.title.trim()
    if (!title) {
      setErrors({ title: 'O título é obrigatório.' })
      titleRef.current?.focus()
      return
    }

    setSaving(true)
    setGeneralError(null)
    try {
      await onSubmit({ title, description: form.description.trim(), status: form.status })
      onClose()
    } catch (error) {
      // O backend valida de novo: se ele reclamar de um campo, mostramos ali.
      if (error?.fields) {
        setErrors(error.fields)
      } else {
        setGeneralError(error?.message ?? 'Não foi possível salvar a tarefa.')
      }
      setSaving(false)
    }
  }

  return (
    <dialog
      ref={dialogRef}
      className="dialog"
      onCancel={(event) => {
        event.preventDefault() // deixa o React ser o dono do estado "aberto"
        if (!saving) onClose()
      }}
      onClick={(event) => {
        // Clique no backdrop (fora do conteúdo) fecha o modal.
        if (event.target === dialogRef.current && !saving) onClose()
      }}
      aria-labelledby="dialog-title"
    >
      <form className="dialog__panel" onSubmit={handleSubmit} noValidate>
        <header className="dialog__header">
          <div>
            <p className="eyebrow">{isEditing ? 'Editar' : 'Nova tarefa'}</p>
            <h2 id="dialog-title" className="dialog__title">
              {isEditing ? 'Atualizar tarefa' : 'Adicionar ao quadro'}
            </h2>
          </div>
          <button type="button" className="icon-button" onClick={onClose} disabled={saving} aria-label="Fechar">
            <CloseIcon />
          </button>
        </header>

        <div className="field">
          <label htmlFor="task-title">
            Título <span className="field__required">obrigatório</span>
          </label>
          <input
            id="task-title"
            ref={titleRef}
            type="text"
            value={form.title}
            onChange={update('title')}
            maxLength={MAX_TITLE}
            placeholder="Ex.: Revisar contrato do cliente"
            aria-invalid={Boolean(errors.title)}
            aria-describedby={errors.title ? 'task-title-error' : undefined}
            disabled={saving}
          />
          {errors.title && (
            <p className="field__error" id="task-title-error" role="alert">
              {errors.title}
            </p>
          )}
        </div>

        <div className="field">
          <label htmlFor="task-description">Descrição</label>
          <textarea
            id="task-description"
            rows={4}
            value={form.description}
            onChange={update('description')}
            maxLength={MAX_DESCRIPTION}
            placeholder="Contexto, links, critérios de aceite… (opcional)"
            disabled={saving}
          />
          <p className="field__hint">
            {form.description.length}/{MAX_DESCRIPTION}
          </p>
          {errors.description && (
            <p className="field__error" role="alert">
              {errors.description}
            </p>
          )}
        </div>

        <fieldset className="field" disabled={saving}>
          <legend>Coluna</legend>
          <div className="segmented">
            {COLUMNS.map((column) => (
              <label key={column.id} className="segmented__option" style={{ '--accent': column.accent }}>
                <input
                  type="radio"
                  name="status"
                  value={column.id}
                  checked={form.status === column.id}
                  onChange={update('status')}
                />
                <span>{column.title}</span>
              </label>
            ))}
          </div>
        </fieldset>

        {generalError && (
          <p className="dialog__error" role="alert">
            {generalError}
          </p>
        )}

        <footer className="dialog__footer">
          <button type="button" className="button button--ghost" onClick={onClose} disabled={saving}>
            Cancelar
          </button>
          <button type="submit" className="button button--primary" disabled={saving}>
            {saving ? 'Salvando…' : isEditing ? 'Salvar alterações' : 'Criar tarefa'}
          </button>
        </footer>
      </form>
    </dialog>
  )
}
