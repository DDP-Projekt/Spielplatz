"use strict";

let run_ws = null;
let compiling = false;

async function runProgram(code) {
	if (run_ws) {
		pushOutputMessage('Programm läuft bereits.', MessageTarget.error);
		return;
	} else if (compiling) {
		pushOutputMessage('Programm wird gerade kompiliert.', MessageTarget.error);
		return;
	}

	if (autoClear) {
		clearOutput();
	}

	toggleBtnVisibility();

	compiling = true;
	console.log('requesting to compile the program')
	// send a request to the /compile endpoint using the fetch api
	const compile_result = await fetch('/compile', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({ src: code }),
	}).then(response => response.json())

	if (compile_result.error) {
		console.log('compile error')
		pushOutputMessage("Kompilier Fehler: " + compile_result.error, MessageTarget.error);
		pushOutputMessage(" ", MessageTarget.error);
		pushOutputMessage(compile_result.stderr, MessageTarget.error);

		compiling = false;
		toggleBtnVisibility();
		return
	}

	const token = compile_result.token
	const args = document.getElementById('args').value;
	let argsString = ""
	for (let arg of args.split(';')) {
		argsString += "&args=" + arg;
	}

	console.log('requesting to run the program')
	// connect to the /run endpoint using the websocket api with token as query parameter
	run_ws = new WebSocket(`ws://${window.location.host}/run?token=${token}${argsString}`)
	if (!run_ws){
		console.error('websocket (run) connection failed')
		return;
	} 
	
	run_ws.onopen = () => {
		console.log('websocket (run) connection opened')
	}

	run_ws.onmessage = (event) => {
		console.log(event);
		let msg = JSON.parse(event.data)
		pushOutputMessage(msg.msg, msg.isStderr ? MessageTarget.error : MessageTarget.output);
	}

	run_ws.onclose = (event) => {
		console.log('websocket (run) connection closed: ', event)
		pushOutputMessage(" ", MessageTarget.system)
		pushOutputMessage(event.reason, event.code !== 1000 ? MessageTarget.error : MessageTarget.system)
		run_ws = null;
		compiling = false;

		toggleBtnVisibility();
	}

	run_ws.onerror = (error) => {
		console.log('websocket (run) error')
		pushOutputMessage(error.data, 'stderr');

		toggleBtnVisibility();
	}
}

function stopProgram() {
	console.log("terminating connection");
	if (!run_ws) {
		return;
	}

	pushOutputMessage("Das Programm wurde abgebrochen", MessageTarget.system);
	run_ws.send(JSON.stringify({ msg: "EOF", eof: true }));
	run_ws = null;
	compiling = null;
}

function toggleBtnVisibility() {
	document.getElementById("run-btn").toggleAttribute("hidden");
	document.getElementById("stop-btn").toggleAttribute("hidden");
}