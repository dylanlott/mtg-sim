package main

// This is a Monte Carlo simulation for how fast a 2 card combo can be
// drawn into in Magic: The Gathering. It simplifies the game down to
// just lands and non-lands, with non-lands being the only cards capable
// of being combo pieces. This simulation assumes 2 combo cards in hand
// is a win-con and doesn't attempt to discern if the combo was castable.

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
)

// Card holds the information for a card in the game
type Card struct {
	keyword string // denotes land or non-land
	combo   bool   // denotes a combo piece
}

// Results collates the simulations of a scenario run
type Results struct {
	attempts               int64
	averageDrawsToWin      float64
	openingHandWins        int64
	averageOpeningHandWins float64
}

// Simulation holds the results of the sim's run
type Simulation struct {
	// drawsToWinCon is the number of draws to find the required
	// number of combo pieces
	drawsToWinCon int64
	// openingHandWinCon is true if the first 7 cards drawn
	// contained the required number of combo pieces
	openingHandWin bool
}

// this first scenario models a 37 land deck with 62 permanents and
// 2 combo pieces. this deck is then shuffled and run until it hits
// both combo pieces snd records the turn count that happened.
func main() {
	fmt.Println("🔮 mtg-sim booting up")

	var numSimulations = 10_000_000
	var input = make(chan Simulation, 10_000)

	results, err := runScenario(input, numSimulations)
	if err != nil {
		log.Fatalf("error: %+v", err)
	}

	fmt.Printf("📊 results:\n%+v\n", results)
}

// runScenario runs a deck simulations a given number of times.
func runScenario(input chan Simulation, numSimulations int) (Results, error) {
	var results = Results{}

	wg := &sync.WaitGroup{}
	wg.Add(numSimulations)

	go func(input chan Simulation) {
		for i := 0; i < numSimulations; i++ {
			deck := createDeck()
			input <- runSimulation(deck)
		}
	}(input)

	var drawCount = []int64{}
	var openingWinCount = 0
	go func() {
		for {
			select {
			case sim := <-input:
				results.attempts++
				// record an opening hand win
				if sim.openingHandWin {
					openingWinCount++
				}
				// record draws to required win
				drawCount = append(drawCount, sim.drawsToWinCon)
				wg.Done()
			}
		}
	}()

	wg.Wait()

	// calculate the sum and average of draw counts
	var sum int64 = 0
	for _, value := range drawCount {
		sum += value
	}

	// calculate average draws to find win-con
	average := float64(sum) / float64(len(drawCount))
	results.averageDrawsToWin = average
	// calculate opening hand win average
	results.averageOpeningHandWins = float64(openingWinCount) / float64(results.attempts)
	results.openingHandWins = int64(openingWinCount)

	return results, nil
}

// createDeck creates a deck with the default setup of lands,
// non-lands, and combo pieces.
func createDeck() []Card {
	// setup the distribution of cards for our simulation
	var numLands = 37
	// set the number of non-lands to the rest of the deck
	var numNonLands = 99 - numLands
	// assumes the commander is not a part of the combo strategy
	var numComboPieces = 4

	// create a deck
	var deck []Card

	// add lands to the deck
	for i := 0; i < numLands; i++ {
		deck = append(deck, Card{
			keyword: "land",
		})
	}

	// add non-combo permanents
	for i := 0; i < numNonLands-numComboPieces; i++ {
		deck = append(deck, Card{
			keyword: "non-land",
			combo:   false,
		})
	}

	// finally, add the appropriate number of combo pieces to the deck.
	// it is assumed that all combo pieces must be drawn to trigger
	// the win condition.
	for i := 0; i < numComboPieces; i++ {
		deck = append(deck, Card{
			keyword: "non-land",
			combo:   true,
		})
	}

	return shuffleDeck(deck)
}

// shuffleDeck shuffles a slice of Cards and returns the shuffled slice
func shuffleDeck(deck []Card) []Card {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	return deck
}

// runSimulation starts drawing down until it hits a win con and
// then records the results of the simulation for later analysis
func runSimulation(deck []Card) Simulation {
	var drawCount int64 = 0
	hand, deck := deck[:6], deck[7:]

	if checkComboWin(hand, 2) {
		return Simulation{
			drawsToWinCon:  drawCount,
			openingHandWin: true,
		}
	}

	for i := 0; i < len(deck)-len(hand); i++ {
		drawCount++
		// draw
		drawn := deck[0]
		deck = deck[1:]
		hand = append(hand, drawn)
		// check if enough combo pieces have been hit
		if checkComboWin(hand, 2) {
			return Simulation{
				drawsToWinCon:  drawCount,
				openingHandWin: false,
			}
		}
	}

	return Simulation{
		drawsToWinCon:  drawCount,
		openingHandWin: false,
	}
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
