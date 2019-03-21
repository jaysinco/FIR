package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func play(one Player, two Player) {
	game := NewGame()
	winner, over := game.Winner()
	for !over {
		switch game.LastPlayer() {
		case ONE:
			game.Nextp(two.TakeMove(game))
		case TWO:
			game.Nextp(one.TakeMove(game))
		}
		fmt.Println(game)
		winner, over = game.Winner()
	}
	if winner == ZERO {
		fmt.Printf("even!\n")
	} else {
		fmt.Printf("%s win!\n", winner)
	}
}

type Player interface {
	TakeMove(state *State) *Pos
}

type Human struct{}

func (h *Human) TakeMove(state *State) *Pos {
	var move Pos
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s(row, col)> ", state.LastPlayer().Flip())
		line, _, _ := reader.ReadLine()
		nums := strings.Split(string(line), ",")
		if len(nums) != 2 {
			if len(nums) == 1 && nums[0] == "q" {
				fmt.Println("quit!")
				os.Exit(0)
			}
			continue
		}
		r, err1 := strconv.Atoi(nums[0])
		c, err2 := strconv.Atoi(nums[1])
		if err1 != nil || err2 != nil {
			continue
		}
		move.Row = r - 1
		move.Col = c - 1
		break
	}
	if state.IsValidMove(&move) {
		return &move
	}
	return h.TakeMove(state)
}

type Robot struct {
	MCTSIterMax int
}

func (r *Robot) TakeMove(state *State) *Pos {
	return UCT(state, r.MCTSIterMax)
}
