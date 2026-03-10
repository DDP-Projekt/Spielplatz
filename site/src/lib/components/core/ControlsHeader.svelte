<script lang="ts">
    const { leftControls, rightControls } = $props()
</script>

<header>
    <div>
        {@render leftControls()}
    </div>
    <div>
        {@render rightControls()}
    </div>
</header>

<style>
    header {
        width: 100%;
        height: var(--controls-height);
        padding: 0 0.5rem;
        background-color: var(--controls-background-color);
        display: flex;
        flex-direction: row;
        justify-content: space-between;
        gap: 0.5rem;
    }

    header > div {
        display: flex;
        align-items: center;
        gap: 0.25rem;
    }
</style>