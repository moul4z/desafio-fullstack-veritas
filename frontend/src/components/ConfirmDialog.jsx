import { useEffect, useRef } from 'react'

/**
 * Confirmação antes de uma ação destrutiva. Também usa <dialog> nativo pelos
 * mesmos motivos do formulário: foco preso, Esc e backdrop sem código extra.
 */
export default function ConfirmDialog({ open, title, message, confirmLabel = 'Excluir', onConfirm, onClose }) {
  const dialogRef = useRef(null)
  const confirmRef = useRef(null)

  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) return

    if (open && !dialog.open) {
      dialog.showModal()
      // O foco começa em "Cancelar" (autofocus abaixo) para que Enter no
      // impulso não apague nada sem querer.
      requestAnimationFrame(() => confirmRef.current?.focus())
    } else if (!open && dialog.open) {
      dialog.close()
    }
  }, [open])

  return (
    <dialog
      ref={dialogRef}
      className="dialog"
      onCancel={(event) => {
        event.preventDefault()
        onClose()
      }}
      onClick={(event) => {
        if (event.target === dialogRef.current) onClose()
      }}
      aria-labelledby="confirm-title"
    >
      <div className="dialog__panel dialog__panel--compact">
        <p className="eyebrow">Confirmação</p>
        <h2 id="confirm-title" className="dialog__title">
          {title}
        </h2>
        <p className="dialog__message">{message}</p>

        <footer className="dialog__footer">
          <button type="button" className="button button--ghost" ref={confirmRef} onClick={onClose}>
            Cancelar
          </button>
          <button
            type="button"
            className="button button--danger"
            onClick={() => {
              onConfirm()
              onClose()
            }}
          >
            {confirmLabel}
          </button>
        </footer>
      </div>
    </dialog>
  )
}
