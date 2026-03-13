<script lang="ts">
    type SvgIconProps = {
        size?: number,
        offset?: {x: number, y: number},
        path: string,
        viewBox?: string
    }

    const { 
        size = 24,
        offset = {x: 0, y: 0},
        path,
        viewBox = `${offset.x} ${offset.y} 24 24`
    } : SvgIconProps = $props()
</script>

<svg width={size} height={size} {viewBox}>
    <path d={path} />
</svg>

<style>
    path {
        fill: currentColor;
    }
</style>