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
var last_floor int
var current_floor int

func SingleElevatorFSM(
	ch_drv_floors <-chan int,
	ch_elevator_has_arrived chan bool,
	ch_drv_obstr chan bool,
	ch_new_order <-chan bool,

	// Channel koblet til orders
	// Channel koblet til door time out
	// Channel koblet til elevator_obstrukjson time out
	// Channel

) {
	ch_door_timer_out := make(chan bool)
	ch_door_timer_reset := make(chan bool)
	last_floor = -1
	go OpenAndCloseDoorsTimer(ch_door_timer_out, ch_door_timer_reset)
	go CheckIfElevatorHasArrived(ch_drv_floors, ch_elevator_has_arrived)
	elevator.direction = 0
	elevator.floor = 0
	last_floor = 0
	current_floor = -1
	current_state = idle
	for {
		select {
			case <-ch_new_order:
				switch current_state {
					case idle:
						//Beveg heis til ønsket etasje (hente dette fra en struct som inneholder direction og floor den skal til?)
						if Call_qeuer(elevator.direction) {
							elevio.SetMotorDirection(elevio.MotorDirection(elevator_command.direction))
							fmt.Printf("Moving to floor %+v\n", elevator_command.floor)
							current_state = moving
						} else {
							elevio.SetMotorDirection(elevio.MotorDirection(0))
						}
					case moving:
						fmt.Printf("Moving to floor %+v\n", elevator_command.floor)
					case doorOpen:
						//Vent til dørene lukkes eller personen inni trykker på noe. Hvis doortimer går ut sjekker heisen om det
						if request_here() {
							elevio.SetDoorOpenLamp(false)
							elevio.SetMotorDirection(elevio.MotorDirection(elevator_command.direction))
							current_state = moving
						}
					}
			case <-ch_elevator_has_arrived:
				fmt.Printf("Arrived at floor %+v\n", elevator_command.floor)
				// Send UDP that elevator has arrived so the others can shut of timmer (Don't need for single)
				switch current_state {
					case moving:
						elevio.SetMotorDirection(elevio.MD_Stop)
						elevio.SetDoorOpenLamp(true)
						ch_door_timer_reset <- true
						Update_position(elevator_command.floor, elevator_command.direction)
						current_state = doorOpen
						//Clear call that it has arrived
					default:
						fmt.Printf("Arrived at floor outside of state moving. Something is wrong")
				}
			case <-ch_door_timer_out:
				fmt.Printf("Door time out detected\n")
				switch current_state {
					case doorOpen:
						elevio.SetDoorOpenLamp(false)
						if Call_qeuer(elevator_command.direction) {
							elevio.SetMotorDirection(elevio.MotorDirection(elevator_command.direction))
							fmt.Printf("Moving to floor %+v\n", elevator_command.floor)
							current_state = moving
						} else {
							current_state = idle
					}
				}
		}
	}
}

func CheckIfElevatorHasArrived(ch_drv_floors <-chan int, ch_elevator_has_arrived chan bool) {
	for {
		select {
		case msg := <-ch_drv_floors:
			elevator.floor = msg
			if msg == 3 {
				elevator_command.direction = -1
			} else if msg == 0 {
				elevator_command.direction = 1
			}
			fmt.Printf("%d\n", msg)
			if elevator_command.floor == msg && last_floor != elevator_command.floor {
				last_floor = elevator_command.floor
				ch_elevator_has_arrived <- true //Kan være denne vil fortsette å kjøre så kan hende vi må fikse
			}
		}
	}
}
