"use strict";

document.addEventListener("DOMContentLoaded", () => {
	const splitBtn = document.getElementById("split-btn");
	const main = document.getElementById("main");

	if (splitBtn && window.localStorage.getItem("vertical") === "false") {
		setHorizontal();

		window.localStorage.setItem("vertical", "false");
	}

	if (splitBtn && window.innerWidth <= 600) {
		splitBtn.setAttribute('hidden', 'true')
		setHorizontal()
	}

	const exampleSelect = document.getElementById("example-select")
	if (exampleSelect && window.localStorage.getItem("content") === null) {
		exampleSelect.value = "HalloWelt";
	}

	const argsContainer = document.getElementById("args");
	const args = window.localStorage.getItem("args");
	if (args !== null) {
		argsContainer.value = args;
	}

	main.style.visibility = "";
});

window.addEventListener("resize", (event) => {
	const splitBtn = document.getElementById("split-btn");
	if (!splitBtn) return;

	if (event.target.innerWidth <= 600) {
		setHorizontal();
		splitBtn.setAttribute('hidden', 'true')
	}
	else {
		splitBtn.removeAttribute('hidden');
	}
});

let saveCount = 0;
document.addEventListener("keydown", function (e) {
	if (e.key == "s" && (navigator.platform.match("Mac") ? e.metaKey : e.ctrlKey)) {
		e.preventDefault();
		saveCount++;
		if (saveCount > 1) {
			alert('NOCHMAL \uD83D\uDE2D\uD83D\uDE2D\uD83D\uDE2D\uD83D\uDE2D')
			return;
		}
		alert('BRUDER HAT VERSUCHT ZU SPEICHERN \uD83D\uDC80\uD83D\uDC80\uD83D\uDC80')
	}
}, false);

function setHorizontal() {
	const splitBtn = document.getElementById("split-btn");
	const main = document.getElementById("main");
	const editorContainer = document.getElementById("editor-container");
	const outputContainer = document.getElementById("output-container");
	const spacer = document.getElementById("spacer");

	main.setAttribute('horizontal', 'true');
	editorContainer.setAttribute('horizontal', 'true');
	outputContainer.setAttribute('horizontal', 'true');
	splitBtn.setAttribute('src', 'static/img/view-split-vertical.svg');
	spacer.setAttribute('horizontal', 'true');
}

function toggleSettings() {
	const panel = document.getElementById('settings-panel')
	panel.style.left = editor.getLayoutInfo().contentLeft + "px";
	panel.toggleAttribute('hidden');
}

function clearOutput() {
	document.getElementById('outputText').innerHTML = '';
}

let autoClear = false;
function toggleAutoClear() {
	autoClear = !autoClear;
	document.getElementById('auto-clear-btn').toggleAttribute('active');
}

function split() {
	const splitBtn = document.getElementById("split-btn");
	const main = document.getElementById("main");
	const editorContainer = document.getElementById("editor-container");
	const outputContainer = document.getElementById("output-container");

	main.toggleAttribute('horizontal');
	editorContainer.toggleAttribute('horizontal');
	outputContainer.toggleAttribute('horizontal');
	document.getElementById("spacer").toggleAttribute('horizontal');

	if (window.localStorage.getItem("vertical") === "true") {
		splitBtn.setAttribute('src', 'static/img/view-split-vertical.svg');
		window.localStorage.setItem("vertical", "false");
	}
	else {
		splitBtn.setAttribute('src', 'static/img/view-split-horizontal.svg');
		window.localStorage.setItem("vertical", "true");
	}
}

function loadExample(val) {
	if (val === "") return;
	if (val === "HalloWelt") {
		editor.setValue('Binde "Duden/Ausgabe" ein.\nSchreibe "Hallo Welt!".');
		return;
	}

	fetch(`https://raw.githubusercontent.com/DDP-Projekt/Kompilierer/master/examples/${val}.ddp`)
		.then(r => r.text())
		.then(t => editor.setValue(t))
}

const root = document.documentElement;

function spacerDragStart(ev) {
	if (ev instanceof DragEvent) {
		ev.dataTransfer.setDragImage(document.createElement('br'), 0, 0);
	}

	root.style.setProperty('--none-if-dragging', 'none');
}

function spacerDragEnd() {
	root.style.setProperty('--none-if-dragging', '');
}

function spacerDrag(ev) {
	// for mobile touch support
	let ev2 = ev
	if (window.TouchEvent && ev instanceof TouchEvent) {
    	ev2 = ev.touches[0]
    }

	const main = document.getElementById('main');

	if (main.hasAttribute('horizontal')) {
		let y = ev2.clientY / ev.view.innerHeight;
		if (y < 0.1 || y > 0.85) return;
		root.style.setProperty('--editor-container-size', y * 100 + '%');
	}
	else {
		let x = ev2.clientX / ev.view.innerWidth;
		if (x < 0.3 || x > 0.8) return;
		root.style.setProperty('--editor-container-size', x * 100 + '%');
	}
}

function copyOutput() {
	navigator.clipboard.writeText(document.getElementById('output').innerText);
}

function copyCode() {
	navigator.clipboard.writeText(editor.getValue());
}

function share() {
	const lzCode = LZUTF8.compress(editor.getValue(), {outputEncoding: "Base64"});

	const newUrl = `${window.location.origin}${window.location.pathname}?code=${encodeURIComponent(lzCode)}`;
	navigator.clipboard.writeText(newUrl);
	alert("Link in die Zwischenablage kopiert.")
}

function toggleDarkMode() {
	document.querySelector('html').toggleAttribute('light')

	if (document.querySelector('html').hasAttribute('light')) {
		monaco.editor.setTheme("ddp-theme-light");
		window.localStorage.setItem("dark-mode", "false");
	}
	else {
		monaco.editor.setTheme("ddp-theme-dark");
		window.localStorage.setItem("dark-mode", "true");
	}
}

let lockScroll = false;
function toggleLockScroll() {
	lockScroll = !lockScroll;
	document.getElementById('lock-scroll-btn').toggleAttribute('active');
}

const MessageTarget = {
	input: "stdin",
	output: "stdout",
	error: "stderr",
	system: "sysmsg"
}

function pushOutputMessage(message, target) {
	const outputText = document.getElementById('outputText');
	const span = document.createElement('span');
	span.classList.add(target);
	span.innerHTML = message;
	outputText.appendChild(span);

	if (!lockScroll) {
		updateOutputScrollbar();
	}
}

function updateOutputScrollbar() {
	var element = document.getElementById("output");
	element.scrollTop = element.scrollHeight;
}

function inputEnter(ev) {
	let eof = false;
	let msg = ev.target.value + '\n';

	if (ev.ctrlKey && ev.key === 'c') {
		eof = true
		msg = "EOF";
	} else if (ev.key !== "Enter") {
		return;
	}
	
	ev.preventDefault();
	pushOutputMessage(msg, MessageTarget.input);
	if (run_ws) {
		run_ws.send(JSON.stringify({ msg: msg, eof: eof }));
	}
	ev.target.value = "";
}

function showOutput() {
	document.getElementById('output-container').style.display = 'flex';
	document.getElementById('spacer').style.display = 'flex';
	document.querySelector('html').style = '';
}
