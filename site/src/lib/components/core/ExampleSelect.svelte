<script lang="ts">
    type OnSelectCallback = (code: string) => void;
    let { onSelect } : { onSelect: OnSelectCallback } = $props()

    function loadExample(event: Event) {
        const val = (event.target as HTMLSelectElement)?.value;
        if (!val || val === "") return;
        if (val === "HalloWelt") {
            onSelect('Binde "Duden/Ausgabe" ein.\nSchreibe "Hallo Welt!".');
            return;
        }

        fetch(`https://raw.githubusercontent.com/DDP-Projekt/Kompilierer/master/examples/${val}.ddp`)
            .then(r => r.text())
            .then(t => onSelect(t))
    }
</script>

<select id="example-select" onchange={loadExample}>
    <option hidden value="">Beispiele</option>
    <option value="HalloWelt">Hallo Welt</option>
    <option value="Brainfuck">Brainfuck</option>
    <option value="Fibonacci">Fibonacci</option>
    <option value="Fizzbuzz">Fizzbuzz</option>
    <option value="Mandelbrot">Mandelbrot</option>
    <option value="Roemische_Zahlen">Roemische_Zahlen</option>
    <option value="Tictactoe">Tictactoe</option>
    <option value="reverse_string">reverse_string</option>
</select>

<style>
    select {
        appearance: none;
        width: 100%;
        max-width: 100px;
        min-width: 20px;
        text-overflow: ellipsis;

        background-color: var(--input-background-color);
        color: var(--text-color);
        border-radius: 0.5rem;
        border: var(--input-border-color) solid var(--input-border-width);
        margin: 0 1rem;
        padding: 2px 5px;
    }
</style>