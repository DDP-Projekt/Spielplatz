// place files you want to import through the `$lib` alias in this folder.
export type OutputMessage = {
    msg: string,
    type: 'stdin' | 'stdout' | 'stderr' | 'sysmsg'
}

export function withQuery(path: string, params: Record<string, string>) {
    return `${path}?${new URLSearchParams(params).toString()}`
}