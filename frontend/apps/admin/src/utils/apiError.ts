import axios from 'axios'

/**
 * Extracts the human-readable message from a failed API call.
 *
 * The Go API serialises every error through a single `RespondWithError` helper
 * as `{"msg": string, "code": number}` (api/internal/handlers/helpers.go), so
 * `msg` is the only field ever populated. Call sites used to reach for
 * `data.message` or `data.error` as well, which are always undefined and made
 * the server's explanation fall through to axios's generic
 * "Request failed with status code N".
 */
export function apiErrorMessage(err: unknown, fallback: string): string {
  if (axios.isAxiosError<{ msg?: string }>(err)) {
    return err.response?.data?.msg || err.message || fallback
  }
  if (err instanceof Error) {
    return err.message || fallback
  }
  return fallback
}
