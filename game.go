package main

import (
	"bytes"
	"fmt"
	"math/rand"

	"github.com/jaysinco/Tools/core"
)

func NewGame() *State {
	return &State{new([_row][_col]Symbol), nil}
}

type State struct {
	chessBoard *[_row][_col]Symbol
	lastMove   *Pos
}

func (s *State) Nextp(move *Pos) {
	if !s.IsValidMove(move) {
		core.Fatal("invalid move: %v", move)
	}
	nowPlayer := s.LastPlayer().Flip()
	s.chessBoard[move.Row][move.Col] = nowPlayer
	s.lastMove = move
}

func (s *State) Next(row, col int) {
	s.Nextp(&Pos{row, col})
}

func (s *State) NextRandom() {
	s.Nextp(s.GetMoves()[0])
}

func (s *State) IsValidMove(move *Pos) bool {
	r, c := move.Row, move.Col
	return r < _row && r >= 0 && c < _col && c >= 0 && s.Get(r, c) == ZERO
}

func (s *State) Clone() *State {
	cloned, board, last := new(State), new([_row][_col]Symbol), (*Pos)(nil)
	*board = *(s.chessBoard)
	if s.lastMove != nil {
		last = new(Pos)
		*last = *(s.lastMove)
	}
	cloned.chessBoard, cloned.lastMove = board, last
	return cloned
}

func (s *State) Get(row, col int) Symbol {
	return s.chessBoard[row][col]
}

func (s *State) Getp(p *Pos) Symbol {
	return s.chessBoard[p.Row][p.Col]
}

func (s *State) LastPlayer() Symbol {
	if s.lastMove == nil {
		return TWO
	}
	return s.Getp(s.lastMove)
}

func (s *State) LastMove() *Pos {
	return s.lastMove
}

func (s *State) GetMoves() (options []*Pos) {
	for r := 0; r < _row; r++ {
		for c := 0; c < _col; c++ {
			if s.Get(r, c) == ZERO {
				options = append(options, &Pos{r, c})
			}
		}
	}
	for i := 0; i < len(options); i++ {
		a := rand.Intn(len(options))
		b := rand.Intn(len(options))
		options[a], options[b] = options[b], options[a]
	}
	return
}

func (s *State) Winner() (player Symbol, over bool) {
	round := 0
	for rx := 0; rx < _row; rx++ {
		for cy := 0; cy < _col; cy++ {
			sym := s.Get(rx, cy)
			if sym != ZERO {
				round++
				for _, dir := range [4][2]int{{0, 1}, {1, 0}, {-1, 1}, {1, 1}} {
					total, dr, dc := 0, dir[0], dir[1]
					for _, m := range [2]int{1, -1} {
						for r, c := rx, cy; r < _row && r >= 0 && c < _col && c >= 0 && s.Get(r, c) == sym; {
							total++
							r += dr * m
							c += dc * m
						}
					}
					if total-1 >= _suc {
						return sym, true
					}
				}
			}
		}
	}
	if round >= _row*_col {
		return ZERO, true
	}
	return ZERO, false
}

func (s *State) String() string {
	var buf bytes.Buffer
	for r := 0; r < _row; r++ {
		buf.WriteString(fmt.Sprintf("%2d|", (r+1)%10))
		for c := 0; c < _col; c++ {
			buf.WriteString(s.Get(r, c).String() + "|")
		}
		buf.WriteRune('\n')
	}
	buf.WriteString("   ")
	for c := 0; c < _col; c++ {
		buf.WriteString(fmt.Sprintf("%-2d", (c+1)%10))
	}
	buf.WriteRune('\n')
	buf.WriteString(fmt.Sprintf("last move: %s%v", s.LastPlayer(), s.LastMove()))
	return buf.String()
}

type Pos struct {
	Row int
	Col int
}

func (p *Pos) String() string {
	return fmt.Sprintf("(%02d, %02d)", p.Row+1, p.Col+1)
}

const (
	ZERO Symbol = iota
	ONE
	TWO
)

type Symbol int8

func (s Symbol) Flip() Symbol {
	if s == ZERO {
		core.Fatal("can not flip non-player chess peice")
	}
	return Symbol(3 - s)
}

func (s Symbol) String() string {
	switch s {
	case ZERO:
		return " "
	case ONE:
		return "●"
	case TWO:
		return "○"
	default:
		core.Fatal("impossible chess piece symbol: %v", s)
	}
	return "can not reach here"
}
