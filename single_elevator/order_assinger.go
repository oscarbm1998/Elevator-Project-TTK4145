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
var elevator elevator_status         //where elevator is
var elevator_command elevator_status //where elevator should go

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
				case elevio.BT_HallUp: //opp
					floor[a.Floor].up = true
					add_order_to_node.command = "update order up"
					add_order_to_node.update_value = a.Floor
					ch_update_elevator_node_order <- add_order_to_node
					elevio.SetButtonLamp(elevio.BT_HallUp, a.Floor, true)
				case elevio.BT_HallDown: //ned
					floor[a.Floor].down = true
					add_order_to_node.command = "update order down"
					add_order_to_node.update_value = a.Floor
					ch_update_elevator_node_order <- add_order_to_node
					elevio.SetButtonLamp(elevio.BT_HallDown, a.Floor, true)
				case elevio.BT_Cab: //cab call
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

func Remove_order(level int, direction int, ch_remove_elevator_node_order chan update_elevator_node) { //removes an order
	floor[level].cab = false              //removes here call as the elevator has arrived there
	elevio.SetButtonLamp(2, level, false) //turns off cab light
	file, _ := os.OpenFile("cabcalls.json", os.O_RDWR|os.O_CREATE, 0666)
	cabCalls[level] = false
	bytes, _ := json.Marshal(cabCalls)
	file.Truncate(0)
	file.WriteAt(bytes, 0)
	file.Close()
	if direction == int(elevio.MD_Up) { //if the direction is up or there are no orders below and orders above
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
		//disables the up direction
	} else if direction == int(elevio.MD_Down) { //if the direction is down or there are no orders above and orders below
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

func request_above() bool { //checks if there are any active calls above the elevator and updates the "command struct"
	for i := elevator.floor + 1; i < config.NUMBER_OF_FLOORS; i++ { //checks from the last known floor of the elevator to the top
		if floor[i].up || floor[i].down || floor[i].cab { //if a floor with call up is found
			elevator_command.floor = i                     //updates the command value
			elevator_command.direction = int(elevio.MD_Up) //sets the direction up just in case
			return true
		}
	}
	return false
}

func request_here() bool {
	if floor[elevator.floor].up || floor[elevator.floor].down || floor[elevator.floor].cab {
		elevator_command.floor = elevator.floor //updates the command value
		elevator_command.direction = 0          //sets the direction down just in case
		return true
	}
	return false
}

func request_below() bool { //checks if there are any active calls below the elevator and updates the "command struct"
	for i := elevator.floor - 1; i >= 0; i-- { //checks from the last known floor of the elevator to the botton
		if floor[i].down || floor[i].up || floor[i].cab { //if a floor with call down is found
			elevator_command.floor = i                       //updates the command value
			elevator_command.direction = int(elevio.MD_Down) //sets the direction down just in case
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
		} else if request_here() {
			return true
		} else if request_below() {
			return true
		}

	case elevio.MD_Down: //down
		if request_below() {
			return true
		} else if request_here() {
			return true
		} else if request_above() {
			return true
		}

	case elevio.MD_Stop: // here
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

func Update_position(level int, direction int, ch_remove_elevator_node_order chan update_elevator_node) { //Er denne nødvendig?
	elevator.floor = level
	elevator.direction = direction
	Remove_order(level, direction, ch_remove_elevator_node_order)
}
