package main

import (
	"encoding/gob"
	"log"
	"net"
	"os"

	"github.com/gdamore/tcell"
)

type GameState struct {
	Player1Y  int
	Player2Y  int
	BallX     int
	BallY     int
	BallVX    int
	BallVY    int
	PlayerNum int
}

var screen tcell.Screen

func drawPaddle(x, y int) {
	for i := 0; i < 4; i++ {
		screen.SetContent(x, y+i, '|', nil, tcell.StyleDefault)
	}
}

func drawBall(x, y int) {
	screen.SetContent(x, y, 'O', nil, tcell.StyleDefault)
}

func handleInput(conn net.Conn) {
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEsc, tcell.KeyCtrlC:
				screen.Fini()
				conn.Close()
				os.Exit(0)
			case tcell.KeyRune:
				var input string
				switch ev.Rune() {
				case 'w':
					input = "up"
				case 's':
					input = "down"
				case 'i':
					input = "up"
				case 'k':
					input = "down"
				}
				if input != "" {
					encoder := gob.NewEncoder(conn)
					if err := encoder.Encode(input); err != nil {
						log.Println("Error encoding input:", err)
					}
				}
			}
		}
	}
}

func drawScreen(gameState GameState) {
	screen.Clear()

	drawPaddle(0, gameState.Player1Y)
	drawPaddle(79, gameState.Player2Y)
	drawBall(gameState.BallX, gameState.BallY)
	screen.Show()
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatalf("Error creating screen: %v", err)
	}

	if err = screen.Init(); err != nil {
		log.Fatalf("Error initializing screen: %v", err)
	}

	defer screen.Fini()

	go handleInput(conn)

	decoder := gob.NewDecoder(conn)

	for {

		var gameState GameState
		if err := decoder.Decode(&gameState); err != nil {
			log.Println("Error decoding game state:", err)
			return
		}

		drawScreen(gameState)
	}
}
