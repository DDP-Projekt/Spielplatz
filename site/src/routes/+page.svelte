<script lang="ts">
    import { page } from "$app/state";
    import { onMount } from "svelte";
    import type * as MonacoEditor from "monaco-editor";

    import logoImg from "$lib/assets/ddp-logo.svg";
    import { mdiArrowVerticalLock, mdiClipboardOutline, mdiCloseOctagonOutline, mdiDeleteClockOutline, mdiGithub, mdiHelpCircleOutline, mdiPlayOutline, mdiShare, mdiTrashCanOutline, mdiViewSplitHorizontal, mdiViewSplitVertical, mdiWeatherNight, mdiWeatherSunny } from "@mdi/js";

    import ControlsHeader from "$lib/components/core/ControlsHeader.svelte";
    import Seperator from "$lib/components/core/Seperator.svelte";
    import EditorComponent, { type EditorDisplaySettings } from "$lib/components/core/EditorComponent.svelte";
    import OutputComponent from "$lib/components/core/OutputComponent.svelte";

    import ImgButton from "$lib/components/common/ImgButton.svelte";
    import ImgCheckbox from "$lib/components/common/ImgCheckbox.svelte";
    import ImgLink from "$lib/components/common/ImgLink.svelte";
    import ImgToggleButton from "$lib/components/common/ImgToggleButton.svelte";
    import SettingsComponent from "$lib/components/core/SettingsComponent.svelte";
    import ExampleSelect from "$lib/components/core/ExampleSelect.svelte";

    let editor: MonacoEditor.editor.IStandaloneCodeEditor | undefined = $state()
    let editorSettings: EditorDisplaySettings | undefined = $state();
    let editorTheme: "ddp-theme-dark" | "ddp-theme-light" = $state("ddp-theme-dark");
    
    let seperatorDragging = $state(false)
    let seperatorStart = $state(70)

    let autoClear = $state(false)
    let scrollLock = $state(false)
    let lightMode = $state(false)
    let vertical = $state(false)

    let args = $state<string[]>([])
    let output = $state<OutputMessage[]>([])
    let outputComponent: { pushOutput: (message: OutputMessage) => Promise<void> } | undefined = $state();

    async function pushOutputMessage(msg: string, target: OutputMessage["type"]) {
        await outputComponent?.pushOutput({ msg, type: target });
    }

    function toggleBtnVisibility() {
        // Kept as a compatibility shim until run/stop buttons are wired.
    }

    onMount(() => {
        const urlParams = page.url.searchParams;
        editorSettings = {
            readOnly: urlParams.has("readonly"),
            nolines: urlParams.has("nolines"),
            noscroll: urlParams.has("noscroll"),
            embedded: false
        }

        lightMode = localStorage.getItem("dark-mode") === "false"
        vertical = localStorage.getItem("vertical") === "true"
        editorTheme = lightMode ? "ddp-theme-light" : "ddp-theme-dark"
    })

    function onThemeChange() {
        localStorage.setItem("dark-mode", !lightMode + "")
        editorTheme = lightMode ? "ddp-theme-light" : "ddp-theme-dark"
    }

    function onViewChange() {
        localStorage.setItem("vertical", vertical + "")
        editor?.layout()
    }

    async function copyCode() {
        await navigator.clipboard.writeText(editor?.getValue() || "")
    }

    async function copyOutput() {
        await navigator.clipboard.writeText(output.map(x => x.msg).join('\n'))
    }

    function clearOutput() {
        output = []
    }


    let run_ws: WebSocket | null = $state(null);
    let compiling = $state(false);

    async function runProgram() {
        const code = editor?.getValue()

        if (run_ws) {
            pushOutputMessage('Programm läuft bereits.', 'stderr');
            return;
        } else if (compiling) {
            pushOutputMessage('Programm wird gerade kompiliert.', 'stderr');
            return;
        }

        if (autoClear) {
            clearOutput();
        }

        toggleBtnVisibility();

        compiling = true;
        // send a request to the /compile endpoint using the fetch api
        const compile_result = await fetch('compile', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ src: code }),
        }).then(response => response.json())

        if (compile_result.error) {
            pushOutputMessage("Kompilier Fehler: " + compile_result.error, 'stderr');
            pushOutputMessage(" ", 'stderr');
            pushOutputMessage(compile_result.stderr, 'stderr');

            compiling = false;
            toggleBtnVisibility();
            return
        }

        const token = compile_result.token
        let argsString = ""
        for (let arg of args) {
            argsString += "&args=" + arg;
        }

        // connect to the /run endpoint using the websocket api with token as query parameter
        let ws_protocol = location.protocol === 'https:' ? "wss": "ws"
        run_ws = new WebSocket(`${ws_protocol}://${window.location.host}/run?token=${token}${argsString}`)
        if (!run_ws){
            console.error('websocket (run) connection failed')
            return;
        } 
        
        // focus input
        //document.getElementById('input').focus();

        run_ws.onopen = () => {
            //console.log('websocket (run) connection opened')
        }

        run_ws.onmessage = (event) => {
            let msg = JSON.parse(event.data)
            pushOutputMessage(msg.msg, msg.isStderr ? 'stderr' : 'stdout');
        }

        run_ws.onclose = (event) => {
            //console.log('websocket (run) connection closed: ', event)
            pushOutputMessage(" ", 'sysmsg')
            pushOutputMessage(event.reason, event.code !== 1000 ? 'stderr' : 'sysmsg')
            run_ws = null;
            compiling = false;

            toggleBtnVisibility();
        }

        run_ws.onerror = () => {
            console.error('websocket (run) error')
            pushOutputMessage('websocket (run) error', 'stderr');

            toggleBtnVisibility();
        }
    }

    function stopProgram() {
        if (!run_ws) {
            return;
        }

        pushOutputMessage("Das Programm wurde abgebrochen", 'sysmsg');
        run_ws.send(JSON.stringify({ msg: "EOF", eof: true }));
        run_ws = null;
        compiling = false;
    }
</script>

<main style:--editor-container-size="{seperatorStart}%" data-vertical={vertical} data-dragging={seperatorDragging}>
    <div class="container">
        <ControlsHeader>
            {#snippet leftControls()}
                <a href="https://ddp.im/" target="_blank">
                    <img src={logoImg} alt="ddp-logo" width="24px" height="24px" title="DDP Homepage">
                </a>
                <ImgButton title="Ausführen" onclick={runProgram} path={mdiPlayOutline} offset={{x: -2, y: 0}} />
                <ImgButton title="Stop" onclick={stopProgram} path={mdiCloseOctagonOutline} />
                <SettingsComponent bind:args={args} />
            {/snippet}
    
            {#snippet rightControls()}
                <ExampleSelect onSelect={(code) => editor?.setValue(code) } />
                <ImgButton onclick={""} path={mdiShare} title="Code teilen" />
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
            bind:editor={editor}
            initialContent={null}
            settings={editorSettings!}
            theme={editorTheme}
        />
    </div>

    <Seperator
        bind:dragging={seperatorDragging} 
        bind:start={seperatorStart} 
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
                <ImgToggleButton title="Sicht ändern"
                    bind:checked={vertical} onchange={onViewChange} 
                    onPath={mdiViewSplitVertical} offPath={mdiViewSplitHorizontal} 
                />
            {/snippet}
        </ControlsHeader>

        <OutputComponent bind:this={outputComponent} {output} {scrollLock} {run_ws} />
    </div>
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