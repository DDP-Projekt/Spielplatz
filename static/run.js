async function runProgram() {
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
	let arguments = ""
	for (let arg of args.split(';')) {
		arguments += "&args=" + arg;
	}

	console.log('requesting to run the program')
	// connect to the /run endpoint using the websocket api with token as query parameter
	const ws = new WebSocket(`ws://${window.location.host}/run?token=${token}${arguments}`)
	if (ws) {
		ws.onopen = () => {
			console.log('websocket (run) connection opened')
			ws.send(JSON.stringify({ msg: 'ein bisschen input', eof: false }));
			ws.send(JSON.stringify({ msg: 'noch mehr input', eof: false }));
			ws.send(JSON.stringify({ msg: '', eof: true }));
			console.log('input sent')
		}

		ws.onmessage = (event) => {
			console.log(event);
			pushOutputMessage(event.data, 'stdout');
		}

		ws.onclose = () => {
			console.log('websocket (run) connection closed')
		}

		ws.onerror = (error) => {
			console.log('websocket (run) error')
			pushOutputMessage(error.data, 'stderr');
		}
	} else {
		console.error('websocket (run) connection failed')
	}
}