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
			input,
			textarea {
				width: 94%;
				padding: 2%;
				margin: 3%;
				font-size: 100%;
				box-sizing: border-box;
				border: 1px solid #dedede;
			}
			textarea {
				height: 4.8em;
				resize: none;
				margin-top: 0;
				margin-bottom: 0;
				line-height: 120%;
			}
			button {
				border: 1px solid;
				width: 29.333%;
				padding: 2.5% 0;
				font-size: 100%;
				cursor: pointer;
				margin: 1.5%;
				background: white;
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
				margin: 2% 0;
				padding: 1% 3%;
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
			.open {
				background: #f1f1f1;
			}
			.total {
				margin: 10% 3%;
				font-size: 90%;
			}
			.edit {
				position: fixed;
				-webkit-backface-visibility: hidden;
				bottom: 0;
				background: #f1f1f1;
				width: 100%;
				box-shadow: 0 -2px 6px 2px white;
			}
			.actions {
				width: 100%;
				margin: 1.5%;
				background: #f1f1f1;
			}
			.delete-prompt {
				position: absolute;
				bottom: 0;
			}
			.half {
				width: 45.5%;
			}
			.update {
				position: fixed;
				-webkit-backface-visibility: hidden;
				bottom: 0;
				width: 100%;
				text-align: center;
				padding: 7%;
				box-sizing: border-box;
				background: #f1f1f1;
			}
			button,
			.search,
			.edit,
			.update,
			.total,
			.empty {
				-webkit-touch-callout: none;
				-webkit-user-select: none;
				-khtml-user-select: none;
				-moz-user-select: none;
				-ms-user-select: none;
				user-select: none;
			}
			button,
			.phrase {
				-webkit-tap-highlight-color: rgba(0, 0, 0, .1);
			}
			button::-moz-focus-inner {
				border: 0;
			}
			button:hover,
			button:focus,
			.phrase:hover {
				outline: none;
				background-color: #dedede;
			}
			input:focus,
			textarea:focus {
				outline: none;
				border: 1px solid #939393;
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
				<div id="delete-prompt" class="delete-prompt actions hide">
					<button id="delete-confirm" class="half fail">
						{{.Label.DeleteConfirm}}
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

			var msgTimeout
			var msgTimeoutEl
			function msg(el) {
				el.classList.remove('hide')
				msgTimeoutEl = el
				msgTimeout = setTimeout(function() {
					el.classList.add('hide')
				}, 2000)
			}

			var edit = document.getElementById('edit')
			var editPhrase = document.getElementById('edit-phrase')
			var editExplanation = document.getElementById('edit-explanation')
			phrases.forEach(function(p, i) {
				var el = items[i]
				el.addEventListener('click', function() {
					closeEdit()

					if (msgTimeout) {
						clearTimeout(msgTimeout)
						msgTimeoutEl.classList.add('hide')
						msgTimeout = undefined
						msgTimeoutEl = undefined
					}

					edit.classList.remove('hide')

					editPhrase.value = p.phrase
					editExplanation.value = p.explanation

					el.classList.add('open')

					editI = Array.prototype.indexOf.call(container.children, el)
				})
			})

			function closeEdit() {
				if (editI === undefined) return
				items[editI].classList.remove('open')
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
					delete phraseStates[editI]
					container.removeChild(items[editI])
					items = document.getElementsByClassName('phrase')
					handleEmpty()
					edit.classList.add('hide')
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
