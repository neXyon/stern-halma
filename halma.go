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
	"math/rand"
)

type HalmaColor int

const (
	EMPTY HalmaColor = iota
	RED
	GREEN
	BLUE
)

type HalmaGameState int

const (
	NEW HalmaGameState = iota
	RUNNING
	DONE
)

type HalmaField struct {
	Type HalmaColor
	Pin HalmaColor
}

type Player struct {
	Name string
	Password string
}

type HalmaGame struct {
	ID int
	State HalmaGameState
	Fields [17][17]*HalmaField
	CurrentPlayer HalmaColor
	Players []*HalmaGamePlayer
	Clients []HalmaClient
}

type HalmaGamePlayer struct {
	Player *Player
	Game *HalmaGame
	Color HalmaColor
}

type HalmaClient interface {
	NotifyMove(color HalmaColor, from Position, to Position)
	NotifyTurn(current HalmaColor)
}

type Position struct {
	X int
	Y int
}

func abs(x int) int {
	if(x < 0) {
		return -x
	}
	return x
}

func (g *HalmaGame) createField(pos Position, Type HalmaColor, Pin HalmaColor) {
	g.Fields[pos.X + 8][pos.Y + 8] = &HalmaField{Type, Pin}
}

func (g *HalmaGame) fieldValid(pos Position) bool {
	return abs(pos.X) <= 8 && abs(pos.Y) <= 8 && g.getField(pos) != nil
}

func (g *HalmaGame) getField(pos Position) *HalmaField {
	return g.Fields[pos.X + 8][pos.Y + 8]
}

func (g *HalmaGame) calculatePossible(position Position) []Position {
	var visited [17][17]bool
	
	neighbors := []Position{
		{1, 0}, {1, -1}, {0, -1},
		{-1, 0}, {-1, 1}, {0, 1},
	}
	
	todo := []Position{position}
	possible := make([]Position, 0)
	first := true
	
	var pos Position

	for len(todo) > 0 {
		pos, todo = todo[len(todo) - 1], todo[:len(todo) - 1]
		if visited[pos.X + 8][pos.Y + 8] {
			continue
		}
		
		if !first {
			possible = append(possible, pos)
		}
		visited[pos.X + 8][pos.Y + 8] = true
		
		for i := 0; i < len(neighbors); i++ {
			nb1 := Position{pos.X + neighbors[i].X, pos.Y + neighbors[i].Y}
			nb2 := Position{pos.X + 2 * neighbors[i].X, pos.Y + 2 * neighbors[i].Y}
			if g.fieldValid(nb1) && g.fieldValid(nb2) && g.getField(nb1).Pin != EMPTY && g.getField(nb2).Pin == EMPTY {
				todo = append(todo, nb2)
			}
			if first && g.fieldValid(nb1) && g.getField(nb1).Pin == EMPTY {
				possible = append(possible, nb1)
			}
		}
		
		first = false
	}
	
	return possible
}

func (g *HalmaGame) checkWinner(color HalmaColor) bool {
	for x := 0; x < len(g.Fields); x++ {
		for y := 0; y < len(g.Fields[x]); y++ {
			if g.Fields[x][y] != nil {
				if g.Fields[x][y].Type == color && g.Fields[x][y].Type != g.Fields[x][y].Pin {
					return false
				}
			}
		}
	}
	
	return true
}

func (g *HalmaGame) GetFreeColor() HalmaColor {
	color := HalmaColor(len(g.Players) + 1)
	if color > 3 {
		color = EMPTY
	}
	return color
}

func (g *HalmaGame) GetPlayerColor(player *Player) HalmaColor {
	hgp := g.GetPlayer(player)

	if hgp != nil {
		return hgp.Color
	}

	return EMPTY
}

func NewHalmaGame() *HalmaGame {
	g := &HalmaGame{len(gameList), NEW, [17][17]*HalmaField{}, EMPTY, nil, nil}
	gameList = append(gameList, g)

	return g
}

func (g *HalmaGame) Start() {
	green := EMPTY
	blue := EMPTY
	
	if len(g.Players) > 1 {
		green = GREEN
		
		if len(g.Players) > 2 {
			blue = BLUE
		}
	}
	
	for u := -8; u <= 8; u++ {
		for v := -8; v <= 8; v++ {
			tu := abs(u)
			tv := abs(v)
			tw := abs(-u - v)
			
			if (tu <= 8 && tv <= 4 && tw <= 4) || (tu <= 4 && tv <= 8 && tw <=4) || (tu <= 4 && tv <= 4 && tw <=8) {
				if (tu + tv + tw) / 2 < 4 {
					g.createField(Position{u, v}, EMPTY, EMPTY)
				} else if tv >= 4 && tv > tu && tv > tw  {
					if v > 0 {
						g.createField(Position{u, v}, EMPTY, RED)
					} else {
						g.createField(Position{u, v}, RED, EMPTY)
					}
				} else if tw >= 4 && tw > tv && tw > tu {
					if -u - v > 0 {
						g.createField(Position{u, v}, EMPTY, green)
					} else {
						g.createField(Position{u, v}, GREEN, EMPTY)
					}
				} else if tu >= 4 && tu > tv && tu > tw {
					if u > 0 {
						g.createField(Position{u, v}, EMPTY, blue)
					} else {
						g.createField(Position{u, v}, BLUE, EMPTY)
					}
				}
			}
		}
	}
	
	g.createField(Position{0, -4}, RED, green)
	g.createField(Position{-4, 0}, BLUE, green)
	g.createField(Position{-4, 4}, BLUE, RED)
	g.createField(Position{0, 4}, GREEN, RED)
	g.createField(Position{4, 0}, GREEN, blue)
	g.createField(Position{4, -4}, RED, blue)
	
	g.State = RUNNING
	g.CurrentPlayer = HalmaColor(rand.Intn(len(g.Players)) + 1)
}

type HalmaMessageType int

const (
	REGISTER HalmaMessageType = iota
	LOGIN
	NEW_GAME
	JOIN_GAME
	CHANGE_GAME
	MOVE
	GAME_INFO
	TURN_INFO
	FIELD_INFO
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

func (g *HalmaGame) Move(color HalmaColor, from Position, to Position) {
	if color != g.CurrentPlayer || !g.fieldValid(from) || !g.fieldValid(to) || g.getField(from).Pin != color {
		return
	}

	possible := g.calculatePossible(from)
	
	for _, pos := range(possible) {
		if pos.X == to.X && pos.Y == to.Y {
			g.Fields[to.X + 8][to.Y + 8].Pin = g.Fields[from.X + 8][from.Y + 8].Pin
			g.Fields[from.X + 8][from.Y + 8].Pin = EMPTY
			
			if g.checkWinner(g.CurrentPlayer) {
				g.CurrentPlayer = EMPTY
			} else {
				g.CurrentPlayer++
				if int(g.CurrentPlayer) > len(g.Players) {
					g.CurrentPlayer = RED
				}
			}
			
			for _, client := range(g.Clients) {
				client.NotifyMove(color, from, to)
				client.NotifyTurn(g.CurrentPlayer)
			}
		}
	}
}

func (g *HalmaGame) Join(player *Player) *HalmaGamePlayer {
	color := g.GetFreeColor()
	
	if color == EMPTY {
		return nil
	}
	
	hgp := &HalmaGamePlayer{player, g, color}
	g.Players = append(g.Players, hgp)
	
	return hgp
}

func (g *HalmaGame) GetPlayer(player *Player) *HalmaGamePlayer {
	for _, hgp := range(g.Players) {
		if hgp.Player == player {
			return hgp
		}
	}
	
	return nil
}

func NewPlayer(name string, password string) *Player {
	return &Player{name, password}
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
			case MOVE:
				if client != nil && client.Player != nil {
					game.Move(client.Player.Color, message.Move.From, message.Move.To)
				}
				break
			case FIELD_INFO:
				var response HalmaMessage
				response.Type = FIELD_INFO
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
			case REGISTER:
				if message.Player != nil {
					player = RegisterPlayer(message.Player.Name, message.Player.Password)
					var response HalmaMessage
					response.Type = LOGIN
					response.Player = &PlayerMessage{message.Player.Name, "", player != nil}
					websocket.JSON.Send(ws, &response)
				}
				break
			case LOGIN:
				if message.Player != nil {
					player = LoginPlayer(message.Player.Name, message.Player.Password)
					var response HalmaMessage
					response.Type = LOGIN
					response.Player = &PlayerMessage{message.Player.Name, "", player != nil}
					websocket.JSON.Send(ws, &response)
				}
				break
			case GAME_INFO:
				var response HalmaMessage
				response.Type = GAME_INFO
				response.Games = make([]GameInfo, len(gameList))
				for i, game := range(gameList) {
					response.Games[i].ID = game.ID
					response.Games[i].Player = game.GetPlayerColor(player)
					response.Games[i].CurrentPlayer = game.CurrentPlayer
				}
				websocket.JSON.Send(ws, &response)
				break
			case CHANGE_GAME:
				if message.Game == nil || message.Game.ID < 0 || message.Game.ID >= len(gameList) {
					break
				}
				
				game = gameList[message.Game.ID]
				
				hgp := game.GetPlayer(player)
				
				DestroyClient(client)
				client = &HalmaWebsocketClient{ws, hgp}
				game.Clients = append(game.Clients, client)
				
				var response HalmaMessage
				response.Type = CHANGE_GAME
				if hgp == nil {
					response.Game = &GameInfo{game.ID, EMPTY, game.CurrentPlayer}
				} else {
					if game.State == NEW {
						game.Start()
					}
					
					response.Game = &GameInfo{game.ID, hgp.Color, game.CurrentPlayer}
				}
				websocket.JSON.Send(ws, &response)
				break
			case TURN_INFO:
				var response HalmaMessage
				response.Type = TURN_INFO
				response.Turn = &TurnMessage{game.CurrentPlayer}
				websocket.JSON.Send(ws, &response)
				break
			case NEW_GAME:
				game := NewHalmaGame()
				game.Join(player)
				
				var response HalmaMessage
				response.Type = GAME_INFO
				response.Games = []GameInfo{GameInfo{game.ID, game.GetPlayerColor(player), game.CurrentPlayer}}
				websocket.JSON.Send(ws, &response)
				break
			case JOIN_GAME:
				if message.Game == nil || message.Game.ID < 0 || message.Game.ID >= len(gameList) {
					break
				}
				
				game := gameList[message.Game.ID]
				
				if game.State != NEW {
					break
				}
				
				hgp := game.Join(player)
				
				var response HalmaMessage
				response.Type = GAME_INFO
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

type HalmaWebsocketClient struct {
	Connection *websocket.Conn
	Player *HalmaGamePlayer
}

func (c *HalmaWebsocketClient) NotifyMove(color HalmaColor, from Position, to Position) {
	var message HalmaMessage
	
	message.Type = FIELD_INFO
	message.Fields = make([]FieldInfo, 2)
	message.Fields[0].Pos.X = from.X
	message.Fields[0].Pos.Y = from.Y
	message.Fields[0].Pin = EMPTY
	message.Fields[1].Pos.X = to.X
	message.Fields[1].Pos.Y = to.Y
	message.Fields[1].Pin = color

	websocket.JSON.Send(c.Connection, &message)
}

func (c *HalmaWebsocketClient) NotifyTurn(current HalmaColor) {
	var message HalmaMessage
	message.Type = TURN_INFO
	message.Turn = &TurnMessage{current}
	websocket.JSON.Send(c.Connection, &message)
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
