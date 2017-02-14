package main

import (
	"errors"
	"fmt"

	"github.com/processout/fsm"
)

// It is recommended to read the comments ;).

// Transaction which will be the data that flows.
// Has to implement IDer (see `ID()`) for the fsm package.
type Transaction struct {
	Status      string
	Amount      float64
	AdressLine1 string
	// pending, failed bool
}

// ID will return the status as to define the different states.
// You don't have to return a string here, you can return anything/
// You could also return a structure which would mean that your
// states would depend on these two fields (see states below)
func (t Transaction) ID() fsm.ID { return t.Status }

var (
	// Here we describe the states that are possible
	// It's your data structure with the IDable fields defined
	// in `ID()` set to their values.
	statePdg = Transaction{Status: "pending"}
	stateAth = Transaction{Status: "authorized"}
	stateCpt = Transaction{Status: "captured"}
	// Important:
	// If our `ID()` returned the transaction structure with three
	// fields (Status, pending, failed),
	// we could define states as:
	/*
	 * stateAthP = Transaction{Status: "authorization", pending: true}
	 * stateAthF = Transaction{Status: "authorization", failed: true}
	 * stateAthS = Transaction{Status: "authorization", failed: false}
	 * stateCapF = Transaction{Status: "capture", failed: true}
	 * stateCapS = Transaction{Status: "capture", failed: false}
	 */
	// This allows you to define flows without changing your data
	// structures. Here the flow could be: (see image)
)

// machine is our actual FSM
var machine = fsm.Machine{}

// sampleGuard will be a guard for checking if a transition
// can go through
func sampleGuard(start fsm.State, goal fsm.State) error {
	// I() returns the interface (of the transaction)
	if start.I().(Transaction).Amount <= 0 {
		return errors.New("Can't transition, amount is <= 0")
	}
	if start.I().(Transaction).Amount != goal.I().(Transaction).Amount {
		return errors.New("Can't transition, amount is different")
	}
	return nil
}

func init() {
	// Define our machine's rules
	rules := fsm.Ruleset{}

	// Transition: Pending -> Authorized with rule sampleGuard
	rules.AddRule(
		fsm.NewTransition(statePdg, stateAth),
		sampleGuard)
	// Transition: Authorized -> Captured with no additional rules
	rules.AddTransition(
		fsm.NewTransition(stateAth, stateCpt))

	machine.Rules = &rules
}

func main() {
	testFlow()
}

// Now that we have the machine and it's rules defined, we can
// test the flow like this:
func testFlow() {
	// Imagine you have your data that's trying to go from
	// pending to authorized.
	pending := Transaction{Status: "pending"}
	authorized := Transaction{Status: "authorized"}

	machine.State = fsm.NewState(pending)
	if err := machine.Transition(fsm.NewState(authorized)); err != nil {
		fmt.Println("1. Transition failed:", err)
	}

	pending.Amount = 1.99
	authorized.Amount = 1.99
	machine.State = fsm.NewState(pending)
	if err := machine.Transition(fsm.NewState(authorized)); err != nil {
		fmt.Println("2. Transition failed:", err)
	}

	// Now from authorized to authorized (not possible in current schema)
	if err := machine.Transition(fsm.NewState(authorized)); err != nil {
		fmt.Println("3. Transition failed:", err)
	}

	// First 'if' will pass, second 'if' will fail.
}
