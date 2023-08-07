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
		window.localStorage.setItem("vertical", "false");
	}

	if (window.localStorage.getItem("content") === null) {
		document.getElementById("example-select").value = "HalloWelt";
	}
});

let saveCount = 0;
document.addEventListener("keydown", function(e) {
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
	document.getElementById('output').innerHTML = '';
}

function split() {
	const splitBtn = document.getElementById("split-btn");
	const main = document.getElementById("main");
	const editorContainer = document.getElementById("editor-container");
	const outputContainer = document.getElementById("output-container");

	main.toggleAttribute('horizontal');
	editorContainer.toggleAttribute('horizontal');
	outputContainer.toggleAttribute('horizontal');

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