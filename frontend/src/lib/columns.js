/**
 * As três colunas fixas do quadro.
 *
 * `id` é o valor aceito pelo backend; `accent` é a variável CSS que dá a cor
 * de cada status. Toda a interface (colunas, cartões, barra de progresso e
 * seletor do formulário) deriva desta lista — acrescentar uma quarta coluna
 * seria acrescentar um item aqui e o status correspondente no Go.
 */
export const COLUMNS = [
  {
    id: 'todo',
    title: 'A Fazer',
    accent: 'var(--todo)',
    empty: 'Nada por aqui. Adicione a primeira tarefa.',
  },
  {
    id: 'doing',
    title: 'Em Progresso',
    accent: 'var(--doing)',
    empty: 'Arraste uma tarefa para começar.',
  },
  {
    id: 'done',
    title: 'Concluídas',
    accent: 'var(--done)',
    empty: 'O que for concluído aparece aqui.',
  },
]

export const COLUMN_IDS = COLUMNS.map((column) => column.id)

export const getColumn = (id) => COLUMNS.find((column) => column.id === id) ?? COLUMNS[0]
