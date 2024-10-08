package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type GameState struct {
	Player           Player
	CurrentLocation  string
	VisitedLocations map[string]bool
	Context          []int
}

type Player struct {
	Inventory  []string
	Health     int
	Experience int
  Action     string
}

type OllamaRequest struct {
	Model   string `json:"model"`
	System  string `json:"system"`
	Prompt  string `json:"prompt"`
	Stream  bool   `json:"stream"`
	Context []int  `json:"context,omitempty"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Context  []int  `json:"context"`
}

func callOllama(prompt string, gameState *GameState) (string, error) {
	client := &http.Client{}
	reqBody, err := json.Marshal(OllamaRequest{
		Model:   "llama3.2",
		System:  "You are a dungeon master in a d&d game. Your task is to assess the situation and interpret the player's actions. You should keep responses short and to the point, and provide enough information to keep the game moving.",
		Prompt:  prompt,
		Stream:  false,
		Context: gameState.Context,
	})
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	resp, err := client.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("error unmarshaling response: %v", err)
	}

	if ollamaResp.Response == "" {
		return "", fmt.Errorf("received empty response from Ollama")
	}

	gameState.Context = ollamaResp.Context
	return ollamaResp.Response, nil
}

func mainGameLoop(gameState *GameState) error {
	reader := bufio.NewReader(os.Stdin)
	for {
		// Get current situation from Ollama
    situation, err := callOllama(fmt.Sprintf("The current location of the plauer is %s. The player says: %s, what will the dungeon master do say?", gameState.CurrentLocation, gameState.Player.Action), gameState)
		if err != nil {
			return fmt.Errorf("error getting situation: %v", err)
		}
		fmt.Println(situation)
		fmt.Println("Context length:", len(gameState.Context))

		// Get player input
		fmt.Print("> ")
		action, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading input: %v", err)
		}
		gameState.Player.Action = strings.TrimSpace(action)

		// Update game state based on result
		updateGameState(gameState, situation)

		// Check for game end conditions
		if checkGameEnd(gameState) {
			break
		}
	}
	return nil
}

func updateGameState(gameState *GameState, result string) {
	gameState.CurrentLocation = result
}

func checkGameEnd(gameState *GameState) bool {
	// Implement win/lose conditions
	return false
}

func main() {
	gameState := &GameState{
		Player: Player{
			Inventory:  []string{},
			Health:     100,
			Experience: 0,
      Action:     "I am walking vigilantly",
		},
		CurrentLocation:  "a mystical forest",
		VisitedLocations: make(map[string]bool),
		Context:          []int{},
	}
	if err := mainGameLoop(gameState); err != nil {
		log.Fatal(err)
	}
}
