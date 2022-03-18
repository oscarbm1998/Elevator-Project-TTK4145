package singleElevator

import (
	"PROJECT-GROUP-10/config"
	"PROJECT-GROUP-10/elevio"
	"PROJECT-GROUP-10/networking"
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
var elevator_door_blocked bool

func SingleElevatorFSM(
	ch_drv_floors <-chan int,
	ch_elevator_has_arrived chan bool,
	ch_obstr_detected <-chan bool,
	ch_new_order <-chan bool,
	ch_drv_stop <-chan bool,
	ch_req_ID chan int,
	ch_req_data chan networking.Elevator_node, //Should be write only
	ch_write_data chan networking.Elevator_node,
) {
	ch_door_timer_out := make(chan bool)
	ch_door_timer_reset := make(chan bool)
	go OpenAndCloseDoorsTimer(ch_door_timer_out, ch_door_timer_reset)
	go CheckIfElevatorHasArrived(ch_drv_floors, ch_elevator_has_arrived, ch_req_ID, ch_req_data, ch_write_data)
	RemoveAllLights()
	elevator.direction = 0
	elevator.floor = 0
	current_state = idle
	last_floor = -1
	for {
		select {
		case <-ch_new_order:
			switch current_state {
			case idle:
				if Request_next_action(elevator.direction) {
					elevio.SetMotorDirection(elevio.MotorDirection(elevator_command.direction))
					fmt.Printf("Moving to floor %+v\n", elevator_command.floor)
					update_elevator_node("direction", elevator_command.direction, ch_req_ID, ch_req_data, ch_write_data)
					update_elevator_node("destination", elevator_command.floor, ch_req_ID, ch_req_data, ch_write_data)
					fmt.Printf("Here")
					current_state = moving
				} else {
					elevio.SetMotorDirection(elevio.MotorDirection(0))
					update_elevator_node("direction", elevator_command.direction, ch_req_ID, ch_req_data, ch_write_data)
				}
			case moving:
				fmt.Printf("Moving to floor %+v\n", elevator_command.floor)
				Request_next_action(elevator.direction)
			case doorOpen:
				if request_cab() {
					elevio.SetDoorOpenLamp(false)
					elevio.SetMotorDirection(elevio.MotorDirection(elevator_command.direction))
					update_elevator_node("direction", elevator_command.direction, ch_req_ID, ch_req_data, ch_write_data)
					update_elevator_node("destination", elevator_command.floor, ch_req_ID, ch_req_data, ch_write_data)
					current_state = moving
				}
			}
		case <-ch_elevator_has_arrived:
			fmt.Printf("Arrived at floor %+v\n", elevator_command.floor)
			switch current_state {
			case moving:
				elevio.SetMotorDirection(elevio.MD_Stop)
				update_elevator_node("direction", elevio.MD_Stop, ch_req_ID, ch_req_data, ch_write_data)
				elevio.SetDoorOpenLamp(true)
				ch_door_timer_reset <- true
				Update_position(elevator_command.floor, elevator_command.direction) //Bytt navn ?
				current_state = doorOpen
			default:
				fmt.Printf("Arrived at floor outside of state moving. Something is wrong")
			}
		case <-ch_door_timer_out:
			fmt.Printf("Door time out detected\n")
			switch current_state {
			case doorOpen:
				if elevator_door_blocked {
					fmt.Printf("Door blocked by obstruction, can't close door\n")
					fmt.Printf("Waiting 3 more seconds\n")
					ch_door_timer_reset <- true
				} else {
					elevio.SetDoorOpenLamp(false)
					if Request_next_action(elevator_command.direction) {
						elevio.SetMotorDirection(elevio.MotorDirection(elevator_command.direction))
						update_elevator_node("direction", elevator_command.direction, ch_req_ID, ch_req_data, ch_write_data)
						fmt.Printf("Moving to floor %+v\n", elevator_command.floor)
						if elevator_command.floor == elevator.floor {
							current_state = doorOpen
							elevio.SetDoorOpenLamp(true)
							Update_position(elevator_command.floor, (elevator_command.direction))
							ch_door_timer_reset <- true
						} else {
							current_state = moving
						}
					} else {
						fmt.Printf("No new orders, returning to idle\n")
						current_state = idle
					}
				}
			}
		case msg := <-ch_drv_stop: //Maybe change the name on channel to make it more clear
			if msg {
				elevio.SetMotorDirection(elevio.MD_Stop)
				fmt.Printf("Elevator stopped")
				update_elevator_node("direction", elevio.MD_Stop, ch_req_ID, ch_req_data, ch_write_data)
			} else {
				elevio.SetMotorDirection(elevio.MotorDirection(elevator_command.direction))
				fmt.Printf("Elevator running")
				update_elevator_node("direction", elevator_command.direction, ch_req_ID, ch_req_data, ch_write_data)
			}
		case msg := <-ch_obstr_detected:
			if msg {
				elevator_door_blocked = true
			} else {
				elevator_door_blocked = false
			}
		}
	}
}

func CheckIfElevatorHasArrived(ch_drv_floors <-chan int,
	ch_elevator_has_arrived chan bool,
	ch_req_ID chan int,
	ch_req_data chan networking.Elevator_node,
	ch_write_data chan networking.Elevator_node) {
	for {
		select {
		case msg := <-ch_drv_floors:
			elevator.floor = msg
			update_elevator_node("floor", msg, ch_req_ID, ch_req_data, ch_write_data)
			if last_floor == -1 {
				last_floor = elevator.floor
			}
			elevio.SetFloorIndicator(msg)
			if msg == 3 {
				elevator_command.direction = -1
			} else if msg == 0 {
				elevator_command.direction = 1
			}
			fmt.Printf("%d\n", msg)
			if elevator_command.floor == msg && last_floor != elevator_command.floor {
				last_floor = elevator_command.floor
				ch_elevator_has_arrived <- true
			}
		}
	}
}

func RemoveAllLights() { //Move this when time
	for i := 0; i < floor_ammount; i++ {
		elevio.SetButtonLamp(0, i, false)
		elevio.SetButtonLamp(1, i, false)
		elevio.SetButtonLamp(2, i, false)
	}
}

func update_elevator_node(
	input string,
	value int,
	ch_req_ID chan int,
	ch_req_data, ch_write_data chan networking.Elevator_node) {
	updated_elevator_node := networking.Node_get_data(
		config.ELEVATOR_ID,
		ch_req_ID,
		ch_req_data)
	switch input {
	case "floor":
		updated_elevator_node.Floor = value
	case "direction":
		updated_elevator_node.Direction = value
	case "destination":
		updated_elevator_node.Destination = value
	}
	updated_elevator_node.ID = config.ELEVATOR_ID
	//Samme for alt annet som må oppdaterers
	ch_write_data <- updated_elevator_node
}
