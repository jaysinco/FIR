package main

const (
	_row = 7
	_col = 7
	_suc = 4
)

func main() {
	player1 := &Robot{10000}
	player2 := new(Human)
	//player2 := &Robot{10000}
	play(player1, player2)
}
