package singleElevator

const floor_ammount int = 4

type elevator_status struct {
	floor     int
	direction int //1 up -1 down 0 idle
}

type floor_info struct {
	hall_call bool
	cab_call  bool
}

var floor [floor_ammount]floor_info
var elevator elevator_status         //where elevator is
var Elevator_command elevator_status //where elevator should go

//algo for utside
func hall_calls() {
	switch elevator.direction {
	case 1:
		for i := elevator.floor; i <= floor_ammount; i++ {
			if floor[i].hall_call {
				elevator_command.floor = i
				break
			}
		}
	case -1:
		for i := elevator.floor; i <= 0; i-- {
			if floor[i].hall_call {
				elevator_command.floor = i
				break
			}
		}
	case 0:
		for i:= 0; i<= floor_ammount; i++
			if floor[i].hall_call {
				elevator_command.floor = i
				if floor[i].hall_call >> elevator.floor {
					elevator_command.direction = 1
				} else {
					elevator_command.direction = -1
				}
				break
			}
		}
	}
}

func cab_calls () {
	for i:= 0; i<= floor_ammount; i++ {
		if floor[i].cab_call {
			elevator_command.direction = i
			break
		}
	}
}
