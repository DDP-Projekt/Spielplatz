<script lang="ts">
    import SvgIcon from "./SvgIcon.svelte";

    const { 
        href,
        title,
        path,
        size = undefined,
        offset = undefined,
        ...rest
    } = $props()
</script>

<a {href} target="_blank" {title} {...rest}>
    <SvgIcon {path} {size} {offset} />
</a>

<style>
    a {
        background: transparent;
        transition: background-color 200ms ease;
        border-radius: .5rem;
        padding: 2px;
        border: 1px solid transparent;
        cursor: pointer;
        color: var(--text-color);

        &:hover {
            background-color: var(--controls-btn-hover-color);
        }
    }
</style>