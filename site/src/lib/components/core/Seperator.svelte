<script lang="ts">
let { dragging = $bindable(false), start = $bindable(0), vertical, ondrag = undefined } = $props()

function seperatorMouseDown() {
    dragging = true;
}

function seperatorMouseUp() {
    dragging = false;
}

function seperatorMouseMove(ev: MouseEvent) {
    if (!dragging) return;
    
    if (vertical) {
        let y = ev.clientY / window.innerHeight;
        if (y < 0.1 || y > 0.85) return;
        start = y * 100;
    }
    else {
        let x = ev.clientX / window.innerWidth;
        if (x < 0.3 || x > 0.8) return;
        start = x * 100;
    }

    ondrag?.();
}

function seperatorKeyDown(ev: KeyboardEvent) {
    const step = 2;
    if (ev.key === 'ArrowRight' || ev.key === 'ArrowDown') {
        ev.preventDefault();
        start = Math.min(80, start + step);
        ondrag?.();
    } else if (ev.key === 'ArrowLeft' || ev.key === 'ArrowUp') {
        ev.preventDefault();
        start = Math.max(30, start - step);
        ondrag?.();
    }
}

function seperatorTouchMove(ev: TouchEvent) {
    if (!dragging) return;
    let touch = ev.touches[0];

    if (vertical) {
        let y = touch.clientY / window.innerHeight;
        if (y < 0.1 || y > 0.85) return;
        start = y * 100;
    }
    else {
        let x = touch.clientX / window.innerWidth;
        if (x < 0.3 || x > 0.8) return;
        start = x * 100;
    }
    ondrag?.();
}

</script>

<svelte:window onmousemove={seperatorMouseMove} onmouseup={seperatorMouseUp} />

<div 
    id="seperator" 
    role="slider"
    tabindex="0"
    aria-label="Resize separator"
    aria-valuenow={Math.round(start)}
    aria-valuemin="30"
    aria-valuemax="80"
    onmousedown={seperatorMouseDown}
    ontouchstart={seperatorMouseDown}
    ontouchend={seperatorMouseUp}
    onkeydown={seperatorKeyDown}
    ontouchmove={seperatorTouchMove}
    data-vertical={vertical}
>
</div>

<style>
    #seperator {
        display: flex;
        justify-content: center;
        align-items: center;
        cursor: col-resize;
        background-color: var(--seperator-background-color);

        &:focus-visible {
            outline: 2px solid var(--focus-color, #4A90E2);
            outline-offset: -1px;
        }

        &::before {
            content: "⣿";
        }
    }

    #seperator[data-vertical=true] {
        cursor: row-resize;

        &::before {
            transform: rotate(90deg);
        }
    }
</style>