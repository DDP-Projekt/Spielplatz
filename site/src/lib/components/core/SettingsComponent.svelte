<script lang="ts">
    import { mdiCog, mdiMinusBoxOutline, mdiPlusBoxOutline } from "@mdi/js";
    import ImgButton from "../common/ImgButton.svelte";

    let { args = $bindable<string[]>([]) } = $props()

    function addArg() {
        args.push("")
    }

    function removeArg(index: number) {
        args.splice(index, 1)
    }
</script>

<ImgButton title="Einstellungen" path={mdiCog} popovertarget="settings-panel" id="settings-button" />
<div id="settings-panel" popover>
    <div id="arg-header">
        <p>Befehlszeilenargumente</p>
        <ImgButton title="Argument hinzufügen" path={mdiPlusBoxOutline} size={"20px"} onclick={addArg} />
    </div>
    <div id="arg-list">
        <div class="arg">
            <label for="arg-1">1:</label>
            <input type="text" id="arg-1" disabled value="<Programmname>">
            <div style="width: 26px;"></div>
        </div>
        {#each args as _, i}
            <div class="arg">
                <label for="arg-{i+2}">{i+2}:</label>
                <input type="text" id="arg-{i+2}" autocomplete="off" bind:value={args[i]}>
                <ImgButton title="Argument entfernen" path={mdiMinusBoxOutline} size={"20px"} onclick={() => removeArg(i)} />
            </div>
        {/each}
    </div>
</div>

<style>
    :global(#settings-button) {
        anchor-name: --settings-anchor;
    }

    #settings-panel {
        position: fixed;
        position-anchor: --settings-anchor;
        top: var(--controls-height);
        left: anchor(left);
        
        background-color: var(--controls-background-color);
        padding: 0.75rem;
        border: 1px solid var(--input-background-color);
        border-top-color: var(--controls-background-color);
        border-radius: 0 0 1rem 1rem;
    }

    #settings-panel:popover-open {
        display: flex;
        flex-direction: column;
        width: 18rem;
    }

    #arg-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    #arg-list {
        display: flex;
        flex-direction: column;
        row-gap: 0.25rem;
    }

    .arg {
        display: grid;
        grid-template-columns: auto 1fr auto;
        align-items: center;
        gap: 0.5rem;
    }

    .arg > input {
        width: 100%;
        background-color: var(--input-background-color);
        color: var(--text-color);
        border-radius: 0.25rem;
        border: var(--input-border-color) solid var(--input-border-width);

        &:disabled {
            color: rgb(from var(--text-color) r g b / 0.6);
            border-color: rgb(from var(--input-border-color) r g b / 0.6);
        }
    }
</style>