<script lang="ts">
    import { page } from "$app/state";
    import { onMount, tick } from "svelte";
    import type * as MonacoEditor from "monaco-editor";

    import { mdiClipboardOutline } from "@mdi/js";

    import ControlsHeader from "$lib/components/core/ControlsHeader.svelte";
    import Separator from "$lib/components/core/Separator.svelte";
    import EditorComponent, { type EditorDisplaySettings } from "$lib/components/core/EditorComponent.svelte";
    import OutputComponent from "$lib/components/core/OutputComponent.svelte";

    import ImgButton from "$lib/components/common/ImgButton.svelte";
    import RunButton from "$lib/components/core/RunButton.svelte";
    import SettingsComponent from "$lib/components/core/SettingsComponent.svelte";
    import { withQuery, type OutputMessage } from "$lib";

    let editor: MonacoEditor.editor.IStandaloneCodeEditor | undefined = $state()
    let editorSettings: EditorDisplaySettings | undefined = $state();
    
    let separatorDragging = $state(false)
    let separatorStart = $state(70)

    let scrollLock = $state(false)
    let lightMode = $state(false)

    let args = $state<string[]>([])
    let output = $state<OutputMessage[]>([])

    let run_ws: WebSocket | null = $state(null);
    let outputElement: HTMLDivElement | undefined = $state();

    onMount(async () => {
        const urlParams = page.url.searchParams;
        
        lightMode = localStorage.getItem("dark-mode") === "false" || urlParams.has('light')
        args = localStorage.getItem("args")?.split(";") || []

        let editorContent: string | undefined;
        if (urlParams.has("share")) {
            const resp: {code: string} = await (await fetch(withQuery("/api/get_share_data", { code: urlParams.get("share")! }))).json()
            editorContent = resp.code
        } else {
            editorContent = localStorage.getItem("content") || undefined
        }

        editorSettings = {
            initialContent: editorContent,
            theme: lightMode ? "ddp-theme-light" : "ddp-theme-dark",
            readOnly: urlParams.has("readonly"),
            nolines: urlParams.has("nolines"),
            noscroll: urlParams.has("noscroll"),
            embedded: true
        }
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

<main style:--editor-container-size="{separatorStart}%" data-vertical={true} data-dragging={separatorDragging}>
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
            settings={editorSettings!}
        />
    </div>

    {#if output.length !== 0}
        <Separator
            bind:dragging={separatorDragging} 
            bind:start={separatorStart} 
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
            var(--separator-width) 
            calc(100% - var(--editor-container-size) - var(--separator-width));
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