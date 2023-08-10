document.addEventListener("DOMContentLoaded", () => {
	const splitBtn = document.getElementById("split-btn");
	const main = document.getElementById("main");
	const editorContainer = document.getElementById("editor-container");
	const outputContainer = document.getElementById("output-container");

	if (window.localStorage.getItem("vertical") === "false") {
		main.setAttribute('horizontal', 'true');
		editorContainer.setAttribute('horizontal', 'true');
		outputContainer.setAttribute('horizontal', 'true');
		splitBtn.setAttribute('src', 'img/view-split-vertical.svg');
		document.getElementById("spacer").setAttribute('horizontal', 'true');

		window.localStorage.setItem("vertical", "false");
	}

	if (window.localStorage.getItem("content") === null) {
		document.getElementById("example-select").value = "HalloWelt";
	}

	main.style.visibility = "";
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

function clearOutput() {
	document.getElementById('outputText').innerHTML = '';
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
		splitBtn.setAttribute('src', 'img/view-split-vertical.svg');
		window.localStorage.setItem("vertical", "false");
	}
	else {
		splitBtn.setAttribute('src', 'img/view-split-horizontal.svg');
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

	ev.dataTransfer.setDragImage(document.createElement('br'), 0, 0);

	root.style.setProperty('--none-if-dragging', 'none');
}

function spacerDragEnd() {
	root.style.setProperty('--none-if-dragging', '');
}

function spacerDrag(ev) {
	const main = document.getElementById('main');

	if (main.hasAttribute('horizontal')) {
		let y = ev.clientY / ev.view.innerHeight;
		if (y < 0.1 || y > 0.9) return;
		root.style.setProperty('--editor-container-size', y * 100 + '%');
	}
	else {
		let x = ev.clientX / ev.view.innerWidth;
		if (x < 0.1 || x > 0.9) return;
		root.style.setProperty('--editor-container-size', x * 100 + '%');
	}
}

function copy() {
	navigator.clipboard.writeText(document.getElementById('output').innerText);
}

function openDocs() {
	window.open("https://ddp-projekt.github.io/Bedienungsanleitung/", "_blank").focus();
}

function pushOutputMessage(message, target) {
	const outputText = document.getElementById('outputText');
	const span = document.createElement('span');
	span.classList.add(target);
	span.innerHTML = message;
	outputText.appendChild(span);

	updateOutputScrollbar();
}

function updateOutputScrollbar() {
	var element = document.getElementById("output");
	element.scrollTop = element.scrollHeight;
}

function inputEnter(ev) {
	console.log(ev);
	let eof = false;
	let msg = ev.target.value + '\n';
	if (ev.ctrlKey && ev.key === 'c') {
		eof = true
		msg = "EOF";
	} else if (ev.key !== "Enter") {
		return;
	}
	ev.preventDefault();
	pushOutputMessage(msg, 'stdin');
	if (run_ws) {
		run_ws.send(JSON.stringify({ msg: msg, eof: eof }));
	}
	ev.target.value = "";
}