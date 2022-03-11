package singleElevator

import (
	"PROJECT-GROUP-10/elevio"
	"fmt"
)

const floor_ammount int = 4

type elevator_status struct {
	floor      int
	direction  int //1 up -1 down 0 idle
	up_or_down bool
}

type floor_info struct {
	up   bool
	down bool
	cab  bool
}

var floor [floor_ammount]floor_info
var elevator elevator_status         //where elevator is
var elevator_command elevator_status //where elevator should go

func Remove_order(level int, direction int) { //removes an order
	floor[level].cab = false              //removes here call as the elevator has arrived there
	elevio.SetButtonLamp(2, level, false) //turns off light
	if direction == 1 {                   //if the direction is up or there are no orders below and orders above
		if !floor[level].up {
			floor[level].down = false
			elevio.SetButtonLamp(1, level, false) //turns off light
		} else {
			floor[level].up = false
			elevio.SetButtonLamp(0, level, false) //turns off light
		}
		//disables the up direction
	} else if direction == -1 { //if the direction is down or there are no orders above and orders below
		if !floor[level].down {
			floor[level].up = false
			elevio.SetButtonLamp(0, level, false) //turns off light
		} else {
			floor[level].down = false
			elevio.SetButtonLamp(1, level, false) //turns off light
		}
	} else if direction == 0 {
		if !floor[level].down {
			floor[level].up = false
			elevio.SetButtonLamp(0, level, false) //turns off light
		} else {
			floor[level].down = false
			elevio.SetButtonLamp(1, level, false) //turns off light
		}
	}
}

func Hall_order(
	ch_drv_buttons chan elevio.ButtonEvent,
	ch_new_order chan bool,
) {
	for {
		select {
		case a := <-ch_drv_buttons:
			if (floor[a.Floor].up && a.Button == 0) || (floor[a.Floor].down && a.Button == 1) || floor[a.Floor].cab || (a.Floor == elevator.floor) {
				//do nuffin as the order already exists
				fmt.Printf("orders already exists\n")
				//Remove_order(a.Floor, a.Floor)
			} else { //do shit
				switch a.Button {
				case 0: //opp
					floor[a.Floor].up = true
					elevio.SetButtonLamp(0, a.Floor, true) //turns off light
				case 1: //ned
					floor[a.Floor].down = true
					elevio.SetButtonLamp(1, a.Floor, true) //turns off light
				case 2: //cab call
					floor[a.Floor].cab = true
					elevio.SetButtonLamp(2, a.Floor, true) //turns off light
				}
				ch_new_order <- true //forteller at en ny order er tilgjengelig
			}
		}
	}
}

func request_above() bool { //checks if there are any active calls above the elevator and updates the "command struct"
	for i := elevator.floor + 1; i < floor_ammount; i++ { //checks from the last known floor of the elevator to the top
		if floor[i].up || floor[i].down { //if a floor with call up is found
			fmt.Printf("found request above\n")
			elevator_command.floor = i     //updates the command value
			elevator_command.direction = 1 //sets the direction up just in case
			return true
		}
	}
	return false
}

func request_below() bool { //checks if there are any active calls below the elevator and updates the "command struct"
	for i := elevator.floor - 1; i >= 0; i-- { //checks from the last known floor of the elevator to the botton
		if floor[i].down || floor[i].up { //if a floor with call down is found
			fmt.Printf("found request below\n")
			elevator_command.floor = i      //updates the command value
			elevator_command.direction = -1 //sets the direction down just in case
			return true
		}
	}
	return false
}

func request_cab() bool { //tad unshure if this is needed or not but its used for internal calls
	for i := 0; i < floor_ammount; i++ { //checks the entire struct for calls
		if floor[i].cab { //if a call is found
			elevator_command.floor = i //update command struct
			if i > elevator.floor {    //set direction
				elevator_command.direction = 1
			} else {
				elevator_command.direction = -1
			}
			return true
		}
	}
	return false
}

func request_here() bool {
	if floor[elevator.floor].up || floor[elevator.floor].down {
		fmt.Printf("found request here\n")
		elevator_command.floor = elevator.floor //updates the command value
		elevator_command.direction = 0          //sets the direction down just in case
		return true
	}
	return false
}

func Call_qeuer(direction int) bool {
	switch direction {
	case 1: //up
		if request_above() {
			return true
		} else if request_cab() {
			return true
		} else if request_below() {
			return true
		} else if request_here() {
			return true
		}

	case -1: //down
		if request_below() {
			return true
		} else if request_cab() {
			return true
		} else if request_above() {
			return true
		} else if request_here() {
			return true
		}

	case 0: // here
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

/*
2
			elevator_command.floor = i
			if floor[i].cab_call > elevator.floor {
				elevator_command.direction = 1
			} else {
				elevator_command.direction = -1
			}
			return true
			break
		}
	}
	return false
}
*/
