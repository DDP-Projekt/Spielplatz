import { PUBLIC_BACKEND_HOST } from "$env/static/public"

// place files you want to import through the `$lib` alias in this folder.
export type OutputMessage = {
    msg: string,
    type: 'stdin' | 'stdout' | 'stderr' | 'sysmsg'
}

export function withQuery(path: string, params: Record<string, string>) {
    return `${path}?${new URLSearchParams(params).toString()}`
}

export function getWebSocketAddr() {
    const ws_protocol = location.protocol === 'https:' ? "wss" : "ws"
    return `${ws_protocol}://${PUBLIC_BACKEND_HOST}/spielplatz`
}