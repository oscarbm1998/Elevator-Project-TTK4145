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
	ch_new_order chan bool,

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
		case <-ch_order:
			fsm_newOrder()
		case <-ch_elevator_has_arrived:
			fsm_onFloorArival(ch_door_timer_reset)
		case <-ch_door_timer_out:
			fsm_doorTimeOut()
			// Lag en case her for hva som skjer hvis heisen er stuck for lenge
		}
	}
}

func CheckIfElevatorHasArrived(ch_drv_floors <-chan int, ch_elevator_has_arrived chan bool) {
	if elevator_command.floor == ch_drv_floors { //Legg inn hvilken etasje heisen skal til fra et struct
		ch_elevator_has_arrived <- true //Kan være denne vil fortsette å kjøre så kan hende vi må fikse
	}
}

func fsm_newOrder() {
	switch current_state {
	case idle:
		//Beveg heis til ønsket etasje (hente dette fra en struct som inneholder direction og floor den skal til?)
		elevio.SetMotorDirection(elevio.MotorDirection(elevator_command.direction))
		fmt.Printf("Moving to floor " + elevator_command.floor)
		current_state = moving
	case moving:
		//Velger egentlig stop etasje basert på hva som ligger i struct men tror det gjøres hos joel?
	case doorOpen:
		//Vent til dørene lukkes eller personen inni trykker på noe. Hvis doortimer går ut sjekker heisen om det
		//finnes noen nye utvendige calls den skal ta
	}
}

func fsm_onFloorArival(ch_door_timer_reset chan bool) {
	fmt.Printf("Arrived at floor" + elevator_info.floor)
	// Write to a struct somewhere that elevator has arrived on correct floor
	// Send UDP that elevator has arrived so the others can shut of timmer (Don't need for single)
	// Stop heis
	// Skru av etasje lys
	// Åpne dør
	// Skru på dør timer
	// Sett state = doorOpen
	switch current_state {
	case moving:
		elevio.SetMotorDirection(elevio.MD_Stop)
		button_type_to_clear() //Clears the light on the correct button, needs input wheter it is cab or hall
		elevio.SetDoorOpenLamp(true)
		ch_door_timer_reset <- true
		current_state = doorOpen
		//Clear call that it has arrived
	default: 
		fmt.Printf("Arrived at floor outside of state moving. Something is wrong")
	}
}

func fsm_doorTimeOut() {
	fmt.Printf("Door time out detected")
	switch current_state {
	case doorOpen:
		elevio.SetDoorOpenLamp(false)
		// Lukk dør
		// Check if there are any other orders it needs to complete
		// sett heis tilbake til idle
		current_state = idle
	}
}

func button_type_to_clear() {
	if //hall call then, else if cab bare clear cab
	switch elevator.direction //Button type må enten være hall_up, hall_down, cab
	case up:
		elevio.SetButtonLamp(elevio.BT_HallUp, elevator_info.floor, false)
	case down: 
		elevio.SetButtonLamp(elevio.BT_HallDown, elevator_info.floor, false)
}
