// place files you want to import through the `$lib` alias in this folder.
type OutputMessage = {
    msg: string,
    type: 'stdin' | 'stdout' | 'stderr' | 'sysmsg'
}