package webview

const html = `<!DOCTYPE html>
<html>
	<head>
		<title>{{.Label.Title}}</title>

		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width,minimum-scale=1.0,maximum-scale=1.0">

		<style>
			body,
			html {
				padding: 0;
				margin: 0;
			}
			body,
			textarea {
				font-family: Helvetica Neue, Helvetica, Arial, sans-serif;
			}
			.hide {
				display: none;
			}
			.noflow {
				overflow: hidden;
			}
			input[type="text"],
			input[type="search"],
			textarea {
				width: 94%;
				padding: 3%;
				margin: 3%;
				font-size: 100%;
				box-sizing: border-box;
			}
			textarea {
				height: 20%;
				resize: none;
				margin-top: 0;
			}
			button {
				border: 0;
				width: 33.3333%;
				padding: 6%;
				font-size: 110%;
				cursor: pointer;
				user-select: none;
			}
			.empty {
				padding: 10% 3%;
				text-align: center;
				font-weight: bold;
			}
			.phrases {
				list-style: none;
				padding: 0;
			}
			.phrase {
				margin: 4% 3%;
				cursor: pointer;
			}
			.phrase span {
				width: 100%;
				display: inline-block;
				padding: 1% 0;
			}
			.phrase span:first-child {
				font-weight: bold;
			}
			.total {
				margin: 10% 3%;
				font-size: 90%;
			}
			.edit {
				position: fixed;
				top: 0;
				background: white;
				width: 100%;
				height: 100%;
			}
			.actions {
				position: absolute;
				bottom: 0;
				width: 100%;
			}
			.fail {
				background-color: #ff4b4b;
			}
			.warn {
				background-color: #ffbe48;
			}
			.success {
				background-color: #6ae16a;
			}
			.half {
				width: 50%;
			}
			.update {
				position: fixed;
				bottom: 0;
				width: 100%;
				text-align: center;
				padding: 8%;
				font-size: 110%;
				box-sizing: border-box;
				user-select: none;
			}
		</style>
	</head>
	<body>

		<div class="wrapper">
			<div class="listing">
				<input id="search" class="search" type="search" placeholder="{{.Label.Search}}">
				<div id="empty" class="empty hide">{{.Label.Empty}}</div>
				<ul id="phrases" class="phrases"></ul>
				{{if len .Phrases | lt 5}}
				<div class="total">{{len .Phrases}} {{.Label.Phrases}}</div>
				{{end}}
			</div>
			<div id="edit" class="edit hide">
				<input id="edit-phrase" type="text" placeholder="{{.Label.Phrase}}">
				<textarea id="edit-explanation" placeholder="{{.Label.Explanation}}"></textarea>
				<div class="actions">
					<button id="edit-delete" class="fail">
						{{.Label.Delete}}
					</button><button id="edit-cancel" class="warn">
						{{.Label.Cancel}}
					</button><button id="edit-save" class="success">
						{{.Label.Save}}
					</button>
				</div>
				<div id="delete-prompt" class="actions hide">
					<button id="delete-confirm" class="half fail">
						{{.Label.Delete}}
					</button><button id="delete-cancel" class="half warn">
						{{.Label.Cancel}}
					</button>
				</div>
			</div>
			<div id="update-success" class="update success hide">{{.Label.Updated}}</div>
			<div id="delete-success" class="update success hide">{{.Label.Deleted}}</div>
			<div id="error" class="update fail hide">{{.Label.Error}}</div>
		</div>


		<script>

			var phrases = [
				{{range .Phrases}}
				{
					id: {{.ID}},
					phrase: '{{.Phrase}}',
					explanation: '{{.Explanation}}'
				},
				{{end}}
			]
			var phraseStates = {}
			var editI

			var empty = document.getElementById('empty')
			function handleEmpty() {
				if (!phrases.length) {
					empty.classList.remove('hide')
				}
			}
			handleEmpty()

			var container = document.getElementById('phrases')
			container.innerHTML = phrases.map(function(p) {
				return '<li class="phrase">'+
					'<span>'+p.phrase+'</span>'+
					'<span>'+p.explanation+'</span>'+
				'</li>'
			}).join('')

			var items = document.getElementsByClassName('phrase')

			var msgUpdate = document.getElementById('update-success')
			var msgDelete = document.getElementById('delete-success')
			var msgErr = document.getElementById('error')
			function msg(el) {
				el.classList.remove('hide')
				setTimeout(function() {
					el.classList.add('hide')
				}, 2000)
			}

			var edit = document.getElementById('edit')
			var editPhrase = document.getElementById('edit-phrase')
			var editExplanation = document.getElementById('edit-explanation')
			phrases.forEach(function(p, i) {
				items[i].addEventListener('click', function() {
					document.body.classList.add('noflow')
					edit.classList.remove('hide')
					editPhrase.value = p.phrase
					editExplanation.value = p.explanation
					editI = i
				})
			})
			function closeEdit() {
				document.body.classList.remove('noflow')
				edit.classList.add('hide')
				editI = undefined
			}
			document.getElementById('edit-cancel').addEventListener('click', closeEdit)
			var deletePrompt = document.getElementById('delete-prompt')
			document.getElementById('edit-delete').addEventListener('click', function() {
				deletePrompt.classList.remove('hide')
			})
			document.getElementById('delete-cancel').addEventListener('click', function() {
				deletePrompt.classList.add('hide')
			})

			function getURL() {
				return '{{.API}}/'+phrases[editI].id+'?token={{.Token}}'
			}
			document.getElementById('delete-confirm').addEventListener('click', function() {
				deletePrompt.classList.add('hide')
				var request = new XMLHttpRequest();
				request.open('DELETE', getURL(), true);
				request.onload = function() {
					if (request.status >= 400) {
						request.onerror(request.responseText)
						return
					}
					phrases.splice(editI, 1)
					container.removeChild(items[editI])
					handleEmpty()
					closeEdit()
					msg(msgDelete)
				};
				request.onerror = function(err) {
						msg(msgErr)
				};
				request.send();
			})

			document.getElementById('edit-save').addEventListener('click', function() {
				var p = editPhrase.value
				var e = editExplanation.value
				var request = new XMLHttpRequest();
				request.open('PUT', getURL(), true);
				request.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');
				request.onload = function() {
					if (request.status >= 400) {
						request.onerror(request.responseText)
						return
					}
					phrases[editI].phrase = p
					phrases[editI].explanation = e
					items[editI].children[0].innerText = p
					items[editI].children[1].innerText = e
					closeEdit()
					msg(msgUpdate)
				};
				request.onerror = function(msg) {
				};
				request.send(JSON.stringify({ phrase: p, explanation: e }));
			})

			var search = document.getElementById('search')
			search.addEventListener('input', function() {
				var query = search.value.toLowerCase()
				// Toggle phrases
				phrases.forEach(function(p, i) {
					var match = contains(p.phrase, query) || contains(p.explanation, query)
					if (match && phraseStates[i]) {
						items[i].classList.remove('hide')
						delete phraseStates[i]
					} else if (!match && !phraseStates[i]) {
						items[i].classList.add('hide')
						phraseStates[i] = true
					}
				})
				// Toggle placeholder
				if (phrases.length === Object.keys(phraseStates).length) {
					empty.classList.remove('hide')
				} else {
					empty.classList.add('hide')
				}
			})

			function contains(a, b) {
				return a.toLowerCase().indexOf(b) !== -1
			}

		</script>
	</body>
</html>`
