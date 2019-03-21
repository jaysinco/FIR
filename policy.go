package main

import (
	"fmt"
	"math"
	"sort"
)

type MCTSNode struct {
	LastMove     *Pos
	LastPlayer   Symbol
	Parent       *MCTSNode
	Childrens    []*MCTSNode
	Visited      int
	Value        float64
	UntriedMoves []*Pos
}

func (n *MCTSNode) NCTSelectChild() *MCTSNode {
	sort.Slice(n.Childrens, func(i, j int) bool {
		return n.Childrens[i].NCTRank() < n.Childrens[j].NCTRank()
	})
	return n.Childrens[len(n.Childrens)-1]
}

func (n *MCTSNode) NCTRank() float64 {
	fvisit := float64(n.Visited)
	return n.Value/fvisit + math.Sqrt(2*math.Log(fvisit)/fvisit)
}

func (n *MCTSNode) Update(winner Symbol) {
	n.Visited++
	switch winner {
	case ZERO:
		n.Value += 0.5
	case n.LastPlayer:
		n.Value += 1
	default:
		n.Value -= 1
	}
}

func (n *MCTSNode) PopTry() (move *Pos) {
	size := len(n.UntriedMoves)
	if size == 0 {
		return
	}
	move = n.UntriedMoves[size-1]
	n.UntriedMoves = n.UntriedMoves[:size-1]
	return
}

func (n *MCTSNode) AddChild(move *Pos, state *State) *MCTSNode {
	child := &MCTSNode{
		move, state.LastPlayer(), n, nil, 0, 0, state.GetMoves(),
	}
	n.Childrens = append(n.Childrens, child)
	return child
}

func (n *MCTSNode) String() string {
	return fmt.Sprintf("last move: %s/last player: %s/visited: %d/value: %.3f",
		n.LastMove, n.LastPlayer, n.Visited, n.Value)
}

func UCT(root *State, itermax int) (best *Pos) {
	rootNode := &MCTSNode{
		root.LastMove(), root.LastPlayer(), nil, nil, 0, 0, root.GetMoves(),
	}
	for i := 0; i < itermax; i++ {
		node := rootNode
		state := root.Clone()
		for len(node.UntriedMoves) == 0 && len(node.Childrens) > 0 {
			node = node.NCTSelectChild()
			state.Nextp(node.LastMove)
		}
		if len(node.UntriedMoves) > 0 {
			m := node.PopTry()
			state.Nextp(m)
			node = node.AddChild(m, state)
		}
		winner, over := state.Winner()
		for !over {
			state.NextRandom()
			winner, over = state.Winner()
		}
		for node != nil {
			node.Update(winner)
			node = node.Parent
		}
	}
	options := rootNode.Childrens
	sort.Slice(options, func(i, j int) bool {
		return options[i].Visited < options[j].Visited
	})
	return options[len(options)-1].LastMove
}
