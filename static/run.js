async function runProgram() {
	console.log('requesting to compile the program')
	src_code = editor.getValue();
	// send a request to the /compile endpoint using the fetch api
	compile_result = await fetch('/compile', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({ src: src_code }),
	}).then(response => response.json())

	if (compile_result.error) {
		console.log('compile error')
		console.log(compile_result)
		return
	}

	token = compile_result.token
	console.log('requesting to run the program')
	// connect to the /run endpoint using the websocket api with token as query parameter
	ws = new WebSocket(`ws://${window.location.host}/run?token=${token}`)
	if (ws) {
		ws.onopen = () => {
			console.log('websocket (run) connection opened')
			ws.send('ein bisschen input')
		}

		ws.onmessage = (event) => {
			console.log(event.data)
		}

		ws.onclose = () => {
			console.log('websocket (run) connection closed')
		}

		ws.onerror = (error) => {
			console.log('websocket (run) error')
			console.log(error)
		}
	} else {
		console.error('websocket (run) connection failed')
	}
}