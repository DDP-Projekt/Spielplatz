const editor = monaco.editor.create(document.getElementById('container'), {
	value: [
		'Die Zahl z ist 22.',
	].join('\n'),
	language: 'ddp',
	theme: 'vs-dark',
	'semanticHighlighting.enabled': true,
});

// add a new language called ddp to the editor
monaco.languages.register({ id: 'ddp' });

// connect to a websocket on the /ls endpoint
const socket = new WebSocket(`ws://${window.location.host}/ls`);
const file_uri = 'file:///main.ddp';
let initialized = false;

//a function that takes a javascript object and sends it to the language server
function send(msg) {
	// send the msg to the language server and add basic jsonrpc fields
	socket.send(JSON.stringify({
		jsonrpc: '2.0',
		id: 1,
		...msg,
	}));
}

socket.onclose = () => {
	console.log('disconnected from /ls');
};

// holds callback promises for expected responses
let response_queue = [];

function push_response_handler() {
	// return a promise that is fullfilled when socket.onmessage calls resp_handler
	return new Promise((resolve) => {
		response_queue.push(resolve);
	});
}

function pull_response_handler() {
	let resp = response_queue[0];
	response_queue = response_queue.slice(1);
	return resp;
}

function discard_response(resp) {
	console.log('discarding response', resp);
}

socket.onmessage = (event) => {
	// handle langue server protocol messages
	const msg = JSON.parse(event.data);
	// handle errors
	if (msg.error) {
		console.error(msg.error);
		return;
	}

	if (msg.result !== undefined) {
		let resolve = pull_response_handler()
		resolve(msg);
		return;
	}

	switch (msg.method) {
		case 'textDocument/publishDiagnostics':
			console.log('diagnostics');
			// handle diagnostic messages
			const diagnostics = msg.params.diagnostics;
			const markers = [];
			for (const diagnostic of diagnostics) {
				markers.push({
					severity: monaco.MarkerSeverity.Error,
					startLineNumber: diagnostic.range.start.line + 1,
					startColumn: diagnostic.range.start.character + 1,
					endLineNumber: diagnostic.range.end.line + 1,
					endColumn: diagnostic.range.end.character + 1,
					message: diagnostic.message,
					source: 'ddp',
				});
			}
			monaco.editor.setModelMarkers(editor.getModel(), 'ddp', markers);
			break;
		default:
			console.log('unknown message', msg);
	}
};

socket.onopen = () => {
	console.log('connected to /ls');
	// send a language server protocol initialize request
	send({
		method: 'initialize',
		params: {
			dynamicRegistration: false,
			processId: null,	// not used
			rootUri: null,		// not used
			capabilities: {
				textDocument: {
					synchronization: {
						willSave: false,
						willSaveWaitUntil: false,
						didSave: true,
					},
					semanticTokens: {
						requests: {
							full: true,
						},
						tokenTypes: [
							'keyword',
							'number',
							'string',
							'comment',
							'operator',
							'punctuation',
							'variable',
							'function',
							'class',
							'type',
						],
						tokenModifiers: [
							'declaration',
							'definition',
							'reference',
							'call',
							'write',
							'await',
							'import',
							'export',
							'local',
						],
					},
					// add completion capabilities with snipped support disabled
					completion: {
						completionItem: {
							snippetSupport: false,
						},
					},
				},
			},
			clientInfo: {
				name: 'monaco-editor',
				version: '1.0.0',
			},
			initializationOptions: null,
		}
	});
	push_response_handler().then(handleInitializeResponse);
};

let semantic_tokens_lengend = {}
function handleInitializeResponse(resp) {
	initialized = true;
	console.log('initializeResult', resp)

	semantic_tokens_lengend = resp.result.capabilities.semanticTokensProvider.legend;
	monaco.languages.registerDocumentSemanticTokensProvider('ddp', {
		// add the supported token types from DDPLS
		getLegend: () => semantic_tokens_lengend,
		// request semantic tokens
		provideDocumentSemanticTokens: async (model, lastResultId, token) => {
			send({
				method: 'textDocument/semanticTokens/full',
				params: {
					textDocument: {
						uri: file_uri,
					},
				},
			});
			return push_response_handler().then((resp) => {
				console.log('semantic tokens');
				// handle semantic token response
				const tokens = resp.result;
				return tokens;
			});
		},
		releaseDocumentSemanticTokens: (resultId) => { },
	});

	console.log('initialized')
	// send a language server protocol initialized notification	
	send({
		method: 'initialized',
		params: {},
	});
	push_response_handler().then(discard_response);
	// send a language server protocol didOpen notification
	send({
		method: 'textDocument/didOpen',
		params: {
			textDocument: {
				uri: file_uri,
				languageId: 'ddp',
				version: 1,
				text: editor.getValue(),
			},
		},
	});
	push_response_handler().then(discard_response);
}


// when the editor is changed, send a didChange notification to the language server
editor.onDidChangeModelContent((event) => {
	if (!initialized) {
		return;
	}

	send({
		method: 'textDocument/didChange',
		params: {
			textDocument: {
				uri: file_uri,
				version: 2,
			},
			contentChanges: [{
				range: {
					start: {
						line: event.changes[0].range.startLineNumber - 1,
						character: event.changes[0].range.startColumn - 1,
					},
					end: {
						line: event.changes[0].range.endLineNumber - 1,
						character: event.changes[0].range.endColumn - 1,
					},
				},
				rangeLength: event.changes[0].rangeLength,
				text: event.changes[0].text,
			}],
		},
	});
	push_response_handler().then(discard_response);
});

// when the website is closed, send a shutdown request to the language server
window.onbeforeunload = () => {
	send({
		method: 'shutdown',
		params: {},
	});
	push_response_handler().then((resp) => {
		// send a language server protocol exit notification
		send({
			method: 'exit',
			params: {},
		});
		// close the websocket
		socket.close();
	});
}
