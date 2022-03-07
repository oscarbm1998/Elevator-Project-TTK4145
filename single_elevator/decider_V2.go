package singleElevator

import (
	"PROJECT-GROUP-10/elevio"
)

const floor_ammount int = 3

type dir struct {
	up   bool
	down bool
	stop bool
}

type elevator_status struct {
	floor     int
	direction int //1 up -1 down 0 idle
}

type floor_info struct {
	up bool
	down  bool
	here bool
}

var floor [floor_ammount]floor_info
var elevator elevator_status         //where elevator is
var elevator_command elevator_status //where elevator should go

func Remove_order(level int, direction int) { //removes an order
	floor[level].here = false //removes here call as the elevator has arrived there
	if direction == 1 { //if the direction is up
		floor[level].up = false //disables the up direction
		elevio.SetButtonLamp(0, level, false) //turns off light
	} else if direction == -1 { //if the direction is down
		floor[level].down = false //disables the down direction
		elevio.SetButtonLamp(1, level, false) //turns off light
	}
}
/* REVISED may not be needed
func Hall_order(
	ch_drv_buttons chan elevio.ButtonEvent,
	ch_new_order chan bool,
	ch_clear_order chan bool,
	ch_drv_floors chan int,
) {
	for {
		select {
		case a := <-ch_drv_buttons:
			switch a.Button {
			case 1: //opp
				floor[a.Floor].hall_call = 1
				floor[a.Floor].direction.up = true
				break
			case -1: //ned
				floor[a.Floor].hall_call = 1
				floor[a.Floor].direction.down = true
				hall_calls()
				break
			case 0: //cab call
				floor[a.Floor].cab_call = 1
				Cab_calls()
			}
			ch_new_order <- true //forteller at en ny order er tilgjengelig
		}
	}
}
*/
func request_above() bool { //checks if there are any active calls above the elevator and updates the "command struct"
	for i := elevator.floor; i <= floor_ammount; i++ { //checks from the last known floor of the elevator to the top
		if floor[i].up{ //if a floor with call up is found
			elevator_command.floor = i //updates the command value
			elevator_command.direction = 1; //sets the direction up just in case
			return true
		}
	}
	return false
}

func request_below() bool { //checks if there are any active calls below the elevator and updates the "command struct"
	for i := elevator.floor; i >= floor_ammount; i++ { //checks from the last known floor of the elevator to the botton
		if floor[i].down{ //if a floor with call down is found
			elevator_command.floor = i //updates the command value
			elevator_command.direction = -1; //sets the direction down just in case
			return true
		}
	}
	return false
}

func request_here() bool { //tad unshure if this is needed or not but its used for internal calls
	for i := 0; i <= floor_ammount; i++ { //checks the entire struct for calls
		if floor[i].here{ //if a call is found
			elevator_command.floor = i //update command struct
			if floor[i].here > elevator.floor { //set direction
				elevator_command.direction = 1
			} else {
				elevator_command.direction = -1
			}
			return true
		}
	}
	return false
}

func Call_qeuer(direction int) bool{
	switch direction {
	case 1 //up
		if request_above(){
		} else if request_here() {
		} else if request_below() {
		} else {
			//sett den i idle
		}
		return true
	case -1: //down
		if request_below(){
		} else if request_here() {
		} else if  request_above(){
		} else {
			//sett den i idle
		}
		return true
	case 0 // here
		if request_here(){
		} else if request_above() {
		} else if request_below() {
		} else {
			//sett den i idle
		}
		return true
	}
	return false
 //vite når det kommer en new order
 //
}

func Update_position(int floor, int direction){
	elevator.floor = floor
	elevator.direction = direction
	Remove_order(floor, direction)
}

/*
func Cab_calls() (found_call bool) {
	for i := 0; i <= floor_ammount; i++ {
		if floor[i].cab_call == 1 {
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
