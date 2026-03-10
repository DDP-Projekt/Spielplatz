<script lang="ts">
    import { tick } from "svelte";

    let { output, scrollLock, run_ws } = $props()

    let outputElement: HTMLDivElement | undefined;
    export async function pushOutput(message: OutputMessage) {
        output.push(message)

        if (!scrollLock && outputElement) {
            await tick();
            outputElement.scrollTop = outputElement.scrollHeight;
        }
    }

    async function inputEnter(ev: KeyboardEvent) {
        if (!ev.target) return;

        let eof = false;
        const input = ev.target as HTMLInputElement
        let msg = input.value + '\n';

        if (ev.ctrlKey && ev.key === 'c') {
            eof = true
            msg = "EOF";
        } else if (ev.key !== "Enter") {
            return;
        }
        
        ev.preventDefault();
        await pushOutput({msg, type: 'stdin'})
        if (run_ws) {
            run_ws.send(JSON.stringify({ msg: msg, eof: eof }));
        }
        input.value = "";
    }
</script>

<div id="output-container">
    <div id="output" bind:this={outputElement}>
        <pre id="outputText">
            <!-- mustaches get filled server side -->
            <span class="sysmsg">Kompilierer Version: {"{{ . }}"}</span>
            {#each output as msg}
                <span class={msg.type}>{msg.msg}</span>
            {/each}
        </pre>
    </div>
    
    <form id="input-container">
        <label for="input">&gt;</label>
        <input id="input" type="text" autocomplete="off" onkeydown={inputEnter}>
    </form>
</div>

<style>
    #output-container {
        display: grid;
        grid-template-rows: 1fr 2rem;
        background-color: var(--output-background-color);
        min-height: 0;
    }

    #output {
        font-family: monospace;
        overflow: auto;
        min-height: 0;
    }

    #outputText {
        display: flex;
        flex-direction: column;
        padding-left: calc(1ch + 0.25rem + 0.5rem);
        margin: 0;
    }

    #outputText span { display: block; }
    span.stdin { color: var(--stdin-color); }
    span.stderr { color: var(--stderr-color); }
    span.sysmsg { color: var(--sysmsg-color); }

    #input-container {
        display: flex;
        align-items: center;
        padding: 0.25rem;
        gap: 0.5rem;
        font-family: monospace;
    }

    #input-container input {
        flex: 1;
        height: var(--input-height);
        padding: 0;
    }
</style>