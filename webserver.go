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

package main

import (
	"net/http"
	"fmt"
	"code.google.com/p/go.net/websocket"
)

type HalmaMessageType int

const (
	MsgRegister HalmaMessageType = iota
	MsgLogin
	MsgNewGame
	MsgJoinGame
	MsgChangeGame
	MsgMove
	MsgGameInfo
	MsgTurnInfo
	MsgFieldInfo
)

type PlayerMessage struct {
	Name string
	Password string
	OK bool
}

type TurnMessage struct {
	CurrentPlayer HalmaColor
}

type MoveMessage struct {
	From Position
	To Position
}

type HalmaMessage struct {
	Type HalmaMessageType
	Player *PlayerMessage
	Fields []FieldInfo
	Games []GameInfo
	Turn *TurnMessage
	Move *MoveMessage
	Game *GameInfo
}

type FieldInfo struct {
	Pos Position
	Pin HalmaColor
}

type GameInfo struct {
	ID int
	Player HalmaColor
	CurrentPlayer HalmaColor
}

type HalmaWebsocketClient struct {
	Connection *websocket.Conn
	Player *HalmaGamePlayer
}

func (c *HalmaWebsocketClient) NotifyMove(color HalmaColor, from Position, to Position) {
	var message HalmaMessage
	
	message.Type = MsgFieldInfo
	message.Fields = make([]FieldInfo, 2)
	message.Fields[0].Pos.X = from.X
	message.Fields[0].Pos.Y = from.Y
	message.Fields[0].Pin = Empty
	message.Fields[1].Pos.X = to.X
	message.Fields[1].Pos.Y = to.Y
	message.Fields[1].Pin = color

	websocket.JSON.Send(c.Connection, &message)
}

func (c *HalmaWebsocketClient) NotifyTurn(current HalmaColor) {
	var message HalmaMessage
	message.Type = MsgTurnInfo
	message.Turn = &TurnMessage{current}
	websocket.JSON.Send(c.Connection, &message)
}

func NewHalmaGame() *HalmaGame {
	g := &HalmaGame{len(gameList), New, [17][17]*HalmaField{}, Empty, nil, nil}
	gameList = append(gameList, g)

	return g
}

func RegisterPlayer(name string, password string) *Player {
	for _, player := range(playerList) {
		if player.Name == name {
			return nil
		}
	}
	
	player := &Player{name, password}
	playerList = append(playerList, player)
	return player
}

func LoginPlayer(name string, password string) *Player {
	for _, player := range(playerList) {
		if player.Name == name && player.Password == password {
			return player
		}
	}

	return nil
}

func DestroyClient(client *HalmaWebsocketClient) {
	if client != nil && client.Player != nil {
		game := client.Player.Game
		for i := range(game.Clients) {
			if game.Clients[i] == client {
				game.Clients[i], game.Clients = game.Clients[len(game.Clients)-1], game.Clients[:len(game.Clients)-1]
				break
			}
		}
	}
}

func process(ws *websocket.Conn) {
	var player *Player = nil
	var client *HalmaWebsocketClient = nil
	var game *HalmaGame = nil
	
	//fmt.Println("Someone connected!")
	
	for {
		var message HalmaMessage
		err := websocket.JSON.Receive(ws, &message)
		
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
		
		//fmt.Println("Received message!")
		
		switch(message.Type) {
			case MsgMove:
				if client != nil && client.Player != nil {
					game.Move(client.Player.Color, message.Move.From, message.Move.To)
				}
				break
			case MsgFieldInfo:
				var response HalmaMessage
				response.Type = MsgFieldInfo
				response.Fields = make([]FieldInfo, 0)
				for x := 0; x < len(game.Fields); x++ {
					for y := 0; y < len(game.Fields[x]); y++ {
						if game.Fields[x][y] != nil {
							response.Fields = append(response.Fields, FieldInfo{Position{x - 8, y - 8}, game.Fields[x][y].Pin})
						}
					}
				}
				websocket.JSON.Send(ws, &response)
				break
			case MsgRegister:
				if message.Player != nil {
					player = RegisterPlayer(message.Player.Name, message.Player.Password)
					var response HalmaMessage
					response.Type = MsgLogin
					response.Player = &PlayerMessage{message.Player.Name, "", player != nil}
					websocket.JSON.Send(ws, &response)
				}
				break
			case MsgLogin:
				if message.Player != nil {
					player = LoginPlayer(message.Player.Name, message.Player.Password)
					var response HalmaMessage
					response.Type = MsgLogin
					response.Player = &PlayerMessage{message.Player.Name, "", player != nil}
					websocket.JSON.Send(ws, &response)
				}
				break
			case MsgGameInfo:
				var response HalmaMessage
				response.Type = MsgGameInfo
				response.Games = make([]GameInfo, len(gameList))
				for i, game := range(gameList) {
					response.Games[i].ID = game.ID
					response.Games[i].Player = game.GetPlayerColor(player)
					response.Games[i].CurrentPlayer = game.CurrentPlayer
				}
				websocket.JSON.Send(ws, &response)
				break
			case MsgChangeGame:
				if message.Game == nil || message.Game.ID < 0 || message.Game.ID >= len(gameList) {
					break
				}
				
				game = gameList[message.Game.ID]
				
				hgp := game.GetPlayer(player)
				
				DestroyClient(client)
				client = &HalmaWebsocketClient{ws, hgp}
				game.Clients = append(game.Clients, client)
				
				var response HalmaMessage
				response.Type = MsgChangeGame
				if hgp == nil {
					response.Game = &GameInfo{game.ID, Empty, game.CurrentPlayer}
				} else {
					if game.State == New {
						game.Start()
					}
					
					response.Game = &GameInfo{game.ID, hgp.Color, game.CurrentPlayer}
				}
				websocket.JSON.Send(ws, &response)
				break
			case MsgTurnInfo:
				var response HalmaMessage
				response.Type = MsgTurnInfo
				response.Turn = &TurnMessage{game.CurrentPlayer}
				websocket.JSON.Send(ws, &response)
				break
			case MsgNewGame:
				game := NewHalmaGame()
				game.Join(player)
				
				var response HalmaMessage
				response.Type = MsgGameInfo
				response.Games = []GameInfo{GameInfo{game.ID, game.GetPlayerColor(player), game.CurrentPlayer}}
				websocket.JSON.Send(ws, &response)
				break
			case MsgJoinGame:
				if message.Game == nil || message.Game.ID < 0 || message.Game.ID >= len(gameList) {
					break
				}
				
				game := gameList[message.Game.ID]
				
				if game.State != New {
					break
				}
				
				hgp := game.Join(player)
				
				var response HalmaMessage
				response.Type = MsgGameInfo
				response.Games = make([]GameInfo, 1)
				response.Games[0].ID = game.ID
				response.Games[0].Player = hgp.Color
				response.Games[0].CurrentPlayer = game.CurrentPlayer
				websocket.JSON.Send(ws, &response)
				break
		}
	}
	
	DestroyClient(client)
	
	fmt.Println("Closing socket")
}

var gameList []*HalmaGame
var playerList []*Player

func main() {
	gameList = make([]*HalmaGame, 0)
	playerList = make([]*Player, 0)
	
	http.Handle("/websocket/", websocket.Handler(process))
	http.Handle("/", http.FileServer(http.Dir(".")))
	err := http.ListenAndServe(":8000", nil)
	
	if err != nil {
		fmt.Println(err)
	}
}
