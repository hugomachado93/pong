package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	screenWidth  = 80
	screenHeight = 24
	paddleHeight = 4
	tickRate     = 50 * time.Millisecond
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

var (
	gameState GameState
	mutex     sync.Mutex
)

func initGame() {
	gameState = GameState{
		Player1Y: screenHeight / 2,
		Player2Y: screenHeight / 2,
		BallX:    screenWidth / 2,
		BallY:    screenHeight / 2,
		BallVX:   1,
		BallVY:   1,
	}
}

func updateGame() {
	mutex.Lock()
	defer mutex.Unlock()

	gameState.BallX += gameState.BallVX
	gameState.BallY += gameState.BallVY

	if gameState.BallY < 0 || gameState.BallY >= screenHeight {
		gameState.BallVY = -gameState.BallVY
	}

	if (gameState.BallX == 1 && gameState.BallY >= gameState.Player1Y && gameState.BallY < gameState.Player1Y+paddleHeight) ||
		(gameState.BallX == screenWidth-2 && gameState.BallY >= gameState.Player2Y && gameState.BallY < gameState.Player2Y+paddleHeight) {
		gameState.BallVX = -gameState.BallVX
	}

	if gameState.BallX < 0 || gameState.BallX >= screenWidth {
		gameState.BallX = screenWidth / 2
		gameState.BallY = screenHeight / 2
		gameState.BallVX = -gameState.BallVX
	}
}

func handleClient(conn net.Conn, player int) {
	defer conn.Close()

	decoder := gob.NewDecoder(conn)

	for {
		var input string
		if err := decoder.Decode(&input); err != nil {
			log.Println("Error decoding input:", err)
			return
		}

		mutex.Lock()
		switch input {
		case "up":
			if player == 1 && gameState.Player1Y > 0 {
				gameState.Player1Y--
			} else if player == 2 && gameState.Player2Y > 0 {
				gameState.Player2Y--
			}
		case "down":
			if player == 1 && gameState.Player1Y < screenHeight-paddleHeight {
				gameState.Player1Y++
			} else if player == 2 && gameState.Player2Y < screenHeight-paddleHeight {
				gameState.Player2Y++
			}
		}
		mutex.Unlock()

	}
}

func main() {
	connChan := make(chan net.Conn)
	initGame()
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go func() {
		encodPlayers := make([]*gob.Encoder, 2)

		for i := 0; i <= 1; i++ {
			conn := <-connChan
			encodPlayers[i] = gob.NewEncoder(conn)
		}

		for {
			time.Sleep(tickRate)
			updateGame()
			fmt.Println(gameState)
			gameState.PlayerNum = 1
			encodPlayers[0].Encode(gameState)
			gameState.PlayerNum = 2
			// gameState.Player1Y
			// screenWidthP1, _ = screen.Size()
			// screenWidthP1 -= 1
			// screenWidthP2 = 0

			// screenHeightP1 = 23 - gameState.Player1Y
			// screenHeightP2 = 20 - gameState.Player2Y

			// ballXP = 80 - gameState.BallX
			// ballYP = 23 - gameState.BallY
			encodPlayers[1].Encode(gameState)
		}
	}()

	i := 1
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleClient(conn, i)
		connChan <- conn
		if i == 2 {
			i = 0
		} else {
			i++
		}
	}
}
