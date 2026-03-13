<script lang="ts">
    import type { OutputMessage } from "$lib";

    type OutputComponentProps = {
        children?: any,
        outputElement: HTMLDivElement | undefined
        output: OutputMessage[],
        run_ws: WebSocket | null,
        pushOutputMessage: (m: OutputMessage) => Promise<void>
    }
    
    let {
        children, outputElement = $bindable(), output = $bindable([]), run_ws, pushOutputMessage
    }: OutputComponentProps = $props()

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
        await pushOutputMessage({msg, type: 'stdin'})
        if (run_ws) {
            run_ws.send(JSON.stringify({ msg: msg, eof: eof }));
        }
        input.value = "";
    }
</script>

<div id="output-container">
    <div id="output" bind:this={outputElement}>
        <pre id="outputText">
            {@render children?.()}
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
    :global {
        #outputText span.stdin { color: var(--stdin-color); }
        #outputText span.stderr { color: var(--stderr-color); }
        #outputText span.sysmsg { color: var(--sysmsg-color); }
    }

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