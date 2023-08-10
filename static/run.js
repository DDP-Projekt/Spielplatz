"use strict";

let run_ws = null;
let compiling = false;

async function runProgram() {
	if (run_ws) {
		pushOutputMessage('Programm läuft bereits', 'stderr');
		return;
	} else if (compiling) {
		pushOutputMessage('Programm wird gerade kompiliert', 'stderr');
		return;
	}

	compiling = true;
	console.log('requesting to compile the program')
	const src_code = editor.getValue();
	// send a request to the /compile endpoint using the fetch api
	const compile_result = await fetch('/compile', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({ src: src_code }),
	}).then(response => response.json())

	if (compile_result.error) {
		console.log('compile error')
		pushOutputMessage(compile_result.stderr, 'stderr');
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
	if (run_ws) {
		run_ws.onopen = () => {
			console.log('websocket (run) connection opened')
		}

		run_ws.onmessage = (event) => {
			console.log(event);
			pushOutputMessage(event.data, 'stdout');
		}

		run_ws.onclose = () => {
			console.log('websocket (run) connection closed')
			run_ws = null;
			compiling = false;
		}

		run_ws.onerror = (error) => {
			console.log('websocket (run) error')
			pushOutputMessage(error.data, 'stderr');
		}
	} else {
		console.error('websocket (run) connection failed')
	}
}

function stopProgram() {
	console.log("terminating connection");
	if (run_ws) {
		run_ws.send(JSON.stringify({ msg: "EOF", eof: true }));
		run_ws = null;
		compiling = null;
	}
}