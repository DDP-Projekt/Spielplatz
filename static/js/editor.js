"use strict";

monaco.editor.defineTheme('ddp-theme-dark', {
	base: 'vs-dark',
	inherit: true,
	colors: {},
	rules: [
		{ token: 'keyword.control', foreground: 'C586C0' },
		{ token: 'invalid', foreground: 'F44640' },
		{ token: 'string.escape', foreground: 'D7BA7D' },
		{ token: 'keyword.controlFlow', foreground: 'C586C0' },
		{ token: 'variable', foreground: '9CDCFE' },
		{ token: 'parameter', foreground: '9CDCFE' },
		{ token: 'property', foreground: '9CDCFE' },
		{ token: 'support.function', foreground: 'DCDCAA' },
		{ token: 'function', foreground: 'DCDCAA' },
		{ token: 'member', foreground: '4FC1FF' },
		{ token: 'variable.constant', foreground: '4FC1FF' },
		{ token: 'typeParameter', foreground: '4EC9B0' },
	]
});

monaco.editor.defineTheme('ddp-theme-light', {
	base: 'vs',
	inherit: true,
	colors: {},
	rules: [
		{ token: 'keyword.control', foreground: 'AF00DB' },
		{ token: 'invalid', foreground: 'F44640' },
		{ token: 'string.escape', foreground: 'D7BA7D' },
		{ token: 'keyword.controlFlow', foreground: 'AF00DB' },
		{ token: 'variable', foreground: '0070C1' },
		{ token: 'parameter', foreground: '9CDCFE' },
		{ token: 'property', foreground: '9CDCFE' },
		{ token: 'support.function', foreground: '795E26' },
		{ token: 'function', foreground: '795E26' },
		{ token: 'member', foreground: '4FC1FF' },
		{ token: 'variable.constant', foreground: '4FC1FF' },
		{ token: 'typeParameter', foreground: '4EC9B0' },
	]
});

let value = 'Binde "Duden/Ausgabe" ein.\nSchreibe "Hallo Welt".';
const initialContent = window.localStorage.getItem("content");
if (initialContent !== null && !embedded) {
	value = initialContent;
}

const urlParams = new URLSearchParams(window.location.search);
if (urlParams.has("code")) {
	value = decodeURIComponent(LZUTF8.decompress(urlParams.get("code"), { inputEncoding: "Base64" }));
}

let editorTheme = 'ddp-theme-dark';
if (window.localStorage.getItem("dark-mode") === 'false' || urlParams.has('light')) {
	editorTheme = "ddp-theme-light";
	document.querySelector('html').setAttribute('light', '')
}

const isReadOnly = urlParams.has("readonly");
const editorDiv = document.getElementById('editor');
const file_uri = monaco.Uri.parse('inmemory://Spielplatz');
const editor = monaco.editor.create(editorDiv, {
	theme: editorTheme,
	'semanticHighlighting.enabled': true,
	//automaticLayout: true,
	model: monaco.editor.createModel(value, 'ddp', file_uri),
	minimap: { enabled: !embedded },
	readOnly: isReadOnly,
	lineNumbers: !urlParams.has("nolines"),
	scrollbar: {
		vertical: urlParams.has("noscroll") ? "hidden" : undefined,	
		handleMouseWheel: !urlParams.has("noscroll")
	}
});

new ResizeObserver(function (mutations) {
	editor.layout();
}).observe(editorDiv);

// add a new language called ddp to the editor
monaco.languages.register({ id: 'ddp' });
monaco.languages.setLanguageConfiguration('ddp', {
	comments: {
		blockComment: ['[', ']'],
	},
	autoClosingPairs: [
		{ open: '[', close: ']' },
		{ open: '(', close: ')' },
		{ open: '"', close: '"', notIn: ['string'] },
		{ open: '\'', close: '\'', notIn: ['string'] },
	],
	surroundingPairs: [
		{ open: '[', close: ']' },
		{ open: '(', close: ')' },
		{ open: '"', close: '"' },
	]
})

// ToDo: Maybe parse the ddp.tmLanguage.json into the IMonarchLanguage object format.
//https://raw.githubusercontent.com/DDP-Projekt/vscode-ddp/main/syntaxes/ddp.tmLanguage.json

monaco.languages.setMonarchTokensProvider('ddp', {
	escapes: /\\(?:[nrbta\\])/,

	tokenizer: {
		root: [
			// whitespace
			{ include: '@whitespace' },
			
			// functions
			[/(ist\s+in\s+)("[\s\S]+")(\s+definiert)/, ['keyword', 'string', 'keyword']],
			[/([Uu]nd\s+kann\s+so\s+benutzt\s+werden)/, 'keyword.control.ddp'],	
			[/(Der\s+)(Alias\s+)("[\s\S]+")(\s+steht\s+für\s+die\s+Funktion\s+)([\wäöüÄÖÜ]+)/, ['keyword', 'type', 'string', 'keyword', 'function']],
			
			// types
			[/\b((Zahl)|(Kommazahl)|(Boolean)|(Buchstabe[n]?)|(Text)|(Zahlen Liste)|(Kommazahlen Liste)|(Boolean Liste)|(Buchstaben Liste)|(Text Liste)|(Zahlen Referenz)|(Kommazahlen Referenz)|(Boolean Referenz)|(Buchstaben Referenz)|(Text Referenz)|(Zahlen Listen Referenz)|(Kommazahlen Listen Referenz)|(Boolean Listen Referenz)|(Buchstaben Listen Referenz)|(Text Listen Referenz)|(nichts))\b/, 'type.identifier'],
			
			// Keywords
			[/\b(([Ww]enn)|(dann)|([Ss]onst)|(aber)|([Ff](ü|(ue))r)|(jede[n]?)|(in)|([Ss]olange)|([Mm]ach(e|t))|(zur(ü|(ue))ck)|([Gg]ibt?)|([Vv]erlasse die Funktion)|(von)|(vom)|(bis)|(jede)|(jeder)|(Schrittgr(ö|(oe))(ß|(ss))e))|(Mal)|([Ww]iederhole)|((ö|(oe))ffentliche)\b/, 'keyword.control.ddp'],
			[/\b([Dd]er)|([Dd]ie)|([Dd]as)|(de[mn])|(ist)|(an)|(Stelle)|([Ss]peichere das Ergebnis von)|([Ss]peichere)|(einer)|(einen?)|(leere[n]?)|(Liste)|(aus)|(besteht)|(Funktion)|(mit)|(Parameter[n]?)|(Typ)\b/, 'keyword.other.ddp'],
			
			// constants
			[/(wahr)|(falsch)/, 'constant'],
			
			// Operatoren
			[/\b((oder)|(und)|(nicht)|(plus)|(minus)|(mal)|(durch)|(modulo)|(hoch)|(Wurzel)|(logisch)|(kontra)|(gleich)|(ungleich)|(kleiner)|(größer)|(groesser)|(als)|(Logarithmus)|(Betrag)|(Länge)|(Laenge)|(Größe)|(Groesse)|(um)|(Bit)|(verschoben)|(nach)|(links)|(rechts)|(zur)|(Basis)|(verkettet mit)|([Vv]erringere)|([Ee]rhöhe)|([Ee]rhoehe)|([Tt]eile)|([Vv]ervielfache)|([Ss]ubtrahiere)|([Aa]ddiere)|([Mm]ultipliziere)|([Dd]ividiere)|([Nn]egiere))\b/, 'keyword.operator'],
			
			// Einbinden
			[/Binde\s+("[\s\S]*")\s+ein/, 'keyword.ddp'],
			[/(Binde\s+)([\wäöüÄÖÜ]+)(\s+aus\s+)("[\s\S]*")(\s+ein)/, ['keyword', 'function', 'keyword', 'string', 'keyword']],
			[/(Binde\s+)([\wäöüÄÖÜ]+(?:,\s*[\wäöüÄÖÜ]+)*)(\s+und\s+)([\wäöüÄÖÜ]+)(\s+aus\s+)("[\s\S]*")(\s+ein)/, ['keyword', 'function', 'keyword', 'function', 'keyword', 'string', 'keyword']],

			// strings
			[/"([^"\\]|\\.)*$/, 'string.invalid' ],  // non-teminated string
			[/"/,  { token: 'string.quote', bracket: '@open', next: '@string' } ],
			
			// chars
			[/'[^'\n\\]'/, 'string'],
			[/'(@escapes|\\')'/, 'string.escape'],
			[/'[^'\n]{2,}'|''/, 'invalid']
		],
		comment: [
			[/[^\[\]]+/, 'comment'],
			[/\[/, 'comment', '@push'],    // nested comment
			[/\]/, 'comment', '@pop'],
			[/[\]*]/, 'comment']
		],
		string: [
			[/[^\\"]+/,  'string'],
			[/@escapes|\\"/, 'string.escape'],
			[/\\./,      'invalid'],
			[/"/,        { token: 'string.quote', bracket: '@close', next: '@pop' } ]
		],
		whitespace: [
			[/[ \t\r\n]+/, 'white'],
			[/\[/, 'comment', '@comment'],
			[/\/\/.*$/, 'comment'],
		],
	}
});

// connect to a websocket on the /ls endpoint
let ws_protocol = location.protocol === 'https:' ? "wss" : "ws"
const ls_socket = new WebSocket(`${ws_protocol}://${window.location.host}/ls`);
ls_socket.onerror = (error) => {
	console.error('WebSocket error:', error);
};
let initialized = false;

//a function that takes a javascript object and sends it to the language server
function send(msg) {
	// return if the socket is not able to send
	if (ls_socket.readyState !== WebSocket.OPEN) {
		return;
	}
	// send the msg to the language server and add basic jsonrpc fields
	ls_socket.send(JSON.stringify({
		jsonrpc: '2.0',
		id: 1,
		...msg,
	}));
}

ls_socket.onclose = () => {
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
	//console.log('discarding response', resp);
}

ls_socket.onmessage = (event) => {
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

ls_socket.onopen = () => {
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

// map completion.kind from lsp kind to monaco kind
const completion_kind_map = {
	1: monaco.languages.CompletionItemKind.Method,
	2: monaco.languages.CompletionItemKind.Function,
	3: monaco.languages.CompletionItemKind.Constructor,
	4: monaco.languages.CompletionItemKind.Field,
	5: monaco.languages.CompletionItemKind.Variable,
	6: monaco.languages.CompletionItemKind.Class,
	7: monaco.languages.CompletionItemKind.Interface,
	8: monaco.languages.CompletionItemKind.Module,
	9: monaco.languages.CompletionItemKind.Property,
	10: monaco.languages.CompletionItemKind.Unit,
	11: monaco.languages.CompletionItemKind.Value,
	12: monaco.languages.CompletionItemKind.Enum,
	13: monaco.languages.CompletionItemKind.Keyword,
	14: monaco.languages.CompletionItemKind.Snippet,
	15: monaco.languages.CompletionItemKind.Text,
	16: monaco.languages.CompletionItemKind.Color,
	17: monaco.languages.CompletionItemKind.File,
	18: monaco.languages.CompletionItemKind.Reference,
	19: monaco.languages.CompletionItemKind.Folder,
	20: monaco.languages.CompletionItemKind.EnumMember,
	21: monaco.languages.CompletionItemKind.Constant,
	22: monaco.languages.CompletionItemKind.Struct,
	23: monaco.languages.CompletionItemKind.Event,
	24: monaco.languages.CompletionItemKind.Operator,
	25: monaco.languages.CompletionItemKind.TypeParameter,
};
let semantic_tokens_lengend = {};
let cached_readonly_semantic_tokens = null;
let last_completion_request_timestamp = new Date().getTime();
function handleInitializeResponse(resp) {
	initialized = true;

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
				uri: file_uri.toString(),
				languageId: 'ddp',
				version: 1,
				text: editor.getValue(),
			},
		},
	});
	push_response_handler().then(discard_response);

	semantic_tokens_lengend = resp.result.capabilities.semanticTokensProvider.legend;
	monaco.languages.registerDocumentSemanticTokensProvider('ddp', {
		// add the supported token types from DDPLS
		getLegend: () => semantic_tokens_lengend,
		// request semantic tokens
		provideDocumentSemanticTokens: async (model, lastResultId, token) => {
			if (cached_readonly_semantic_tokens !== null) {
				return cached_readonly_semantic_tokens;
			}

			send({
				method: 'textDocument/semanticTokens/full',
				params: {
					textDocument: {
						uri: file_uri.toString(),
					},
				},
			});
			return push_response_handler().then((resp) => {
				if (!resp.result) {
					return null;
				}
				// handle semantic token response
				const tokens = resp.result;
				// if we are readOnly we only need to fetch the tokens once
				// and can then close the websocket connection to safe resources on the server
				if (isReadOnly) {
					cached_readonly_semantic_tokens = tokens;
					ls_socket.close();
				}
				return tokens;
			});
		},
		releaseDocumentSemanticTokens: (resultId) => { },
	});

	// if the editor is readOnly we don't need all those things below because we
	// only want to view and run the code, not edit it (we don't need hover or goto definition etc.)
	if (!isReadOnly) {
		// add support for textDocument/semanticTokens/range
		monaco.languages.registerDocumentRangeSemanticTokensProvider('ddp', {
			getLegend: () => semantic_tokens_lengend,
			// request semantic tokens
			provideDocumentRangeSemanticTokens: async (model, range, token) => {
				send({
					method: 'textDocument/semanticTokens/range',
					params: {
						textDocument: {
							uri: file_uri.toString(),
						},
						range: {
							start: {
								line: range.startLineNumber - 1,
								character: range.startColumn - 1,
							},
							end: {
								line: range.endLineNumber - 1,
								character: range.endColumn - 1,
							},
						},
					},
				});
				return push_response_handler().then((resp) => {
					if (!resp.result) {
						return null;
					}
					// handle semantic token response
					const tokens = resp.result;
					return tokens;
				});
			},
		});

		// register a completion provider
		// commented out for performance reasons and buggy behavior
		/*
		monaco.languages.registerCompletionItemProvider('ddp', {
			triggerCharacters: resp.result.capabilities.completionProvider.triggerCharacters,
			provideCompletionItems: async (model, position, context, token) => {
				// get the current time in milliseconds
				const current_time = new Date().getTime();
				// if the last completion request was less than 500ms ago, don't send another one
				if (current_time - last_completion_request_timestamp < 500) {
					return null;
				}
				last_completion_request_timestamp = current_time;
				// send a language server protocol completion request
				//console.log('requesting completion for trigger character', context.triggerCharacter);
				send({
					method: 'textDocument/completion',
					params: {
						textDocument: {
							uri: file_uri.toString(),
						},
						position: {
							line: position.lineNumber - 1,
							character: position.column - 1,
						},
						context: {
							triggerKind: context.triggerKind,
							triggerCharacter: context.triggerCharacter,
						},
					},
				});
				return push_response_handler().then((resp) => {
					if (!resp.result) {
						return null;
					}

					// handle completion response
					const suggestions = resp.result.map((completion) => {
						const kind = completion_kind_map[completion.kind];
						return {
							label: completion.label,
							kind: kind,
							insertText: completion.insertText ? completion.insertText : completion.label,
							range: {
								startLineNumber: position.lineNumber,
								startColumn: position.column,
								endLineNumber: position.lineNumber,
								endColumn: position.column,
							},
							sortText: String.fromCharCode(97 + kind) + completion.label,
						};
					});
					return {
						suggestions: suggestions,
					};
				});
			},
		});*/

		// register a hover provider
		monaco.languages.registerHoverProvider('ddp', {
			provideHover: async (model, position, token) => {
				// send a language server protocol hover request
				send({
					method: 'textDocument/hover',
					params: {
						textDocument: {
							uri: file_uri.toString(),
						},
						position: {
							line: position.lineNumber - 1,
							character: position.column - 1,
						},
					},
				});
				return push_response_handler().then((resp) => {
					if (!resp.result) {
						return null;
					}
					return {
						contents: [
							{ value: resp.result.contents.value, },
						],
						range: resp.result.range,
					};
				});
			}
		});

		// register a definition provider	
		monaco.languages.registerDefinitionProvider('ddp', {
			provideDefinition: async (model, position, token) => {
				// send a language server protocol definition request
				send({
					method: 'textDocument/definition',
					params: {
						textDocument: {
							uri: file_uri.toString(),
						},
						position: {
							line: position.lineNumber - 1,
							character: position.column - 1,
						},
					},
				});
				return push_response_handler().then((resp) => {
					// handle definition response
					if (!resp.result) {
						return null;
					}
					return {
						uri: resp.result.uri,
						range: {
							startLineNumber: resp.result.range.start.line + 1,
							startColumn: resp.result.range.start.character + 1,
							endLineNumber: resp.result.range.end.line + 1,
							endColumn: resp.result.range.end.character + 1,
						}
					};
				});
			}
		});
	}
}

let cached_changes = {
	length: 0,
	changes: [],
};
// update the language server every 250ms even if there was only a single change
setInterval(() => {
	if (cached_changes.changes.length > 0) {
		doChangeRequest(cached_changes.changes);
		cached_changes.length = 0;
		cached_changes.changes = [];
	}
}, 250);
if (!isReadOnly) {
	// when the editor is changed, send a didChange notification to the language server
	editor.onDidChangeModelContent((event) => {
		if (!initialized) {
			return;
		}

		cached_changes.length += event.changes.reduce((acc, change) => acc + change.text.length, 0);
		cached_changes.changes.push(...event.changes);

		// only update the language server immediately if there are more than 15 changes
		if (cached_changes.length > 15) {
			doChangeRequest(cached_changes.changes);
			cached_changes.length = 0;
			cached_changes.changes = [];
		}
	});
}

function doChangeRequest(changes) {
	send({
		method: 'textDocument/didChange',
		params: {
			textDocument: {
				uri: file_uri.toString(),
				version: 2,
			},
			contentChanges:
				// map all changes to lsp changes
				changes.map((change) => {
					return {
						range: {
							start: {
								line: change.range.startLineNumber - 1,
								character: change.range.startColumn - 1,
							},
							end: {
								line: change.range.endLineNumber - 1,
								character: change.range.endColumn - 1,
							},
						},
						rangeLength: change.rangeLength,
						text: change.text,
					};
				}),
		},
	});
	push_response_handler().then(discard_response);
}

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
		ls_socket.close();
	});

	if (!embedded) {
		window.localStorage.setItem("content", editor.getValue());
	}

	const argsContainer = document.getElementById("args");
	window.localStorage.setItem("args", argsContainer.value);
}
