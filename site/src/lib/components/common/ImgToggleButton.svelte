<script lang="ts">
    import SvgIcon from "./SvgIcon.svelte";

    let {
        checked = $bindable(),
        title, 
        onPath,
        offPath,
        size = undefined,
        offset = undefined,
        id = title.toLowerCase().replaceAll(" ", "-"),
        ...rest
    } = $props()
</script>

<div {title} class="img-checkbox">
    <input type="checkbox" {id} name={id} {...rest} bind:checked={checked}>
    <label for={id}>
        {#if checked}
            <SvgIcon path={onPath} {size} {offset} />
        {:else}
            <SvgIcon path={offPath} {size} {offset} />
        {/if}
    </label>
</div>


<style>
    .img-checkbox {
        display: flex;
    }

    input[type="checkbox"] {
        display: none;
    }

    input[type="checkbox"] + label {
        display: block;
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