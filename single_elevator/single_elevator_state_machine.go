package singleElevator

import (
	"PROJECT-GROUP-10/config"
	"PROJECT-GROUP-10/elevio"
	"PROJECT-GROUP-10/networking"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type elevator_state int

const (
	idle elevator_state = iota
	moving
	doorOpen
)

type update_elevator_node struct {
	command      string
	update_value int
}

var add_order_to_node, remove_order_from_node update_elevator_node
var current_state elevator_state
var last_floor int
var elevator_stuck, restoring_cab_calls, elevator_door_blocked bool
var cabCalls [4]bool

func SingleElevatorFSM(
	ch_drv_floors <-chan int,
	ch_elevator_has_arrived chan bool,
	ch_obstr_detected <-chan bool,
	ch_new_order chan bool,
	ch_drv_stop <-chan bool,
	ch_req_ID chan int,
	ch_req_data chan networking.Elevator_node, //Should be write only
	ch_write_data chan networking.Elevator_node,
	ch_hallCallsTot_updated <-chan [config.NUMBER_OF_FLOORS]networking.HallCall,
	ch_net_command chan elevio.ButtonEvent,
	ch_self_command chan elevio.ButtonEvent,
	ch_take_calls chan int,
) {
	ch_door_timer_out := make(chan bool)
	ch_door_timer_reset := make(chan bool)
	ch_elev_stuck_timer_out := make(chan bool)
	ch_elev_stuck_timer_start := make(chan bool)
	ch_elev_stuck_timer_stop := make(chan bool)
	ch_update_elevator_node_placement := make(chan string, 4)
	ch_update_elevator_node_order := make(chan update_elevator_node, 2)
	ch_remove_elevator_node_order := make(chan update_elevator_node, 2)
	init_elevator()
	go Hall_order(ch_new_order, ch_elevator_has_arrived, ch_net_command, ch_self_command, ch_update_elevator_node_order, ch_remove_elevator_node_order)
	go OpenAndCloseDoorsTimer(ch_door_timer_out, ch_door_timer_reset)
	go ElevatorStuckTimer(ch_elev_stuck_timer_out, ch_elev_stuck_timer_start, ch_elev_stuck_timer_stop)
	go MoveNearestFloor()
	go CheckIfElevatorHasArrived(ch_drv_floors, ch_elevator_has_arrived, ch_update_elevator_node_placement, ch_new_order)
	go Update_hall_lights(ch_hallCallsTot_updated)
	go Update_elevator_node(ch_req_ID, ch_req_data, ch_write_data, ch_update_elevator_node_placement, ch_update_elevator_node_order, ch_remove_elevator_node_order)
	for {
		select {
		case <-ch_new_order:
			switch current_state {
			case idle:
				if Request_next_action(elevator.direction) {
					elevio.SetMotorDirection(elevio.MotorDirection(elevator_command.direction))
					fmt.Printf("Moving to floor %+v\n", elevator_command.floor)
					ch_update_elevator_node_placement <- "direction"
					ch_update_elevator_node_placement <- "destination"
					ch_elev_stuck_timer_start <- true
					current_state = moving
				} else {
					elevio.SetMotorDirection(elevio.MotorDirection(0))
					ch_update_elevator_node_placement <- "direction"
				}
			case moving:
				fmt.Printf("Moving to floor %+v\n", elevator_command.floor)
				Request_next_action(elevator.direction)
			case doorOpen:
			}
		case <-ch_elevator_has_arrived:
			switch current_state {
			case idle:
				fmt.Printf("Elevator already here, opening door\n")
				elevio.SetDoorOpenLamp(true)
				ch_door_timer_reset <- true
				current_state = doorOpen
			case moving:
				fmt.Printf("Arrived at floor %+v\n", elevator_command.floor)
				elevio.SetMotorDirection(elevio.MD_Stop)
				ch_update_elevator_node_placement <- "direction"
				elevio.SetDoorOpenLamp(true)
				ch_door_timer_reset <- true
				ch_elev_stuck_timer_stop <- true
				Update_position(elevator_command.floor, elevator_command.direction, ch_remove_elevator_node_order)
				current_state = doorOpen
			}
		case <-ch_door_timer_out:
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
						ch_update_elevator_node_placement <- "direction"
						fmt.Printf("Moving to floor %+v\n", elevator_command.floor)
						ch_elev_stuck_timer_start <- true
						if elevator_command.floor == elevator.floor {
							current_state = doorOpen
							elevio.SetDoorOpenLamp(true)
							Update_position(elevator_command.floor, elevator_command.direction, ch_remove_elevator_node_order)
							ch_door_timer_reset <- true
						} else {
							current_state = moving
						}
					} else {
						fmt.Printf("No new orders, returning to idle\n")
						current_state = idle
						elevator_command.direction = 0
						ch_update_elevator_node_placement <- "direction"
					}
				}
			}
		case msg := <-ch_drv_stop:
			if msg {
				elevio.SetMotorDirection(elevio.MD_Stop)
				ch_update_elevator_node_placement <- "direction"
				ch_update_elevator_node_placement <- "set error"
			} else {
				elevio.SetMotorDirection(elevio.MotorDirection(elevator_command.direction))
				ch_update_elevator_node_placement <- "direction"
				ch_update_elevator_node_placement <- "reset error"
			}
		case msg := <-ch_obstr_detected:
			if msg {
				elevator_door_blocked = true
				ch_update_elevator_node_placement <- "set error"
			} else {
				elevator_door_blocked = false
				ch_update_elevator_node_placement <- "reset error"
			}
		case <-ch_elev_stuck_timer_out:
			fmt.Println("Elevator: I'm stuck, please call Vakt & Service")
			ch_take_calls <- config.ELEVATOR_ID
			ch_update_elevator_node_placement <- "set error"
			remove_order_from_node.command = "wipe orders"
			remove_order_from_node.update_value = 0
			ch_remove_elevator_node_order <- remove_order_from_node
		}
	}
}

func CheckIfElevatorHasArrived(ch_drv_floors <-chan int,
	ch_elevator_has_arrived chan bool,
	ch_update_elevator_node_placement chan string, ch_new_order chan bool) {
	for {
		msg := <-ch_drv_floors
		elevator.floor = msg
		ch_update_elevator_node_placement <- "floor"
		elevio.SetFloorIndicator(msg)
		if last_floor == -1 {
			last_floor = elevator.floor
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator_stuck = false
		}
		if restoring_cab_calls {
			ch_new_order <- true
		}
		if elevator_command.floor == msg && last_floor != elevator_command.floor {
			last_floor = elevator_command.floor
			ch_elevator_has_arrived <- true
		}
	}
}

func Update_hall_lights(ch_hallCallsTot_updated <-chan [config.NUMBER_OF_FLOORS]networking.HallCall) {
	for {
		msg := <-ch_hallCallsTot_updated
		for i := 0; i < config.NUMBER_OF_FLOORS; i++ {
			if msg[i].Up {
				elevio.SetButtonLamp(elevio.BT_HallUp, i, true)
			} else {
				elevio.SetButtonLamp(elevio.BT_HallUp, i, false)
			}
			if msg[i].Down {
				elevio.SetButtonLamp(elevio.BT_HallDown, i, true)
			} else {
				elevio.SetButtonLamp(elevio.BT_HallDown, i, false)
			}
		}
	}
}

func Update_elevator_node(
	ch_req_ID chan int,
	ch_req_data, ch_write_data chan networking.Elevator_node,
	ch_update_elevator_node_placement chan string,
	ch_update_elevator_node_order chan update_elevator_node,
	ch_remove_elevator_node_order chan update_elevator_node,
) {
	for {
		updated_elevator_node := networking.Node_get_data(
			config.ELEVATOR_ID,
			ch_req_ID,
			ch_req_data)
		select {
		case msg := <-ch_update_elevator_node_placement:
			switch msg {
			case "floor":
				updated_elevator_node.Floor = elevator.floor
			case "direction":
				updated_elevator_node.Direction = elevator_command.direction
			case "destination":
				updated_elevator_node.Destination = elevator_command.floor
			case "set_error":
				updated_elevator_node.Status = 1
			case "reset_error":
				updated_elevator_node.Status = 0
			}
			updated_elevator_node.ID = config.ELEVATOR_ID
			ch_write_data <- updated_elevator_node
		case msg := <-ch_update_elevator_node_order:
			switch msg.command {
			case "update order up":
				updated_elevator_node.HallCalls[msg.update_value].Up = true
			case "update order down":
				updated_elevator_node.HallCalls[msg.update_value].Down = true
			}
			updated_elevator_node.ID = config.ELEVATOR_ID
			ch_write_data <- updated_elevator_node
		case msg := <-ch_remove_elevator_node_order:
			switch msg.command {
			case "remove order up":
				updated_elevator_node.HallCalls[msg.update_value].Up = false
			case "remove order down":
				updated_elevator_node.HallCalls[msg.update_value].Down = false
			case "wipe orders":
				for i := 0; i < config.NUMBER_OF_FLOORS; i++ {
					updated_elevator_node.HallCalls[i].Up = false
					updated_elevator_node.HallCalls[i].Down = false
					floor[i].down = false
					floor[i].up = false
				}
			}
			updated_elevator_node.ID = config.ELEVATOR_ID
			ch_write_data <- updated_elevator_node
		}
	}
}

func init_elevator() {
	for i := 0; i < config.NUMBER_OF_FLOORS; i++ {
		elevio.SetButtonLamp(0, i, false)
		elevio.SetButtonLamp(1, i, false)
		elevio.SetButtonLamp(2, i, false)
	}
	elevator.direction = 0
	elevator.floor = 0
	current_state = idle
	last_floor = -1
	restoring_cab_calls = false
	elevator_stuck = true
	file, _ := os.OpenFile("cabcalls.json", os.O_RDWR|os.O_CREATE, 0666)
	bytes := make([]byte, 50)
	n, _ := file.ReadAt(bytes, 0)
	_ = json.Unmarshal(bytes[:n], &cabCalls)
	for i := 0; i < config.NUMBER_OF_FLOORS; i++ {
		if cabCalls[i] {
			floor[i].cab = true
			elevio.SetButtonLamp(2, i, true)
			restoring_cab_calls = true
		}
	}
}

func MoveNearestFloor() {
	time.Sleep(1 * time.Second)
	if elevator_stuck {
		elevio.SetMotorDirection(elevio.MD_Down)
	}
}
