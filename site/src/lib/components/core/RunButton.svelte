<script lang="ts">
    import { mdiCloseOctagonOutline, mdiPlayOutline } from "@mdi/js";
    import ImgButton from "../common/ImgButton.svelte";
    import type { OutputMessage } from "$lib";

    type RunButtonProps = {
        run_ws: WebSocket | null,
        args: string[],
        editor: any,
        autoClear: boolean,
        clearOutput: () => void,
        pushOutputMessage: (m: OutputMessage) => Promise<void>
    }

    let {
        run_ws = $bindable(), args, editor, autoClear, clearOutput, pushOutputMessage
    } : RunButtonProps = $props()

    let compiling = $state(false);

    async function runProgram() {
        const code = editor?.getValue()

        if (run_ws) {
            await pushOutputMessage({msg: 'Das Programm läuft bereits.', type: 'stderr'});
            return;
        } else if (compiling) {
            await pushOutputMessage({msg: 'Das Programm wird gerade kompiliert.', type: 'stderr'});
            return;
        }

        if (autoClear) {
            clearOutput();
        }

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
            await pushOutputMessage({msg: `Kompilier Fehler: ${compile_result.error}\n${compile_result.stderr}`, type: 'stderr'});
            compiling = false;
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
        if (!run_ws) {
            await pushOutputMessage({msg: 'Websocket (run) Verbindungsfehler', type: "stderr"})
            return;
        } 
        
        // focus input
        //document.getElementById('input').focus();

        run_ws.onopen = (ev) => {
            //console.log('websocket (run) connection opened')
        }

        run_ws.onmessage = async (event) => {
            let msg = JSON.parse(event.data)
            await pushOutputMessage({msg: msg.msg, type: msg.isStderr ? 'stderr' : 'stdout'});
        }

        run_ws.onclose = async (event) => {
            //console.log('websocket (run) connection closed: ', event)
            await pushOutputMessage({msg: event.reason, type: event.code !== 1000 ? 'stderr' : 'sysmsg'})
            run_ws = null;
            compiling = false;
        }

        run_ws.onerror = async () => {
            console.error('websocket (run) error')
            await pushOutputMessage({msg: 'Websocket (run) Fehler', type: 'stderr'});
        }
    }

    async function stopProgram() {
        if (!run_ws) {
            return;
        }

        await pushOutputMessage({msg: "Das Programm wurde abgebrochen", type: 'sysmsg'});
        run_ws.send(JSON.stringify({ msg: "EOF", eof: true }));
        run_ws = null;
        compiling = false;
    }
</script>

{#if !run_ws}
    <ImgButton title="Ausführen" onclick={runProgram} path={mdiPlayOutline} offset={{x: -2, y: 0}} />
{:else}
    <ImgButton title="Stop" onclick={stopProgram} path={mdiCloseOctagonOutline} />
{/if}