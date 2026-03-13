<script lang="ts">
    import SvgIcon from "./SvgIcon.svelte";

    const { 
        onclick = undefined,
        title,
        path,
        size = undefined,
        offset = undefined,
        ...rest
    } = $props()
</script>

<button {title} {onclick} {...rest}>
    <SvgIcon {path} {size} {offset} />
</button>

<style>
    button {
        appearance: none;
        background: transparent;
        transition: background-color 200ms ease;
        border-radius: .5rem;
        padding: 2px;
        border: 1px solid transparent;
        cursor: pointer;
        color: light-dark(black, white);

        &:hover {
            background-color: var(--controls-btn-hover-color);
        }
    }
</style>