interface JwtPayload {
  exp?: number
}

/**
 * Reads a JWT payload without verifying its signature.
 *
 * The server stays the only authority on whether a token is valid. This exists
 * so the client can tell a live token from a dead one instead of treating any
 * non-empty string as proof of being signed in.
 */
function decodePayload(token: string): JwtPayload | null {
  const segment = token.split('.')[1]
  if (!segment) {
    return null
  }
  try {
    // JWTs are base64url, which atob does not accept: restore the standard
    // alphabet and the padding that base64url strips.
    const base64 = segment.replace(/-/g, '+').replace(/_/g, '/')
    const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=')
    return JSON.parse(atob(padded))
  } catch {
    return null
  }
}

/** Expiry as epoch milliseconds, or null if the token has no readable `exp`. */
export function getTokenExpiry(token: string | null): number | null {
  if (!token) {
    return null
  }
  const exp = decodePayload(token)?.exp
  return typeof exp === 'number' ? exp * 1000 : null
}

/**
 * A token with no readable `exp` counts as expired: the API would reject it
 * anyway, and the alternative is treating a malformed token as valid forever,
 * which is the bug this module exists to prevent.
 */
export function isTokenExpired(token: string | null, skewMs = 0): boolean {
  const expiresAt = getTokenExpiry(token)
  if (expiresAt === null) {
    return true
  }
  return Date.now() >= expiresAt - skewMs
}
