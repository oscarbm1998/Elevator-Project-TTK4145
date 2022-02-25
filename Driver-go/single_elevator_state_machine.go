package main

import (
	"Driver-go/elevio"
	"fmt"
)

var elevator_state string

// Channel

func singleElevatorFSM(
	drv_floors <-chan int,
	// Channel koblet til orders
	// Channel koblet til door time out
	// Channel koblet til elevator_obstrukjson time out
	// Channel

) {
	ch_door_timer_out := make(chan bool)
	ch_door_timer_reset := make(chan bool)
	ch_elevator_has_arrived := make(chan bool)

	go OpenAndCloseDoorsTimer(ch_door_timer_out, ch_door_timer_reset)
	go CheckIfElevatorHasArrived(drv_floors, ch_elevator_has_arrived)
	elevator_state = "idle"

	for {
		select {
		case a := <-ch_order:
			// Får varsel på at nye instruksjoner foreligger i struct så sjekk denne og gå dit
			elevio.SetMotorDirection(elevator_Info.direction)
			// Basert på det

		case a := <-ch_elevator_has_arrived:
			// Stop heis
			elevio.SetMotorDirection(elevio.MD_Stop)
			// Skru av etasje lys
			elevio.SetButtonLamp(b, elevator_Info.floor, false)
			// Åpne dør
			ch_door_timer_reset <- true

		case a := <-ch_door_timer_out:
			// Lukk dør
			fmt.Printf("Closing door")
			// sett heis tilbake til idle
			elevator_state = "idle"

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			/*
				if a {
					elevio.SetMotorDirection(elevio.MD_Stop)
				} else {
					elevio.SetMotorDirection(d)
				}
			*/
		}
	}
}

func CheckIfElevatorHasArrived(drv_floors chan int, ch_elevator_has_arrived chan bool) {
	if elevator_Info.floor == drv_floors {
		ch_elevator_has_arrived <- true
	}
}
