// Copyright (c) 2013 by Jörg Müller <nexyon@gmail.com> All rights reserved.
// 
// This file is part of stern-halma.
// 
// stern-halma is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// 
// stern-halma is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with stern-halma.  If not, see <http://www.gnu.org/licenses/>.

var radius = 24;
var drawradius = 6;

var EMPTY = 0;
var RED = 1;
var GREEN = 2;
var BLUE = 3;

var REGISTER = 0;
var LOGIN = 1;
var NEW_GAME = 2;
var JOIN_GAME = 3;
var CHANGE_GAME = 4;
var MOVE = 5;
var GAME_INFO = 6;
var TURN_INFO = 7;
var FIELD_INFO = 8;

var joinGameF;
var changeGameF;

window.onload = function () {
	var grab = {
		mx: 0,
		my: 0,
		drag: false,
		orig: {X: 0, Y:0},
		offsetx: 0,
		offsety: 0,
		pin: EMPTY
	};
	
	function drawCircle(ctx, pos, color) {
		ctx.beginPath();
		ctx.fillStyle = color;
		ctx.lineWidth = 2;
		ctx.arc(pos.X, pos.Y, drawradius, 0, 2 * Math.PI);
		ctx.fill();
		ctx.stroke();
	}
	
	function drawLine(ctx, p1, p2) {
		ctx.beginPath();
		ctx.lineWidth = 2;
		ctx.moveTo(p1.X, p1.Y);
		ctx.lineTo(p2.X, p2.Y);
		ctx.stroke();
	}
	
	function drawTriangle(ctx, p1, p2, p3, color) {
		p1 = gridToPixel(p1);
		p2 = gridToPixel(p2);
		p3 = gridToPixel(p3);
		
		ctx.beginPath();
		ctx.fillStyle = color;
		ctx.moveTo(p1.X, p1.Y);
		ctx.lineTo(p2.X, p2.Y);
		ctx.lineTo(p3.X, p3.Y);
		ctx.closePath();
		ctx.fill();
	}
	
	function dataToGrid(pos) {
		return { X: pos.X - 8, Y: pos.Y - 8 };
	}
	
	function gridToData(pos) {
		return { X: pos.X + 8, Y: pos.Y + 8 };
	}
	
	function gridToPixel(pos) {
		return { X: radius * Math.sqrt(3) * (pos.X + pos.Y / 2 + 6) + drawradius + 1,
		         Y: radius * 3 / 2 * (pos.Y + 8) + drawradius + 1 };
	}
	
	function pixelToGrid(pos) {
		var y = (pos.Y - drawradius - 1) * 2 / 3 / radius - 8;
		var x = (pos.X - drawradius - 1) / Math.sqrt(3) / radius - 6 - y / 2;
		var z = -x - y;
		
		var rx = Math.round(x);
		var ry = Math.round(y);
		var rz = Math.round(z);
		
		x = Math.abs(rx - x);
		y = Math.abs(ry - y);
		z = Math.abs(rz - z);
		
		if(x > y && x > z)
			rx = -ry - rz;
		else if(y > z)
			ry = -rx - rz;
		
		return { X: rx, Y: ry };
	}
	
	function drawCanvas() {
		var canvas = document.getElementById("board");
		var ctx = canvas.getContext("2d");
		ctx.clearRect(0, 0, canvas.width, canvas.height);
		
		drawTriangle(ctx, {X: 4, Y: -8}, {X: 0, Y: -4}, {X: 4, Y: -4}, "#FF0000");
		drawTriangle(ctx, {X: -4, Y: 4}, {X: 0, Y: 4}, {X: -4, Y: 8}, "#FF0000");
		drawTriangle(ctx, {X: -4, Y: -4}, {X: 0, Y: -4}, {X: -4, Y: 0}, "#00FF00");
		drawTriangle(ctx, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}, "#00FF00");
		drawTriangle(ctx, {X: 4, Y: -4}, {X: 8, Y: -4}, {X: 4, Y: 0}, "#0000FF");
		drawTriangle(ctx, {X: -4, Y: 0}, {X: -8, Y: 4}, {X: -4, Y: 4}, "#0000FF");
		
		for(var z = -7; z <= 7; z++) {
			var tz = Math.abs(z);
			
			var count = 8 + tz;
			var from = -4 - z;
			
			if(tz > 4) {
				count = 8 - tz;
			}
			
			if((tz > 4) ^ (z != tz)) {
				from = -4;
			}
			
			var p1 = gridToPixel({X: from, Y: z});
			var p2 = gridToPixel({X: from + count, Y: z});
			
			drawLine(ctx, p1, p2);
			
			p1 = gridToPixel({X: z, Y: -z - from});
			p2 = gridToPixel({X: z, Y: -z - from - count});
			
			drawLine(ctx, p1, p2);
			
			p1 = gridToPixel({X: -z - from, Y: from});
			p2 = gridToPixel({X: -z - from - count, Y: from + count});
			
			drawLine(ctx, p1, p2);
		}
		
		for(var u = -8; u <= 8; u++) {
			for(var v = -8; v <= 8; v++) {
				var tu = Math.abs(u);
				var tv = Math.abs(v);
				var tw = Math.abs(-u - v);
				
				if((tu <= 8 && tv <= 4 && tw <= 4) || (tu <= 4 && tv <= 8 && tw <=4) || (tu <= 4 && tv <= 4 && tw <=8)) {
					if((g = getField({X: u, Y: v})) != null) {
						if(g.pin == RED)
							drawCircle(ctx, gridToPixel({X: u, Y: v}), "#FF0000");
						else if(g.pin == GREEN)
							drawCircle(ctx, gridToPixel({X: u, Y: v}), "#00FF00");
						else if(g.pin == BLUE)
							drawCircle(ctx, gridToPixel({X: u, Y: v}), "#0000FF");
						/*else if((tu + tv + tw) / 2 < 4)
							drawCircle(ctx, gridToPixel({X: u, Y: v}), "#FFFF00");*/
						else
							drawCircle(ctx, gridToPixel({X: u, Y: v}), "#000000");
					}
				}
			}
		}
		
		for(var i = 0; i < possible.length; i++) {
			drawCircle(ctx, gridToPixel(possible[i]), "#FFFF00");
		}
		
		if(grab.drag) {
			var color = "#0000FF";
			if(grab.pin == RED)
				color = "#FF0000";
			else if(grab.pin == GREEN)
				color = "#00FF00";
			drawCircle(ctx, {X: grab.mx + grab.offsetx, Y: grab.my + grab.offsety}, color);
		}
	}
	
	function onMouseMove(event) {
		if(grab.drag) {
			var rect = canvas.getBoundingClientRect();
			
			pos = gridToPixel(pixelToGrid({X: event.clientX - rect.left, Y: event.clientY - rect.top}));
			grab.mx = event.clientX - rect.left;
			grab.my = event.clientY - rect.top;
			
			drawCanvas();
		}
	}
	
	function onMouseDown(event) {
		var rect = canvas.getBoundingClientRect();
		grab.mx = event.clientX - rect.left;
		grab.my = event.clientY - rect.top;
		pos = pixelToGrid({X: grab.mx, Y: grab.my});
		if(fieldValid(pos)) {
			pos = gridToData(pos);
			grab.orig = pos;
			if(fields[pos.X][pos.Y].pin != EMPTY) {
				grab.pin = fields[pos.X][pos.Y].pin;
				pos = gridToPixel(dataToGrid(pos));
				grab.offsetx = pos.X - grab.mx;
				grab.offsety = pos.Y - grab.my;
				if(grab.offsetx * grab.offsetx + grab.offsety * grab.offsety <= drawradius * drawradius) {
					fields[grab.orig.X][grab.orig.Y].pin = EMPTY;
					grab.drag = true;
					calculatePossible(dataToGrid(grab.orig));
					drawCanvas();
				}
			}
		}
	}
	
	function onMouseUp(event) {
		if(grab.drag) {
			grab.drag = false;
			possible = [];
			var worked = false;
			
			var rect = canvas.getBoundingClientRect();
			grab.mx = event.clientX - rect.left;
			grab.my = event.clientY - rect.top;
			pos = pixelToGrid({X: grab.mx + grab.offsetx, Y: grab.my + grab.offsety});
			
			if(fieldValid(pos) && current == me && grab.pin == me) {
				pos = gridToData(pos);
				if(fields[pos.X][pos.Y].pin == EMPTY) {
					//fields[pos.X][pos.Y].pin = grab.pin;
					fields[grab.orig.X][grab.orig.Y].pin = grab.pin;
					worked = true;
					move(grab.orig, pos);
				}
			}
			
			if(!worked)
				fields[grab.orig.X][grab.orig.Y].pin = grab.pin;
			
			drawCanvas();
		}
	}
	
	function getField(pos) {
		return fields[pos.X + 8][pos.Y + 8];
	}
	
	function setField(x, y, value) {
		fields[x + 8][y + 8] = value;
	}
	
	function createField(fields, x, y, type, pin) {
		fields[x + 8][y + 8] = {type: type, pin: pin};
	}
	
	function fieldValid(pos) {
		return Math.abs(pos.X) <= 8 && Math.abs(pos.Y) <= 8 && getField(pos) != null;
	}
	
	function calculatePossible(pos) {
		visited = [];
		for(var i = 0; i < 17; i++) {
			visited[i] = [];
			
			for(var j = 0; j < 17; j++)
				visited[i][j] = false;
		}
		
		var neighbors = [
			[+1,  0],  [+1, -1],  [ 0, -1],
			[-1,  0],  [-1, +1],  [ 0, +1] 
		];
				
		var todo = [pos];
		var first = true;

		while(todo.length > 0) {
			pos = todo.shift();
			if(visited[pos.X + 8][pos.Y + 8])
				continue;
			
			if(!first)
				possible.push(pos);
			visited[pos.X + 8][pos.Y + 8] = true;
			
			for(var i = 0; i < neighbors.length; i++) {
				var nb1 = {X: pos.X + neighbors[i][0], Y: pos.Y + neighbors[i][1]};
				var nb2 = {X: pos.X + 2 * neighbors[i][0], Y: pos.Y + 2 * neighbors[i][1]};
				if(fieldValid(nb1) && fieldValid(nb2) && getField(nb1).pin != EMPTY && getField(nb2).pin == EMPTY) {
					todo.push(nb2);
				}
				if(first && fieldValid(nb1) && getField(nb1).pin == EMPTY) {
					possible.push(nb1);
				}
			}
			
			first = false;
		}
	}
	
	function generateFields() {
		var fields = new Array(17);
		for(var i = 0; i < 17; i++)
			fields[i] = new Array(17);

		for(var u = -8; u <= 8; u++) {
			for(var v = -8; v <= 8; v++) {
				var tu = Math.abs(u);
				var tv = Math.abs(v);
				var tw = Math.abs(-u - v);
				
				if((tu <= 8 && tv <= 4 && tw <= 4) || (tu <= 4 && tv <= 8 && tw <=4) || (tu <= 4 && tv <= 4 && tw <=8)) {
					if((tu + tv + tw) / 2 < 4) {
						createField(fields, u, v, EMPTY, EMPTY);
					}
					else if(tv >= 4 && tv > tu && tv > tw) {
						if(v > 0)
							createField(fields, u, v, EMPTY, EMPTY);//RED);
						else
							createField(fields, u, v, RED, EMPTY);
					}
					else if(tw >= 4 && tw > tv && tw > tu) {
						if(-u - v > 0)
							createField(fields, u, v, EMPTY, EMPTY);//, GREEN);
						else
							createField(fields, u, v, GREEN, EMPTY);
					}
					else if(tu >= 4 && tu > tv && tu > tw) {
						if(u > 0)
							createField(fields, u, v, EMPTY, EMPTY);//, BLUE);
						else
							createField(fields, u, v, BLUE, EMPTY);
					}
				}
			}
		}
		
		createField(fields, 0, -4, RED, EMPTY);//, GREEN);
		createField(fields, -4, 0, BLUE, EMPTY);//, GREEN);
		createField(fields, -4, 4, BLUE, EMPTY);//, RED);
		createField(fields, 0, 4, GREEN, EMPTY);//, RED);
		createField(fields, 4, 0, GREEN, EMPTY);//, BLUE);
		createField(fields, 4, -4, RED, EMPTY);//, BLUE);
		
		return fields;
	}
	
	var fields = generateFields();
	
	var current = RED;
	
	var me = EMPTY;
	
	var possible = [];
	
	drawCanvas();
	
	canvas = document.getElementById("board");
	
	canvas.addEventListener("mousemove", onMouseMove);
	canvas.addEventListener("mousedown", onMouseDown);
	canvas.addEventListener("mouseup", onMouseUp);

	var connection = {
		connected: false,
		socket: null,
		
	};
	
	function colorToText(color) {
		switch(color) {
			case RED:
				return "red";
			case GREEN:
				return "green";
			case BLUE:
				return "blue";
			default:
				return "none";
		}
	}
	
	function onOpen(event) {
		console.log("Connected");
		connection.connected = true;
	}
	
	function onMessage(event) {
		message = JSON.parse(event.data);
		switch(message.Type) {
			case LOGIN:
				if(message.Player.OK) {
					document.getElementById("loginform").style.visibility = "hidden";
					sendMessage({Type: GAME_INFO});
					console.log("Login complete!");
				}
				else {
					// TODO: error message
					console.log("Login failed!");
				}
				break;
			case GAME_INFO:
				for(var i = 0; i < message.Games.length; i++) {
					var element = document.getElementById("game" + message.Games[i].ID);
					
					if(element == null) {
						element = document.createElement("tr");
						element.id = "game" + message.Games[i].ID;
						var table = document.getElementById("games");
						table.appendChild(element);
					}
					
					var newrow = "<td><a href=\"javascript:changeGameF(" + message.Games[i].ID + ");\">" + message.Games[i].ID + "</a></td><td>" + colorToText(message.Games[i].CurrentPlayer) + "</td><td>";
					
					if(message.Games[i].Player == EMPTY) {
						newrow += "<a href=\"javascript:joinGameF(" + message.Games[i].ID + ");\">" + colorToText(message.Games[i].Player) + "</a></td>";
					}
					else {
						newrow += colorToText(message.Games[i].Player) + "</td>";
					}
					
					element.innerHTML = newrow;
				}
				break;
			case CHANGE_GAME:
				sendMessage({Type: FIELD_INFO});
				me = message.Game.Player;
				document.getElementById("me").innerHTML = colorToText(me);
				current = message.Game.CurrentPlayer;
				document.getElementById("current").innerHTML = colorToText(current);
				break;
			case FIELD_INFO:
				for(var i = 0; i < message.Fields.length; i++) {
					pos = gridToData(message.Fields[i].Pos);
					fields[pos.X][pos.Y].pin = message.Fields[i].Pin;
				}
				drawCanvas();
				break;
			case TURN_INFO:
				current = message.Turn.CurrentPlayer;
				document.getElementById("current").innerHTML = colorToText(current);
				break;
			case MOVE:
				fields[message.Move.To.X + 8][message.Move.To.Y + 8].pin = fields[message.Move.From.X + 8][message.Move.From.Y + 8].pin;
				fields[message.Move.From.X + 8][message.Move.From.Y + 8].pin = EMPTY;
				drawCanvas();
				break;
		}
	}
	
	function onClose(event) {
		connection.connected = false;
		console.log("Closed");
	}
	
	function onError(event) {
		disconnect();
	}
	
	function connect() {
		connection.socket = new WebSocket("ws://localhost:8000/websocket/ws");
		connection.socket.onopen = onOpen;
		connection.socket.onmessage = onMessage;
		connection.socket.onerror = onError;
		connection.socket.onclose = onClose;
		console.log("Connecting");
	}

	function disconnect() {
		console.log("Disconnecting");
		connection.socket.close();
		connection.connected = false;
	}
	
	function sendMessage(message) {
		if(connection.connected) {
			connection.socket.send(JSON.stringify(message));
		}
	}
	
	function move(from, to) {
		sendMessage({Type: MOVE, Move: {From: dataToGrid(from), To: dataToGrid(to)}});
	}
	
	function register() {
		var name = document.getElementById("name").value;
		var password = document.getElementById("password").value;
		sendMessage({Type: REGISTER, Player: {Name: name, Password: password}});
	}
	
	function login() {
		var name = document.getElementById("name").value;
		var password = document.getElementById("password").value;
		sendMessage({Type: LOGIN, Player: {Name: name, Password: password}});
	}
	
	function newGame() {
		sendMessage({Type: NEW_GAME});
	}
	
	function refresh() {
		document.getElementById("games").innerHTML = "<tr><td>Game</td><td>Current Turn</td><td>My Color</td></tr>";
		sendMessage({Type: GAME_INFO});
	}
	
	function joinGame(id) {
		sendMessage({Type: JOIN_GAME, Game: {ID: id}});
	}
	
	function changeGame(id) {
		sendMessage({Type: CHANGE_GAME, Game: {ID: id}});
	}
	
	joinGameF = joinGame;
	changeGameF = changeGame;
	
	document.getElementById("register").onclick = register;
	document.getElementById("login").onclick = login;
	document.getElementById("newgame").onclick = newGame;
	document.getElementById("refresh").onclick = refresh;
	
	connect();
};