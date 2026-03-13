<script lang="ts">
    import { page } from "$app/state";
    import { onMount, tick } from "svelte";
    import type * as MonacoEditor from "monaco-editor";

    import logoImg from "$lib/assets/ddp-logo.svg";
    import { mdiArrowVerticalLock, mdiClipboardOutline, mdiDeleteClockOutline, mdiGithub, mdiHelpCircleOutline, mdiShare, mdiTrashCanOutline, mdiViewSplitHorizontal, mdiViewSplitVertical, mdiWeatherNight, mdiWeatherSunny } from "@mdi/js";

    import ControlsHeader from "$lib/components/core/ControlsHeader.svelte";
    import Separator from "$lib/components/core/Separator.svelte";
    import EditorComponent, { type EditorDisplaySettings } from "$lib/components/core/EditorComponent.svelte";
    import OutputComponent from "$lib/components/core/OutputComponent.svelte";

    import ImgButton from "$lib/components/common/ImgButton.svelte";
    import ImgCheckbox from "$lib/components/common/ImgCheckbox.svelte";
    import ImgLink from "$lib/components/common/ImgLink.svelte";
    import ImgToggleButton from "$lib/components/common/ImgToggleButton.svelte";
    import SettingsComponent from "$lib/components/core/SettingsComponent.svelte";
    import ExampleSelect from "$lib/components/core/ExampleSelect.svelte";
    import RunButton from "$lib/components/core/RunButton.svelte";
    import { withQuery, type OutputMessage } from "$lib";

    let editor: MonacoEditor.editor.IStandaloneCodeEditor | undefined = $state()
    let editorSettings: EditorDisplaySettings | undefined = $state();
    
    let separatorDragging = $state(false)
    let separatorStart = $state(70)

    let autoClear = $state(false)
    let scrollLock = $state(false)
    let lightMode = $state(false)
    let vertical = $state(true)

    let args = $state<string[]>([])
    let output = $state<OutputMessage[]>([])

    let run_ws: WebSocket | null = $state(null);
    let outputElement: HTMLDivElement | undefined = $state();

    onMount(async () => {
        const urlParams = page.url.searchParams;
        
        lightMode = localStorage.getItem("dark-mode") === "false" || urlParams.has('light')
        vertical = localStorage.getItem("vertical") === "true" || window.innerWidth <= 768
        args = localStorage.getItem("args")?.split(";") || []

        editorSettings = {
            initialContent: localStorage.getItem("content") || undefined,
            theme: lightMode ? "ddp-theme-light" : "ddp-theme-dark",
            readOnly: urlParams.has("readonly"),
            nolines: urlParams.has("nolines"),
            noscroll: urlParams.has("noscroll"),
            embedded: false
        }

        if (urlParams.has("share")) {
            const resp: {code: string} = await (await fetch(withQuery("/api/get_share_data", { code: urlParams.get("share")! }))).json()
            editorSettings.initialContent = resp.code
        }
    })

    function onbeforeunload() {
        if (editor) {
            localStorage.setItem("content", editor.getValue());
        }

        if (args.length !== 0) {
            localStorage.setItem("args", args.join(";"));
        } else {
            localStorage.removeItem("args")
        }
    }

    function onThemeChange() {
        localStorage.setItem("dark-mode", !lightMode + "")
        editorSettings!.theme = lightMode ? "ddp-theme-light" : "ddp-theme-dark"
    }

    function onViewChange() {
        localStorage.setItem("vertical", vertical + "")
        editor?.layout()
    }

    async function shareCode() {
        if (!editor) return;

        const shareResp: { share_code: string } = await fetch('/api/create_share_code', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ code: editor.getValue() }),
        }).then(response => response.json())

        if (!shareResp.share_code) {
            alert("Fehler beim Erstellen des Share-Links.")
            return;
        }

        prompt("Share link", `${window.location.origin}/?share=${shareResp.share_code}`)
    }

    async function copyCode() {
        await navigator.clipboard.writeText(editor?.getValue() || "")
    }

    async function copyOutput() {
        await navigator.clipboard.writeText(output.map(x => x.msg).join('\n'))
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

<svelte:window {onbeforeunload} />

<main style:--editor-container-size="{separatorStart}%" data-vertical={vertical} data-dragging={separatorDragging}>
    <div class="container">
        <ControlsHeader>
            {#snippet leftControls()}
                <a href="https://ddp.im/" target="_blank">
                    <img src={logoImg} alt="ddp-logo" width="24px" height="24px" title="DDP Homepage">
                </a>

                <RunButton bind:run_ws {args} {editor} {autoClear} {clearOutput} {pushOutputMessage} />
                <SettingsComponent bind:args />
            {/snippet}
    
            {#snippet rightControls()}
                <ExampleSelect onSelect={(code) => editor?.setValue(code) } />
                <ImgButton onclick={shareCode} path={mdiShare} title="Code teilen" />
                <ImgButton onclick={copyCode} path={mdiClipboardOutline} title="Code kopieren" />
                <ImgLink href="https://github.com/DDP-Projekt/Spielplatz" path={mdiGithub} title="Spielplatz GitHub" />

                <!-- id "theme-switch" affects global style in :root! -->
                <ImgToggleButton id="theme-switch"
                    title="Hell-/Dunkelmodus umschalten"
                    onPath={mdiWeatherNight} offPath={mdiWeatherSunny} 
                    bind:checked={lightMode} onchange={onThemeChange} 
                />

                <ImgLink href="https://doku.ddp.im" path={mdiHelpCircleOutline} title="Bedienungsanleitung öffnen" />
            {/snippet}
        </ControlsHeader>

        <EditorComponent 
            bind:editor
            settings={editorSettings!}
        />
    </div>

    <Separator
        bind:dragging={separatorDragging} 
        bind:start={separatorStart} 
        {vertical}
    />

    <div class="container">
        <ControlsHeader>
            {#snippet leftControls()}
                <ImgButton onclick={clearOutput} path={mdiTrashCanOutline} title="Ausgabe leeren" />
                <ImgButton onclick={copyOutput} path={mdiClipboardOutline} title="Ausgabe kopieren" />
                <ImgCheckbox bind:checked={autoClear} path={mdiDeleteClockOutline} title="Automatisch leeren" offset={{x: 2, y: 0}} />
                <ImgCheckbox bind:checked={scrollLock} path={mdiArrowVerticalLock} title="Scrollen sperren" />
            {/snippet}

            {#snippet rightControls()}
                <div id="view-toggle-container">
                    <ImgToggleButton title="Sicht ändern"
                        bind:checked={vertical} onchange={onViewChange} 
                        onPath={mdiViewSplitVertical} offPath={mdiViewSplitHorizontal} 
                    />
                </div>
            {/snippet}
        </ControlsHeader>

        <OutputComponent bind:outputElement bind:output {run_ws} {pushOutputMessage} >
            <span class="sysmsg">Kompilierer Version: v1.0.0</span>
        </OutputComponent>
    </div>
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

    @media (max-width: 768px) {
        #view-toggle-container {
            display: none;
        }
    }
</style>