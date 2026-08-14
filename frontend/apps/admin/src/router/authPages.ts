/**
 * Pages reachable without a session.
 *
 * Shared rather than duplicated because three places need the same answer: the
 * navigation guard, `App.vue` (which hides the app chrome on these pages), and
 * `auth-service` (which must not set a post-login redirect back to a login
 * page). They previously disagreed, which is how an expired session could leave
 * the app rendering its authenticated shell on top of the login route.
 */
export const authPages = ['/login', '/register', '/reset']

export function isAuthPath(path: string): boolean {
  return authPages.includes(path)
}
