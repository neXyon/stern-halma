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
	"math/rand"
)

type HalmaColor int

const (
	Empty HalmaColor = iota
	Red
	Green
	Blue
)

type HalmaGameState int

const (
	New HalmaGameState = iota
	Running
	Done
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
			if g.fieldValid(nb1) && g.fieldValid(nb2) && g.getField(nb1).Pin != Empty && g.getField(nb2).Pin == Empty {
				todo = append(todo, nb2)
			}
			if first && g.fieldValid(nb1) && g.getField(nb1).Pin == Empty {
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
		color = Empty
	}
	return color
}

func (g *HalmaGame) GetPlayerColor(player *Player) HalmaColor {
	hgp := g.GetPlayer(player)

	if hgp != nil {
		return hgp.Color
	}

	return Empty
}

func (g *HalmaGame) Start() {
	green := Empty
	blue := Empty
	
	if len(g.Players) > 1 {
		green = Green
		
		if len(g.Players) > 2 {
			blue = Blue
		}
	}
	
	for u := -8; u <= 8; u++ {
		for v := -8; v <= 8; v++ {
			tu := abs(u)
			tv := abs(v)
			tw := abs(-u - v)
			
			if (tu <= 8 && tv <= 4 && tw <= 4) || (tu <= 4 && tv <= 8 && tw <=4) || (tu <= 4 && tv <= 4 && tw <=8) {
				if (tu + tv + tw) / 2 < 4 {
					g.createField(Position{u, v}, Empty, Empty)
				} else if tv >= 4 && tv > tu && tv > tw  {
					if v > 0 {
						g.createField(Position{u, v}, Empty, Red)
					} else {
						g.createField(Position{u, v}, Red, Empty)
					}
				} else if tw >= 4 && tw > tv && tw > tu {
					if -u - v > 0 {
						g.createField(Position{u, v}, Empty, green)
					} else {
						g.createField(Position{u, v}, Green, Empty)
					}
				} else if tu >= 4 && tu > tv && tu > tw {
					if u > 0 {
						g.createField(Position{u, v}, Empty, blue)
					} else {
						g.createField(Position{u, v}, Blue, Empty)
					}
				}
			}
		}
	}
	
	g.createField(Position{0, -4}, Red, green)
	g.createField(Position{-4, 0}, Blue, green)
	g.createField(Position{-4, 4}, Blue, Red)
	g.createField(Position{0, 4}, Green, Red)
	g.createField(Position{4, 0}, Green, blue)
	g.createField(Position{4, -4}, Red, blue)
	
	g.State = Running
	g.CurrentPlayer = HalmaColor(rand.Intn(len(g.Players)) + 1)
}

func (g *HalmaGame) Move(color HalmaColor, from Position, to Position) {
	if color != g.CurrentPlayer || !g.fieldValid(from) || !g.fieldValid(to) || g.getField(from).Pin != color {
		return
	}

	possible := g.calculatePossible(from)
	
	for _, pos := range(possible) {
		if pos.X == to.X && pos.Y == to.Y {
			g.Fields[to.X + 8][to.Y + 8].Pin = g.Fields[from.X + 8][from.Y + 8].Pin
			g.Fields[from.X + 8][from.Y + 8].Pin = Empty
			
			if g.checkWinner(g.CurrentPlayer) {
				g.CurrentPlayer = Empty
			} else {
				g.CurrentPlayer++
				if int(g.CurrentPlayer) > len(g.Players) {
					g.CurrentPlayer = Red
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
	
	if color == Empty {
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

