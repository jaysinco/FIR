package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	p1 := newRobot(0.1, 0.1, "")
	p2 := newRobot(0.1, 0.1, "")
	train(p1, p2, 5000, true)
	// p1.SavePolicy("black_3x3by3.plc")
	// p2.SavePolicy("white_3x3by3.plc")
	p1.ExploreRate = 0.0
	p2.ExploreRate = 0.0
	train(p1, p2, 1000, false)
	h1 := new(human)
	play(p1, h1, false, true)
}

const (
	_row = 3
	_col = 3
	_suc = 3
)

const (
	vacant symbol = iota
	black
	white
	even
)

var seed = rand.New(rand.NewSource(time.Now().Unix()))

type symbol int8

func (s symbol) String() string {
	switch s {
	case vacant:
		return " "
	case black:
		return "●"
	case white:
		return "○"
	case even:
		return "nobody"
	default:
		return "impossible"
	}
}

type state [_row][_col]symbol

func (s *state) Reset() {
	for r := 0; r < _row; r++ {
		for c := 0; c < _col; c++ {
			s[r][c] = vacant
		}
	}
}

func (s *state) MoveOptions(sym symbol) []move {
	options := make([]move, 0)
	for r := 0; r < _row; r++ {
		for c := 0; c < _col; c++ {
			if s[r][c] == vacant {
				options = append(options, move{sym, postion{r, c}})
			}
		}
	}
	n := len(options)
	for i := 0; i < n; i++ {
		a, b := seed.Intn(n), seed.Intn(n)
		options[a], options[b] = options[b], options[a]
	}
	return options
}

func (s *state) Winner() symbol {
	round := 0
	for rx := 0; rx < _row; rx++ {
		for cy := 0; cy < _col; cy++ {
			if s[rx][cy] != vacant {
				round++
				sym := s[rx][cy]
				for _, dir := range [4][2]int{{0, 1}, {1, 0}, {-1, 1}, {1, 1}} {
					total, dr, dc := 0, dir[0], dir[1]
					for _, m := range [2]int{1, -1} {
						for r, c := rx, cy; r < _row && r >= 0 && c < _col && c >= 0 && s[r][c] == sym; {
							total++
							r += dr * m
							c += dc * m
						}
					}
					if total-1 >= _suc {
						return sym
					}
				}
			}
		}
	}
	if round >= _row*_col {
		return even
	}
	return vacant
}

func (s *state) String() string {
	var buf bytes.Buffer
	for r := 0; r < _row; r++ {
		buf.WriteString(fmt.Sprintf("%2d|", (r+1)%10))
		for c := 0; c < _col; c++ {
			buf.WriteString(s[r][c].String() + "|")
		}
		buf.WriteRune('\n')
	}
	buf.WriteString("   ")
	for c := 0; c < _col; c++ {
		buf.WriteString(fmt.Sprintf("%-2d", (c+1)%10))
	}
	return buf.String()
}

type policy map[state]float64

func plc2str(policyFile string) string {
	p := newRobot(0.0, 0.0, policyFile)
	return p.Policy.String()
}

func (p policy) String() string {
	var buf bytes.Buffer
	for st, val := range p {
		for i := 0; i < _col*2+3; i++ {
			buf.WriteRune('-')
		}
		buf.WriteString(fmt.Sprintf(">> %.10f", val))
		buf.WriteRune('\n')
		buf.WriteString(st.String())
		buf.WriteRune('\n')
	}
	return buf.String()
}

type postion struct {
	Row int
	Col int
}

type move struct {
	Sym symbol
	Pos postion
}

type player interface {
	SetSymbol(sym symbol)
	PickMove(current state, options []move) move
	FeedReward(history []state, reward float64)
}

type robot struct {
	Symbol      symbol
	Policy      policy
	LearnSpeed  float64
	ExploreRate float64
}

func newRobot(learnSpeed, exploreRate float64, policyFile string) *robot {
	rbt := new(robot)
	rbt.LearnSpeed = learnSpeed
	rbt.ExploreRate = exploreRate
	if policyFile != "" {
		rbt.Policy = make(policy)
		if err := rbt.LoadPolicy(policyFile); err != nil {
			panic(err)
		}
	}
	return rbt
}

func (r *robot) SetSymbol(sym symbol) {
	r.Symbol = sym
	if r.Policy == nil {
		r.Policy = make(policy)
	}
}

func (r *robot) LoadPolicy(filename string) error {
	source, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open policy file: %v", err)
	}
	defer source.Close()
	decoder := gob.NewDecoder(source)
	if err = decoder.Decode(&r.Policy); err != nil {
		return fmt.Errorf("deserialize policy: %v", err)
	}
	return nil
}

func (r *robot) SavePolicy(filename string) error {
	dest, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create policy file: %v", err)
	}
	defer dest.Close()
	encoder := gob.NewEncoder(dest)
	if err := encoder.Encode(r.Policy); err != nil {
		return fmt.Errorf("serialize policy: %v", err)
	}
	return nil
}

func (r *robot) PickMove(current state, options []move) move {
	if seed.Float64() < r.ExploreRate {
		return options[0]
	}
	var best move
	maxValue := -1.0
	for _, mv := range options {
		future := current
		future[mv.Pos.Row][mv.Pos.Col] = r.Symbol
		value, ok := r.Policy[future]
		if !ok {
			switch future.Winner() {
			case r.Symbol:
				value = 1.0
			case vacant:
				value = 0.5
			default:
				value = 0.0
			}
			r.Policy[future] = value
		}
		if value > maxValue {
			best = mv
			maxValue = value
		}
	}
	return best
}

func (r *robot) FeedReward(history []state, reward float64) {
	target := reward
	for i := len(history) - 1; i >= 0; i-- {
		state := history[i]
		value, ok := r.Policy[state]
		if !ok {
			switch state.Winner() {
			case r.Symbol:
				value = 1.0
			case vacant:
				value = 0.5
			default:
				value = 0.0
			}
		}
		value += r.LearnSpeed * (target - value)
		r.Policy[state] = value
		target = value
	}
}

type human struct {
	Symbol symbol
}

func (h *human) SetSymbol(sym symbol) {
	h.Symbol = sym
}

func (h *human) PickMove(current state, options []move) move {
	var pos postion
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s(row, col)> ", h.Symbol)
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
		pos.Row = r - 1
		pos.Col = c - 1
		break
	}
	mv := move{h.Symbol, pos}
	for _, opt := range options {
		if opt == mv {
			return opt
		}
	}
	return h.PickMove(current, options)
}

func (h *human) FeedReward(history []state, reward float64) {
	return
}

type result struct {
	Rounds int
	Winner symbol
	Record []move
}

func play(offender, defender player, feedback bool, show bool) result {
	board := state{}
	sym, actor := black, offender
	round := 0
	winner := vacant
	record := make([]move, 0)
	history := make([]state, 0)
	offender.SetSymbol(black)
	defender.SetSymbol(white)
	for ; winner == vacant; winner = board.Winner() {
		if show {
			fmt.Println(&board)
		}
		mv := actor.PickMove(board, board.MoveOptions(sym))
		board[mv.Pos.Row][mv.Pos.Col] = mv.Sym
		record = append(record, mv)
		history = append(history, board)
		if sym == black {
			sym, actor = white, defender
		} else {
			sym, actor = black, offender
		}
		round++
	}
	history = append(history, board)
	if show {
		fmt.Printf("%s win in %d rounds!\n", winner, round)
		fmt.Println(&board)
	}
	if feedback {
		switch winner {
		case black:
			offender.FeedReward(history, 1.0)
			defender.FeedReward(history, 0.0)
		case white:
			offender.FeedReward(history, 0.0)
			defender.FeedReward(history, 1.0)
		case even:
			offender.FeedReward(history, 0.1)
			defender.FeedReward(history, 0.5)
		}
	}
	return result{Rounds: round, Winner: winner, Record: record}
}

func train(p1, p2 *robot, epoch int, feedback bool) {
	var p1w, p2w, p0w int
	for i := 0; i < epoch; i++ {
		result := play(p1, p2, feedback, false)
		switch result.Winner {
		case p1.Symbol:
			p1w++
		case p2.Symbol:
			p2w++
		case even:
			p0w++
		}
		fmt.Printf("\r%8d >> [black/%d] VS [white/%d] VS [even/%d]", i+1, p1w, p2w, p0w)
	}
	fmt.Print("\n")
}
