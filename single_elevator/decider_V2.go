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
	hall_call int
	cab_call  int
	direction dir
}

var floor [floor_ammount]floor_info
var elevator elevator_status         //where elevator is
var elevator_command elevator_status //where elevator should go

func Remove_order(level int, direction int) {
	floor[level].hall_call = 0
	floor[level].cab_call = 0
	if direction == 1 {
		floor[level].direction.up = true
		elevio.SetButtonLamp(0, level, false)
	} else if direction == -1 {
		floor[level].direction.down = true
		elevio.SetButtonLamp(1, level, false)
	}
}

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

func request_above() bool {
	for i := elevator.floor; i <= floor_ammount; i++ {
		if floor[i].hall_call == 1 {
			elevator_command.floor = i
			if floor[i].hall_call > elevator.floor {
				elevator_command.direction = 1
			} else {
				elevator_command.direction = -1
			}
			return true
		}
	}
	return false
}

func request_below() bool {
	for i := elevator.floor; i <= 0; i-- {
		if floor[i].hall_call == 1 {
			elevator_command.floor = i
			if floor[i].hall_call > elevator.floor {
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
	for i := 0; i <= floor_ammount; i++ {
		if floor[i].hall_call == 1 {
			elevator_command.floor = i
			if floor[i].hall_call > elevator.floor {
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
	case -1: //down
		if request_below(){
		} else if request_here() {
		} else if  request_above(){
		} else {
			//sett den i idle
		}
	case 0 //
		if request_here(){
		} else if request_above() {
		} else if request_below() {
		} else {
			//sett den i idle
		}
	case
	}
	return true
 //vite når det kommer en new order
 //
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
