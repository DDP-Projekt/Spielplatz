<script lang="ts">
    import { page } from "$app/state";
    import { onMount, tick } from "svelte";
    import type * as MonacoEditor from "monaco-editor";

    import { mdiClipboardOutline } from "@mdi/js";

    import ControlsHeader from "$lib/components/core/ControlsHeader.svelte";
    import Seperator from "$lib/components/core/Seperator.svelte";
    import EditorComponent, { type EditorDisplaySettings } from "$lib/components/core/EditorComponent.svelte";
    import OutputComponent from "$lib/components/core/OutputComponent.svelte";

    import ImgButton from "$lib/components/common/ImgButton.svelte";
    import RunButton from "$lib/components/core/RunButton.svelte";
    import SettingsComponent from "$lib/components/core/SettingsComponent.svelte";

    let editor: MonacoEditor.editor.IStandaloneCodeEditor | undefined = $state()
    let editorSettings: EditorDisplaySettings | undefined = $state();
    let editorTheme: "ddp-theme-dark" | "ddp-theme-light" = $state("ddp-theme-dark");
    
    let seperatorDragging = $state(false)
    let seperatorStart = $state(70)

    let scrollLock = $state(false)
    let lightMode = $state(false)

    let args = $state<string[]>([])
    let output = $state<OutputMessage[]>([])

    let run_ws: WebSocket | null = $state(null);
    let outputElement: HTMLDivElement | undefined = $state();

    onMount(() => {
        const urlParams = page.url.searchParams;
        editorSettings = {
            readOnly: urlParams.has("readonly"),
            nolines: urlParams.has("nolines"),
            noscroll: urlParams.has("noscroll"),
            embedded: true
        }

        lightMode = localStorage.getItem("dark-mode") === "false"
        editorTheme = lightMode ? "ddp-theme-light" : "ddp-theme-dark"
    })

    async function copyCode() {
        await navigator.clipboard.writeText(editor?.getValue() || "")
    }

    async function pushOutputMessage(message: OutputMessage) {
        output.push(message)

        if (!scrollLock && outputElement) {
            await tick();
            outputElement.scrollTop = outputElement.scrollHeight;
        }
    }

    function clearOutput() {
        output = []
    }
</script>

<main style:--editor-container-size="{seperatorStart}%" data-vertical={true} data-dragging={seperatorDragging}>
    {#if output.length === 0}
        <style>
            main {
                --grid-template: 1fr !important;
            }
        </style>
    {/if}
    <div class="container">
        <ControlsHeader>
            {#snippet leftControls()}
                <RunButton bind:run_ws {args} {editor} autoClear={true} {clearOutput} {pushOutputMessage} />
                <SettingsComponent bind:args />
            {/snippet}
    
            {#snippet rightControls()}
                <ImgButton onclick={copyCode} path={mdiClipboardOutline} title="Code kopieren" />
            {/snippet}
        </ControlsHeader>

        <EditorComponent 
            bind:editor
            initialContent={null}
            settings={editorSettings!}
            theme={editorTheme}
        />
    </div>

    {#if output.length !== 0}
        <Seperator
            bind:dragging={seperatorDragging} 
            bind:start={seperatorStart} 
            vertical={true}
        />

        <OutputComponent bind:outputElement bind:output {run_ws} {pushOutputMessage} />
    {/if}
</main>

<style>
    main {
        width: 100%;
        height: 100%;

        --grid-template: 
            var(--editor-container-size) 
            var(--seperator-width) 
            calc(100% - var(--editor-container-size) - var(--seperator-width));
        display: grid;
        grid-template-columns: var(--grid-template);
        grid-template-rows: 100%;

        &[data-vertical="true"] {
            grid-template-rows: var(--grid-template);
            grid-template-columns: 100%;
            --drag-cursor: row-resize;
        }

        --drag-cursor: col-resize;
        &[data-dragging="true"] {
            cursor: var(--drag-cursor);
        }

        &[data-dragging="true"] .container {
            user-select: none;
            pointer-events: none;
        }
    }

    .container {
        display: grid;
        grid-template-rows: var(--controls-height) 1fr;
    }
</style>