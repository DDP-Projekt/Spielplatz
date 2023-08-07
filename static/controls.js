document.addEventListener("DOMContentLoaded", () => {
	const splitBtn = document.getElementById("split-btn");
	const main = document.getElementById("main");
	const editorContainer = document.getElementById("editor-container");
	const outputContainer = document.getElementById("output-container");

	console.log(window.localStorage.getItem("vertical"));
	if (window.localStorage.getItem("vertical") === "false") {
		main.setAttribute('horizontal', 'true');
		editorContainer.setAttribute('horizontal', 'true');
		outputContainer.setAttribute('horizontal', 'true');
		splitBtn.setAttribute('src', 'img/view-split-vertical.svg');
		window.localStorage.setItem("vertical", "false");
	}
});

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