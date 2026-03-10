<script lang="ts">
    import SvgIcon from "./SvgIcon.svelte";

    let {
        checked = $bindable(),
        title, 
        path,
        size = undefined,
        offset = undefined,
        id = title.toLowerCase().replaceAll(" ", "-"),
        ...rest
    } = $props()
</script>

<div {title} class="img-checkbox">
    <input type="checkbox" {id} name={id} {...rest} bind:checked={checked}>
    <label for={id}>
        <SvgIcon {path} {size} {offset} />
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
        color: light-dark(black, white);

        &:hover {
            background-color: var(--controls-btn-hover-color);
        }
    }

    input[type="checkbox"]:checked + label {
        background-color: var(--controls-btn-hover-color);
        border: 1px solid var(--controls-btn-active-color);
    }
</style>