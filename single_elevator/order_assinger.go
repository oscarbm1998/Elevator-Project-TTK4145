package singleElevator

import (
	"PROJECT-GROUP-10/config"
	"PROJECT-GROUP-10/elevio"
	"PROJECT-GROUP-10/networking"
	"fmt"
)

type elevator_status struct {
	floor     int
	direction int //1 up -1 down 0 idle
}

type floor_info struct {
	up   bool
	down bool
	cab  bool
}

var floor [config.NUMBER_OF_FLOORS]floor_info
var elevator elevator_status         //where elevator is
var elevator_command elevator_status //where elevator should go

func Remove_order(level int, direction int) { //removes an order
	floor[level].cab = false              //removes here call as the elevator has arrived there
	elevio.SetButtonLamp(2, level, false) //turns off cab light
	if direction == int(elevio.MD_Up) {   //if the direction is up or there are no orders below and orders above
		if !floor[level].up {
			floor[level].down = false
		} else {
			floor[level].up = false
		}
		//disables the up direction
	} else if direction == int(elevio.MD_Down) { //if the direction is down or there are no orders above and orders below
		if !floor[level].down {
			floor[level].up = false
		} else {
			floor[level].down = false
		}
	} else if direction == int(elevio.MD_Stop) {
		if !floor[level].down {
			floor[level].up = false
		} else {
			floor[level].down = false
		}
	}
}

func Hall_order(
	ch_new_order chan bool,
	ch_net_command chan elevio.ButtonEvent,
	ch_self_command chan elevio.ButtonEvent,
	ch_req_ID chan int,
	ch_req_data chan networking.Elevator_node,
	ch_write_data chan networking.Elevator_node,
) {
	for {
		select {
		case a := <-ch_net_command:
			if (floor[a.Floor].up && a.Button == 0) || (floor[a.Floor].down && a.Button == 1) || floor[a.Floor].cab || (a.Floor == elevator.floor) {
				fmt.Printf("orders already exists\n")
			} else {
				switch a.Button {
				case elevio.BT_HallUp: //opp
					floor[a.Floor].up = true
					update_elevator_node("update order up", a.Floor, ch_req_ID, ch_req_data, ch_write_data)
				case elevio.BT_HallDown: //ned
					floor[a.Floor].down = true
					update_elevator_node("update order down", a.Floor, ch_req_ID, ch_req_data, ch_write_data)
				case elevio.BT_Cab: //cab call
					floor[a.Floor].cab = true
					elevio.SetButtonLamp(2, a.Floor, true) //turns off light
				}
				ch_new_order <- true //forteller at en ny order er tilgjengelig
			}
		case a := <-ch_self_command:
			if (floor[a.Floor].up && a.Button == 0) || (floor[a.Floor].down && a.Button == 1) || floor[a.Floor].cab || (a.Floor == elevator.floor) {
				//do nuffin as the order already exists
				fmt.Printf("orders already exists\n")
				//Remove_order(a.Floor, a.Floor)
			} else { //do shit
				switch a.Button {
				case elevio.BT_HallUp: //opp
					floor[a.Floor].up = true
					update_elevator_node("update order up", a.Floor, ch_req_ID, ch_req_data, ch_write_data)
				case elevio.BT_HallDown: //ned
					floor[a.Floor].down = true
					update_elevator_node("update order down", a.Floor, ch_req_ID, ch_req_data, ch_write_data)
				case elevio.BT_Cab: //cab call
					floor[a.Floor].cab = true
					elevio.SetButtonLamp(2, a.Floor, true) //turns off light
				}
				ch_new_order <- true //forteller at en ny order er tilgjengelig
			}
		}
	}
}

func request_above() bool { //checks if there are any active calls above the elevator and updates the "command struct"
	for i := elevator.floor + 1; i < config.NUMBER_OF_FLOORS; i++ { //checks from the last known floor of the elevator to the top
		if floor[i].up || floor[i].down { //if a floor with call up is found
			elevator_command.floor = i                     //updates the command value
			elevator_command.direction = int(elevio.MD_Up) //sets the direction up just in case
			return true
		}
	}
	return false
}

func request_here() bool {
	if floor[elevator.floor].up || floor[elevator.floor].down {
		elevator_command.floor = elevator.floor //updates the command value
		elevator_command.direction = 0          //sets the direction down just in case
		return true
	}
	return false
}

func request_below() bool { //checks if there are any active calls below the elevator and updates the "command struct"
	for i := elevator.floor - 1; i >= 0; i-- { //checks from the last known floor of the elevator to the botton
		if floor[i].down || floor[i].up { //if a floor with call down is found
			elevator_command.floor = i      //updates the command value
			elevator_command.direction = -1 //sets the direction down just in case
			return true
		}
	}
	return false
}

func request_cab() bool { //tad unshure if this is needed or not but its used for internal calls
	for i := 0; i < config.NUMBER_OF_FLOORS; i++ { //checks the entire struct for calls
		if floor[i].cab { //if a call is found
			elevator_command.floor = i //update command struct
			if i > elevator.floor {    //set direction
				elevator_command.direction = int(elevio.MD_Up)
			} else {
				elevator_command.direction = int(elevio.MD_Down)
			}
			return true
		}
	}
	return false
}

func Request_next_action(direction int) bool {
	switch direction {
	case int(elevio.MD_Up): //up
		if request_above() {
			return true
		} else if request_cab() {
			return true
		} else if request_below() {
			return true
		} else if request_here() {
			return true
		}

	case elevio.MD_Down: //down
		if request_below() {
			return true
		} else if request_cab() {
			return true
		} else if request_above() {
			return true
		} else if request_here() {
			return true
		}

	case elevio.MD_Stop: // here
		if request_cab() {
			return true
		} else if request_above() {
			return true
		} else if request_below() {
			return true
		}
	}
	return false
}

func Update_position(level int, direction int) {
	elevator.floor = level
	elevator.direction = direction
	Remove_order(level, direction)
}
