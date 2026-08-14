import { useEffect } from 'react'
import { AlertIcon, CheckIcon, CloseIcon } from './Icons'

/**
 * Aviso curto no rodapé. Erros ficam até o usuário fechar; confirmações somem
 * sozinhas depois de três segundos.
 */
export default function Toast({ notice, onDismiss }) {
  const tone = notice?.tone

  useEffect(() => {
    if (!notice || tone === 'error') return
    const timer = setTimeout(onDismiss, 3000)
    return () => clearTimeout(timer)
  }, [notice, tone, onDismiss])

  if (!notice) return null

  return (
    <div className={`toast toast--${tone}`} role="status" aria-live="polite">
      {tone === 'error' ? <AlertIcon /> : <CheckIcon />}
      <span>{notice.message}</span>
      <button type="button" className="icon-button" onClick={onDismiss} aria-label="Fechar aviso">
        <CloseIcon />
      </button>
    </div>
  )
}
