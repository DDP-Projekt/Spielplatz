<script lang="ts">
    import { browser } from "$app/environment";
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
    import { getInitialContent, type OutputMessage } from "$lib";

    const initLightMode = browser ? document.documentElement.dataset.theme === "light" : false

    let autoClear = $state(false)
    let scrollLock = $state(false)
    let vertical = $state(true)

    let separatorDragging = $state(false)
    let separatorStart = $state(70)

    let editor: MonacoEditor.editor.IStandaloneCodeEditor | undefined = $state()
    let editorSettings: EditorDisplaySettings = $state(getEditorSettings());
    let args = $state<string[]>([])
    let output = $state<OutputMessage[]>([])

    let run_ws: WebSocket | null = $state(null);
    let outputElement: HTMLDivElement | undefined = $state();

    onMount(() => {        
        vertical = localStorage.getItem("vertical") !== "false" || window.innerWidth <= 768
        args = localStorage.getItem("args")?.split(";") || []
    })

    function getEditorSettings() {
        const urlParams = page.url.searchParams;

        const settings: EditorDisplaySettings = {
            initialContent: getInitialContent(urlParams),
            theme: initLightMode ? "ddp-theme-light" : "ddp-theme-dark",
            readOnly: urlParams.has("readonly"),
            nolines: urlParams.has("nolines"),
            noscroll: urlParams.has("noscroll"),
            embedded: false
        }
        return settings
    }

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

    function onThemeChange(ev: Event) {
        const checked = (ev.target as HTMLInputElement)?.checked
        localStorage.setItem("dark-mode", !checked + "")
        document.documentElement.dataset.theme = checked ? "light" : "dark"
        editorSettings.theme = checked ? "ddp-theme-light" : "ddp-theme-dark"
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

    async function getDDPVersion() {
        type HealthCheckResponse = {
            healty: boolean,
            'kddp-status': {
                healthy: boolean,
                version: string,
                "exit-status": number
            }
        }

        const health: HealthCheckResponse = await fetch("/api/health").then(x => x.json())
        return "Kompilierer Version: " + health["kddp-status"].version.trim()
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
                    checked={initLightMode} onchange={onThemeChange} 
                />

                <ImgLink href="https://doku.ddp.im" path={mdiHelpCircleOutline} title="Bedienungsanleitung öffnen" />
            {/snippet}
        </ControlsHeader>

        <EditorComponent 
            bind:editor
            settings={editorSettings}
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
            {#await getDDPVersion() then version}
                <span class="sysmsg">{version}</span>
            {:catch}
                <span class="stderr">Kompilierer nicht verbunden!</span>
            {/await}
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