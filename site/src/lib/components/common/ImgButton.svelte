<script module>
    let counter = 0; /* counter to generate unique css anchors */
</script>

<script lang="ts">
    import SvgIcon from "./SvgIcon.svelte";

    const {
        onclick = undefined,
        toast = undefined,
        title,
        path,
        size = undefined,
        offset = undefined,
        ...rest
    } = $props()

    const anchorName = $derived(toast ? `--btn-anchor-${counter++}` : undefined);

    let toastEl: HTMLElement | undefined = $state();
    let toastTimeout: NodeJS.Timeout | undefined;

    function onBtnClick(a: any) {
        onclick?.(a)

        if (toast) {
            clearTimeout(toastTimeout);
            toastEl?.showPopover();
            toastTimeout = setTimeout(() => toastEl?.hidePopover(), 2000);
        }
    }
</script>

<button {title} onclick={onBtnClick} style:anchor-name={anchorName} {...rest}>
    <SvgIcon {path} {size} {offset} />
</button>

{#if toast}
    <div bind:this={toastEl} 
        popover="manual" 
        class="toast" 
        style="position-anchor: {anchorName}"
    >
        {toast}
    </div>
{/if}

<style>
    button {
        appearance: none;
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

    .toast {
        position: absolute;
        top: calc(anchor(bottom) + 0.4rem);
        position-area: bottom;

        background: #333;
        padding: 0.33rem 1rem;
        border-radius: 0.5rem;
        border: 1px solid #555;
        font-size: 0.85rem;
        pointer-events: none;

        opacity: 0; /* for fade out */
        transition: 
            display 0.5s allow-discrete,
            overlay 0.5s allow-discrete, /* brower support bad */
            opacity 0.5s;
    }

    .toast:popover-open {
        display: grid;
        opacity: 1;

        /* for fade in */
        @starting-style {
            opacity: 0;
        }
    }
</style>