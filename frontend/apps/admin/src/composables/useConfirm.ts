import { useDialog } from 'naive-ui'
import { useI18n } from 'vue-i18n'

interface ConfirmOptions {
  title: string
  content?: string
  /** Label for the confirming button. Defaults to `common.confirm`. */
  positiveText?: string
  /**
   * 'destructive' (the default) warns and paints the confirm button red.
   * 'neutral' is for reversible steps — publishing, advancing a status — where
   * red would overstate the risk.
   */
  tone?: 'destructive' | 'neutral'
  onConfirm: () => unknown | Promise<unknown>
}

/**
 * Single entry point for "are you sure?" prompts. Every call site that used to
 * reach for `window.confirm` or hand-roll an NModal goes through here, so
 * confirmations share one look and one button wording across the app.
 *
 * Returning the promise from `onConfirm` keeps naive's dialog open with a
 * loading state until the action settles.
 */
export function useConfirm() {
  const dialog = useDialog()
  const { t } = useI18n()

  return function confirm({
    title,
    content,
    positiveText,
    tone = 'destructive',
    onConfirm
  }: ConfirmOptions) {
    const destructive = tone === 'destructive'

    dialog[destructive ? 'warning' : 'info']({
      title,
      content,
      positiveText: positiveText ?? t('common.confirm'),
      negativeText: t('common.cancel'),
      positiveButtonProps: destructive ? { type: 'error' } : undefined,
      onPositiveClick: onConfirm
    })
  }
}
