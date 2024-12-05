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
	// send a request to the /compile endpoint using the fetch api
	const compile_result = await fetch('compile', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({ src: code }),
	}).then(response => response.json())

	if (compile_result.error) {
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

	// connect to the /run endpoint using the websocket api with token as query parameter
	let ws_protocol = location.protocol === 'https:' ? "wss": "ws"
	run_ws = new WebSocket(`${ws_protocol}://${window.location.host}/run?token=${token}${argsString}`)
	if (!run_ws){
		console.error('websocket (run) connection failed')
		return;
	} 
	
	// focus input
	document.getElementById('input').focus();

	run_ws.onopen = () => {
		//console.log('websocket (run) connection opened')
	}

	run_ws.onmessage = (event) => {
		let msg = JSON.parse(event.data)
		pushOutputMessage(msg.msg, msg.isStderr ? MessageTarget.error : MessageTarget.output);
	}

	run_ws.onclose = (event) => {
		//console.log('websocket (run) connection closed: ', event)
		pushOutputMessage(" ", MessageTarget.system)
		pushOutputMessage(event.reason, event.code !== 1000 ? MessageTarget.error : MessageTarget.system)
		run_ws = null;
		compiling = false;

		toggleBtnVisibility();
	}

	run_ws.onerror = (error) => {
		console.error('websocket (run) error')
		pushOutputMessage(error.data, 'stderr');

		toggleBtnVisibility();
	}
}

function stopProgram() {
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