package singleElevator

import (
	"PROJECT-GROUP-10/elevio"
)

const floor_ammount int = 4

type elevator_status struct {
	floor     int
	direction int //1 up -1 down 0 idle
}
type dir struct {
	up   bool
	down bool
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
	if direction == 0 {
		floor[level].direction.up = false
		elevio.SetButtonLamp(0, level, false)
	} else if direction == 1 {
		floor[level].direction.down = false
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
			case 0: //opp
				floor[a.Floor].hall_call = 1
				floor[a.Floor].direction.up = true
				hall_calls()
			case 1: //ned
				floor[a.Floor].hall_call = 1
				floor[a.Floor].direction.down = true
				hall_calls()
			case 2: //cab call
				floor[a.Floor].cab_call = 1
				if Cab_calls() {
				}
			}
			ch_new_order <- true //forteller at en ny order er tilgjengelig
		}
	}
}

//algo for utside
func hall_calls() {
	switch elevator.direction {
	case 1:
		for i := elevator.floor; i <= floor_ammount; i++ {
			if floor[i].hall_call == 1 {
				elevator_command.floor = i
				break
			}
		}
	case -1:
		for i := elevator.floor; i <= 0; i-- {
			if floor[i].hall_call == 1 {
				elevator_command.floor = i
				break
			}
		}
	case 0:
		for i := 0; i <= floor_ammount; i++ {
			if floor[i].hall_call == 1 {
				elevator_command.floor = i
				if floor[i].hall_call > elevator.floor {
					elevator_command.direction = 1
				} else {
					elevator_command.direction = -1
				}
				break
			}
		}
	}
}

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
