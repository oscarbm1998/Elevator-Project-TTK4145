package singleElevator

import (
	"PROJECT-GROUP-10/elevio"
	"fmt"
)

type elevator_state int

const (
	idle elevator_state = iota
	moving
	doorOpen
)

var current_state elevator_state

func SingleElevatorFSM(
	ch_drv_floors <-chan int,
	ch_elevator_has_arrived chan bool,
	ch_drv_obstr chan bool,

	// Channel koblet til orders
	// Channel koblet til door time out
	// Channel koblet til elevator_obstrukjson time out
	// Channel

) {
	ch_door_timer_out := make(chan bool)
	ch_door_timer_reset := make(chan bool)

	go OpenAndCloseDoorsTimer(ch_door_timer_out, ch_door_timer_reset)
	go CheckIfElevatorHasArrived(ch_drv_floors, ch_elevator_has_arrived, elevator_info.floor)
	current_state = idle

	for {
		select {
		case a := <-ch_order:
			// Får varsel på at nye instruksjoner foreligger i struct så sjekk denne og gå dit
			elevio.SetMotorDirection(elevio.MD_Up) //Fiks denne
			// Skru på lyset for at den er på vei dit
			elevio.SetButtonLamp(b, elevator_info.floor, true)
			//Sett heis i state movement
			// Basert på det

		case <-ch_elevator_has_arrived:
			fsm_onFloorArival(ch_door_timer_reset)
		case <-ch_door_timer_out:
			fsm_doorTimeOut()
			// Lag en case her for hva som skjer hvis heisen er stuck for lenge
		}
	}
}

func CheckIfElevatorHasArrived(ch_drv_floors <-chan int, ch_elevator_has_arrived chan bool, lolxd elevator_info_struct) {
	if lolxd.floor == drv_floors { //Legg inn hvilken etasje heisen skal til fra et struct
		ch_elevator_has_arrived <- true
	}
}

func fsm_onFloorArival(ch_door_timer_reset chan bool) {
	fmt.Printf("Arrived at floor" + elevator_info.floor)
	// Write to a struct somewhere that elevator has arrived on correct floor
	// Send UDP that elevator has arrived so the others can shut of timmer (Don't need for single)
	// Stop heis
	// Skru av etasje lys
	// Åpne dør
	switch current_state {
	case moving:
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevio.SetButtonLamp(b, elevator_info.floor, false)
		elevio.SetDoorOpenLamp(true)
		ch_door_timer_reset <- true
		current_state = doorOpen
	}

}

func fsm_doorTimeOut() {
	fmt.Printf("Door time out detected")
	switch current_state {
	case doorOpen:
		elevio.SetDoorOpenLamp(false)
		// Lukk dør
		// sett heis tilbake til idle
		current_state = idle
	}
}

func fsm_newOrder() {

}
