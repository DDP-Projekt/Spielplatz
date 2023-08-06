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

// ToDo: Maybe parse the ddp.tmLanguage.json into the IMonarchLanguage object format.
//https://raw.githubusercontent.com/DDP-Projekt/vscode-ddp/main/syntaxes/ddp.tmLanguage.json

monaco.languages.setMonarchTokensProvider('ddp', {
	tokenizer: {
		root: [
			// whitespace
			{ include: '@whitespace' },
			[/\b(([Ww]enn)|(dann)|([Ss]onst)|(aber)|([Ff](ü|(ue))r)|(jede[n]?)|(in)|([Ss]olange)|([Mm]ach(e|t))|(zur(ü|(ue))ck)|([Gg]ibt?)|([Vv]erlasse die Funktion)|(von)|(vom)|(bis)|(jede)|(jeder)|(Schrittgr(ö|(oe))(ß|(ss))e))|(Mal)|([Ww]iederhole)|((ö|(oe))ffentliche)\b/, 'keyword.control.ddp'],
			[/\b([Dd]er)|([Dd]ie)|([Dd]as)|(de[mn])|(ist)|(an)|(Stelle)|([Ss]peichere das Ergebnis von)|([Ss]peichere)|(einer)|(eine)|(leere[n]?)|(Liste)|(aus)|(besteht)|(Funktion)|(mit)|(Parameter[n]?)|(Typ)\b/, 'keyword.other.ddp'],
			[/\b((Zahl)|(Kommazahl)|(Boolean)|(Buchstabe[n]?)|(Text)|(Zahlen Liste)|(Kommazahlen Liste)|(Boolean Liste)|(Buchstaben Liste)|(Text Liste)|(Zahlen Referenz)|(Kommazahlen Referenz)|(Boolean Referenz)|(Buchstaben Referenz)|(Text Referenz)|(Zahlen Listen Referenz)|(Kommazahlen Listen Referenz)|(Boolean Listen Referenz)|(Buchstaben Listen Referenz)|(Text Listen Referenz))\b/, 'type.identifier'],
			[/(wahr)|(falsch)/, 'constant'],
			[/\b((oder)|(und)|(nicht)|(plus)|(minus)|(mal)|(durch)|(modulo)|(hoch)|(Wurzel)|(logisch)|(kontra)|(gleich)|(ungleich)|(kleiner)|(größer)|(groesser)|(als)|(Logarithmus)|(Betrag)|(Länge)|(Laenge)|(Größe)|(Groesse)|(um)|(Bit)|(verschoben)|(nach)|(links)|(rechts)|(zur)|(Basis)|(verkettet mit)|([Vv]erringere)|([Ee]rhöhe)|([Ee]rhoehe)|([Tt]eile)|([Vv]ervielfache)|([Ss]ubtrahiere)|([Aa]ddiere)|([Mm]ultipliziere)|([Dd]ividiere)|([Nn]egiere))\b/, 'keyword.operator'],
			[/([Uu]nd\s+kann\s+so\s+benutzt\s+werden)/, 'keyword.control.ddp'],
			[/Binde\s+(\"[\s\S]*\")\s+ein/, 'keyword.ddp']
		],
		comment: [
			[/[^\[\]]+/, 'comment' ],
			[/\[/,    'comment', '@push' ],    // nested comment
			[/\]/,    'comment', '@pop'  ],
			[/[\]*]/,   'comment' ]
		],
		whitespace: [
			[/[ \t\r\n]+/, 'white'],
			[/\[/,       'comment', '@comment' ],
			[/\/\/.*$/,    'comment'],
		],
	}
});

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
