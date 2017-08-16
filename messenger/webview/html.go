package webview

const html = `<!DOCTYPE html>
<html>
	<head>
		<title>{{.Label.Title}}</title>

		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width,minimum-scale=1.0,maximum-scale=1.0">

		<style>
			body {
				font-family: Helvetica Neue, Helvetica, Arial, sans-serif;
			}
			.hide {
				display: none;
			}
			.search {
				width: 100%;
				padding: 3%;
				font-size: 100%;
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
				margin: 6% 3%;
				cursor: pointer;
			}
			.phrase span {
				width: 100%;
				display: inline-block;
				padding: 1.5% 0;
			}
		</style>
	</head>
	<body>

		<div class="wrapper">
			<input id="search" class="search" type="search" placeholder="{{.Label.Search}}" />
			<div id="empty" class="hide">{{.Label.Empty}}</div>
			<ul id="phrases" class="phrases"></ul>
		</div>


		<script>

			var phrases = [
				{{range $id, $phrase := .Phrases}}
				{
					id: {{$id}},
					phrase: '{{$phrase.Phrase}}',
					explanation: '{{$phrase.Explanation}}'
				},
				{{end}}
			]
			var phraseStates = {}

			var empty = document.getElementById('empty')
			if (!phrases.length) {
				empty.classList.remove('hide')
			}

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

			var container = document.getElementById('phrases')
			container.innerHTML = phrases.map(function(p) {
				return '<li class="phrase">'+
					'<span>'+p.phrase+'</span>'+
					'<span>'+p.explanation+'</span>'+
				'</li>'
			}).join('')

			var items = document.getElementsByClassName('phrase')

			function contains(a, b) {
				return a.toLowerCase().indexOf(b) !== -1
			}

		</script>
	</body>
</html>`
