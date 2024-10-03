package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type UnitType string

const (
	Infantry   UnitType = "Infantry"
	Tank       UnitType = "Tank"
	Artillery  UnitType = "Artillery"
	AirSupport UnitType = "AirSupport"
)

type Unit struct {
	Name     string
	Type     UnitType
	Health   int
	Atk      int
	Range    int
	Movement int
	X, Y     int // Position on the battlefield
	Mutex    sync.Mutex
}

// Battlefield size
const MapSize = 10

var wg sync.WaitGroup

// Create a new unit with specified attributes
func NewUnit(name string, unitType UnitType, x, y int) *Unit {
	switch unitType {
	case Infantry:
		return &Unit{Name: name, Type: Infantry, Health: 100, Atk: 10, Range: 1, Movement: 2, X: x, Y: y}
	case Tank:
		return &Unit{Name: name, Type: Tank, Health: 200, Atk: 40, Range: 2, Movement: 3, X: x, Y: y}
	case Artillery:
		return &Unit{Name: name, Type: Artillery, Health: 150, Atk: 60, Range: 4, Movement: 1, X: x, Y: y}
	case AirSupport:
		return &Unit{Name: name, Type: AirSupport, Health: 80, Atk: 100, Range: 6, Movement: 5, X: x, Y: y}
	default:
		return nil
	}
}

// Move a unit on the map in real-time
func (u *Unit) Move(targetX, targetY int) {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()

	if targetX >= 0 && targetX < MapSize && targetY >= 0 && targetY < MapSize {
		fmt.Printf("%s moving to (%d, %d)\n", u.Name, targetX, targetY)
		u.X, u.Y = targetX, targetY
	} else {
		fmt.Printf("%s tried to move out of bounds\n", u.Name)
	}
}

// Attack a target unit in real-time
func (u *Unit) Attack(target *Unit) {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()

	if target.Health <= 0 {
		return
	}

	fmt.Printf("%s attacks %s at (%d, %d)\n", u.Name, target.Name, target.X, target.Y)
	target.Health -= u.Atk
	if target.Health <= 0 {
		fmt.Printf("%s has been destroyed!\n", target.Name)
	}
}

// Handle unit actions in real-time using Goroutines
func handleUnitActions(unit *Unit, enemyUnits []*Unit, done chan struct{}, gameOver *bool, gameOverMutex *sync.Mutex) {
	defer wg.Done()

	for unit.Health > 0 && !checkGameOver(gameOver, gameOverMutex) {
		// Simulate moving randomly every few seconds
		newX := rand.Intn(MapSize)
		newY := rand.Intn(MapSize)
		unit.Move(newX, newY)

		// Look for nearby enemies to attack
		for _, enemy := range enemyUnits {
			if enemy.Health > 0 && abs(unit.X-enemy.X) <= unit.Range && abs(unit.Y-enemy.Y) <= unit.Range {
				unit.Attack(enemy)
				if enemy.Health <= 0 {
					break
				}
			}
		}

		// Sleep to simulate real-time action
		time.Sleep(time.Second * time.Duration(1+rand.Intn(3)))
	}

	// Notify that this unit has finished processing
	done <- struct{}{}
}

// Utility function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Check if all units on one side are destroyed
func checkGameOver(gameOver *bool, gameOverMutex *sync.Mutex) bool {
	gameOverMutex.Lock()
	defer gameOverMutex.Unlock()
	return *gameOver
}

// Set game over flag
func setGameOver(gameOver *bool, gameOverMutex *sync.Mutex) {
	gameOverMutex.Lock()
	*gameOver = true
	gameOverMutex.Unlock()
}

// Check if all units from one side are destroyed
func checkAllUnitsDestroyed(units []*Unit) bool {
	for _, unit := range units {
		if unit.Health > 0 {
			return false
		}
	}
	return true
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Initialize player units
	playerUnits := []*Unit{
		NewUnit("Alpha", Infantry, 0, 0),
		NewUnit("Bravo", Tank, 1, 1),
		NewUnit("Charlie", Artillery, 2, 2),
		NewUnit("Delta", AirSupport, 0, 0),
	}

	// Initialize enemy units
	enemyUnits := []*Unit{
		NewUnit("Enemy 1", Infantry, 8, 8),
		NewUnit("Enemy 2", Tank, 9, 9),
		NewUnit("Enemy 3", Artillery, 7, 7),
		NewUnit("Enemy 4", AirSupport, 9, 9),
	}

	done := make(chan struct{}, len(playerUnits)+len(enemyUnits)) // Buffer large enough to avoid blocking

	// Game over flag
	var gameOver bool
	var gameOverMutex sync.Mutex

	// Run player unit actions concurrently
	for _, unit := range playerUnits {
		wg.Add(1)
		go handleUnitActions(unit, enemyUnits, done, &gameOver, &gameOverMutex)
	}

	// Run enemy unit actions concurrently
	for _, unit := range enemyUnits {
		wg.Add(1)
		go handleUnitActions(unit, playerUnits, done, &gameOver, &gameOverMutex)
	}

	// Monitor the game status in a separate Goroutine
	go func() {
		for {
			time.Sleep(time.Second)

			// Check if player units or enemy units are all destroyed
			if checkAllUnitsDestroyed(playerUnits) {
				setGameOver(&gameOver, &gameOverMutex)
				fmt.Println("Enemy wins!")
				break
			}

			if checkAllUnitsDestroyed(enemyUnits) {
				setGameOver(&gameOver, &gameOverMutex)
				fmt.Println("Player wins!")
				break
			}
		}
	}()

	// Wait for all units to finish
	wg.Wait()

	// Drain the done channel to ensure all Goroutines finish
	for i := 0; i < len(playerUnits)+len(enemyUnits); i++ {
		<-done
	}
}
