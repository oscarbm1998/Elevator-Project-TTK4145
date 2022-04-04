package singleElevator

import (
	"PROJECT-GROUP-10/config"
	"PROJECT-GROUP-10/elevio"
	"encoding/json"
	"os"
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
var elevator elevator_status
var elevator_command elevator_status

func Hall_order(
	ch_new_order chan bool,
	ch_elevator_has_arrived chan bool,
	ch_net_command chan elevio.ButtonEvent,
	ch_self_command chan elevio.ButtonEvent,
	ch_update_elevator_node_order chan update_elevator_node,
	ch_remove_elevator_node_order chan update_elevator_node,
) {
	for {
		select {
		case a := <-ch_net_command:
			if ((floor[a.Floor].up && a.Button == 0) || (floor[a.Floor].down && a.Button == 1) || floor[a.Floor].cab || (a.Floor == elevator.floor)) && current_state != moving {
				ch_elevator_has_arrived <- true
			} else {
				switch a.Button {
				case elevio.BT_HallUp: //opp
					floor[a.Floor].up = true
					add_order_to_node.command = "update order up"
					add_order_to_node.update_value = a.Floor
					ch_update_elevator_node_order <- add_order_to_node
				case elevio.BT_HallDown: //ned
					floor[a.Floor].down = true
					add_order_to_node.command = "update order down"
					add_order_to_node.update_value = a.Floor
					ch_update_elevator_node_order <- add_order_to_node
				}
				if ((floor[a.Floor].up && a.Button == 0) || (floor[a.Floor].down && a.Button == 1) || floor[a.Floor].cab || (a.Floor == elevator.floor)) && current_state != moving {
					ch_new_order <- true //forteller at en ny order er tilgjengelig
				}
			}
		case a := <-ch_self_command:
			if ((floor[a.Floor].up && a.Button == 0) || (floor[a.Floor].down && a.Button == 1) || floor[a.Floor].cab || (a.Floor == elevator.floor)) && current_state != moving {
				ch_elevator_has_arrived <- true
			} else {
				switch a.Button {
				case elevio.BT_HallUp:
					floor[a.Floor].up = true
					add_order_to_node.command = "update order up"
					add_order_to_node.update_value = a.Floor
					ch_update_elevator_node_order <- add_order_to_node
					elevio.SetButtonLamp(elevio.BT_HallUp, a.Floor, true)
				case elevio.BT_HallDown:
					floor[a.Floor].down = true
					add_order_to_node.command = "update order down"
					add_order_to_node.update_value = a.Floor
					ch_update_elevator_node_order <- add_order_to_node
					elevio.SetButtonLamp(elevio.BT_HallDown, a.Floor, true)
				case elevio.BT_Cab:
					floor[a.Floor].cab = true
					elevio.SetButtonLamp(elevio.BT_Cab, a.Floor, true)
					file, _ := os.OpenFile("cabcalls.json", os.O_RDWR|os.O_CREATE, 0666)
					cabCalls[a.Floor] = true
					bytes, _ := json.Marshal(cabCalls)
					file.Truncate(0)
					file.WriteAt(bytes, 0)
					file.Close()
				}
				if ((floor[a.Floor].up && a.Button == 0) || (floor[a.Floor].down && a.Button == 1) || floor[a.Floor].cab || (a.Floor == elevator.floor)) && current_state != moving {
					ch_new_order <- true
				}
			}
		}
	}
}

func Remove_order(level int, direction int, ch_remove_elevator_node_order chan update_elevator_node) {
	floor[level].cab = false
	elevio.SetButtonLamp(2, level, false)
	file, _ := os.OpenFile("cabcalls.json", os.O_RDWR|os.O_CREATE, 0666)
	cabCalls[level] = false
	bytes, _ := json.Marshal(cabCalls)
	file.Truncate(0)
	file.WriteAt(bytes, 0)
	file.Close()
	if direction == int(elevio.MD_Up) {
		if !floor[level].up {
			floor[level].down = false
			remove_order_from_node.command = "remove order down"
			remove_order_from_node.update_value = level
			ch_remove_elevator_node_order <- remove_order_from_node
			elevio.SetButtonLamp(elevio.BT_HallDown, level, false)
		} else {
			floor[level].up = false
			remove_order_from_node.command = "remove order up"
			remove_order_from_node.update_value = level
			ch_remove_elevator_node_order <- remove_order_from_node
			elevio.SetButtonLamp(elevio.BT_HallUp, level, false)
		}
	} else if direction == int(elevio.MD_Down) {
		if !floor[level].down {
			floor[level].up = false
			remove_order_from_node.command = "remove order up"
			remove_order_from_node.update_value = level
			ch_remove_elevator_node_order <- remove_order_from_node
			elevio.SetButtonLamp(elevio.BT_HallUp, level, false)
		} else {
			floor[level].down = false
			remove_order_from_node.command = "remove order down"
			remove_order_from_node.update_value = level
			ch_remove_elevator_node_order <- remove_order_from_node
			elevio.SetButtonLamp(elevio.BT_HallDown, level, false)
		}
	} else if direction == int(elevio.MD_Stop) {
		if !floor[level].down {
			floor[level].up = false
			remove_order_from_node.command = "remove order up"
			remove_order_from_node.update_value = level
			ch_remove_elevator_node_order <- remove_order_from_node
			elevio.SetButtonLamp(elevio.BT_HallUp, level, false)
		} else {
			floor[level].down = false
			remove_order_from_node.command = "remove order down"
			remove_order_from_node.update_value = level
			ch_remove_elevator_node_order <- remove_order_from_node
			elevio.SetButtonLamp(elevio.BT_HallDown, level, false)
		}
	}
}

func request_above() bool {
	for i := elevator.floor + 1; i < config.NUMBER_OF_FLOORS; i++ {
		if floor[i].up || floor[i].down || floor[i].cab {
			elevator_command.floor = i
			elevator_command.direction = int(elevio.MD_Up)
			return true
		}
	}
	return false
}

func request_here() bool {
	if floor[elevator.floor].up || floor[elevator.floor].down || floor[elevator.floor].cab {
		elevator_command.floor = elevator.floor
		elevator_command.direction = 0
		return true
	}
	return false
}

func request_below() bool {
	for i := elevator.floor - 1; i >= 0; i-- {
		if floor[i].down || floor[i].up || floor[i].cab {
			elevator_command.floor = i
			elevator_command.direction = int(elevio.MD_Down)
			return true
		}
	}
	return false
}

func Request_next_action(direction int) bool {
	switch direction {
	case int(elevio.MD_Up):
		if request_above() {
			return true
		} else if request_here() {
			return true
		} else if request_below() {
			return true
		}

	case elevio.MD_Down:
		if request_below() {
			return true
		} else if request_here() {
			return true
		} else if request_above() {
			return true
		}

	case elevio.MD_Stop:
		if request_above() {
			return true
		} else if request_here() {
			return true
		} else if request_below() {
			return true
		}
	}
	return false
}

func Update_position(level int, direction int, ch_remove_elevator_node_order chan update_elevator_node) {
	elevator.floor = level
	elevator.direction = direction
	Remove_order(level, direction, ch_remove_elevator_node_order)
}
