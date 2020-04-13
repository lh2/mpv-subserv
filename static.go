package main

var INDEX = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>{{.title}}</title>
		<link rel="stylesheet" href="/style.css">
	</head>
	<body lang="{{.lang}}">
		<div id="latest"></div>
		<ul id="history"></ul>
		<script src="script.js"></script>
	</body>
</html>
`

const STYLE = `
* {
	margin: 0;
	padding: 0;
}
body {
	font-size: 1.4em;
	background-color: #ffffea;
}
#latest {
	font-size: 2em;
	background-color: #eaffff;
	padding: 20px;
}
#history {
	list-style: none;
}
#history li {
	padding: 5px;
}
#history li:nth-child(even) {
	background-color: #ffffd8;
}
`

const SCRIPT = `
let subs = {};
let lsub = document.getElementById('latest');
let ul = document.getElementById('history');
const addr = 'ws://' + location.host + '/socket';
let ws = new WebSocket(addr);
ws.onmessage = function(e) {
	var msg = JSON.parse(e.data);
	console.dir(msg);
	subs[msg.id] = msg;

	const last = latest.getAttribute('data-id');
	if (last != null) {
		let lastMsg = subs[last];
		let li = document.createElement('LI');
		li.innerText = lastMsg.line;
		li.setAttribute('data-id', lastMsg.id);
		li.setAttribute('data-start', lastMsg.sub_start);
		li.setAttribute('data-end', lastMsg.sub_end);
		ul.prepend(li);
	}

	latest.innerText = msg.line;
	latest.setAttribute('data-id', msg.id)
	latest.setAttribute('data-start', msg.sub_start)
	latest.setAttribute('data-end', msg.sub_end)
};
`
