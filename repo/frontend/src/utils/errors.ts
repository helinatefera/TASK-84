/**
 * Extract the user-facing error message from an API error response.
 *
 * The canonical backend envelope is: { "code": <int|string>, "msg": "<message>" }
 * This helper handles both "msg" and legacy "message" fields for resilience.
 */
export function getErrorMessage(error: any, fallback = 'An error occurred'): string {
  const data = error?.response?.data
  if (!data) return fallback
  return data.msg || data.message || fallback
}
