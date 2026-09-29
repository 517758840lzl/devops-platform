/** SSE 不能走 Vite /api 代理（会缓冲导致收不到事件），开发期直连后端。 */
export function apiBaseForSSE() {
  if (import.meta.env.VITE_API_BASE) {
    return String(import.meta.env.VITE_API_BASE).replace(/\/$/, '')
  }
  const { protocol, hostname } = window.location
  return `${protocol}//${hostname}:8080`
}
