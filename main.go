package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
)

// Card holds the information for a card in the game
type Card struct {
	id      int64
	keyword string
	combo   bool
}

// Results records the results of a scenario run
type Results struct {
	attempts       int64
	averageTurnWin float64
}

type Simulation struct {
	// turn number that drew into combo piece win
	turn int64
}

// this first scenario models a 37 land deck with 62 permanents and
// 2 combo pieces. this deck is then shuffled several times and run
// until it hits it's combo and records the results.
func main() {
	fmt.Println("mtg-sim booting up")
	var input = make(chan Simulation, 10000)
	results, err := runScenario(input)
	if err != nil {
		log.Fatalf("error: %+v", err)
	}
	fmt.Printf("results: %v\n", results)
}

func runScenario(input chan Simulation) (Results, error) {
	var numSimulations = 10_000
	var results = Results{}

	wg := &sync.WaitGroup{}
	wg.Add(numSimulations)

	go func(input chan Simulation) {
		for i := 0; i < numSimulations; i++ {
			deck := createDeck()
			input <- runSimulation(deck)
		}
	}(input)

	var turnCounts = []int64{}
	go func() {
		for {
			select {
			case sim := <-input:
				fmt.Printf("sim: %v\n", sim)
				results.attempts++
				turnCounts = append(turnCounts, sim.turn)
				wg.Done()
			}
		}
	}()

	wg.Wait()

	// calculate the sum
	var sum int64 = 0
	for _, value := range turnCounts {
		sum += value
	}

	// Calculate the average
	average := float64(sum) / float64(len(turnCounts))
	results.averageTurnWin = average

	return results, nil
}

func createDeck() []Card {
	// setup the distribution of cards for our simulation
	var numLands = 37
	// set the number of non-lands to the rest of the deck
	var numNonLands = 99 - numLands
	// assumes the commander is not a part of the combo strategy
	var numComboPieces = 2

	// create a deck
	var deck []Card

	// add lands to the deck
	for i := 0; i < numLands; i++ {
		deck = append(deck, Card{
			id:      int64(i),
			keyword: "land",
		})
	}

	// add non-combo permanents
	for i := 0; i < numNonLands-numComboPieces; i++ {
		deck = append(deck, Card{
			id:      int64(i),
			keyword: "non-land",
			combo:   false,
		})
	}

	// finally, add the appropriate number of combo pieces to the deck.
	// it is assumed that all combo pieces must be drawn to trigger
	// the win condition.
	for i := 0; i < numComboPieces; i++ {
		deck = append(deck, Card{
			id:      int64(i),
			keyword: "non-land",
			combo:   true,
		})
	}

	return shuffleDeck(deck)
}

func shuffleDeck(deck []Card) []Card {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	return deck
}

// runSimulation starts drawing down until it hits a win con and
// then records the results of the simulation for later analysis
func runSimulation(deck []Card) Simulation {
	var turnCount int64 = 0
	hand, deck := deck[:7], deck[7:]
	log.Printf("hand: %+v\n - deck: %+v\n", hand, deck)

	if checkComboWin(hand, 2) {
		return Simulation{
			turn: turnCount,
		}
	}

	for i := 0; i < len(deck)-len(hand); i++ {
		turnCount++
		// draw
		drawn := deck[0]
		deck = deck[1:]
		hand = append(hand, drawn)
		// check if enough combo pieces have been hit
		if checkComboWin(hand, 2) {
			return Simulation{
				turn: turnCount,
			}
		}
	}

	return Simulation{turn: turnCount}
}

// checks if the required number of combo cards has been drawn
// into hand for a naive win-con check
func checkComboWin(hand []Card, required int64) bool {
	var count int64 = 0
	for i := 0; i < len(hand); i++ {
		if hand[i].combo {
			count++
			if count == required {
				return true
			}
		}
	}
	return false
}
