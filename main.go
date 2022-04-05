package main

import (
	config "PROJECT-GROUP-10/config"
	elevio "PROJECT-GROUP-10/elevio"
	lift_assigner "PROJECT-GROUP-10/lift_assigner"
	networking "PROJECT-GROUP-10/networking"
	singleElevator "PROJECT-GROUP-10/single_elevator"
	"flag"
)

func main() {
	flag.IntVar(&config.ELEVATOR_ID, "id", 1, "id of this peer")
	flag.StringVar(&config.ELEVATOR_LOCAL_HOST, "host", "localhost:15657", "host")
	flag.Parse()

	if config.ELEVATOR_ID > config.NUMBER_OF_ELEVATORS {
		panic("Illegal ID, must be within the range of defined number of elevators")
	}

	//Elevator driver
	elevio.Init()
	ch_drv_buttons := make(chan elevio.ButtonEvent, 6)
	ch_drv_floors := make(chan int)
	ch_obstr_detected := make(chan bool)
	ch_drv_stop := make(chan bool)
	ch_elevator_has_arrived := make(chan bool)
	ch_command_elev := make(chan elevio.ButtonEvent, 10)

	//Networking
	ch_new_order := make(chan bool)
	ch_hallCallsTot_updated := make(chan [config.NUMBER_OF_FLOORS]networking.HallCall)
	ch_take_calls := make(chan int)
	ch_new_data := make(chan int)

	//Multiple data modueles to avoid a deadlock
	var ch_req_ID [3]chan int
	var ch_req_data, ch_write_data [3]chan networking.Elevator_node
	for i := range ch_req_ID {
		ch_req_ID[i] = make(chan int)
		ch_req_data[i] = make(chan networking.Elevator_node)
		ch_write_data[i] = make(chan networking.Elevator_node)
	}

	go elevio.PollButtons(ch_drv_buttons)
	go elevio.PollFloorSensor(ch_drv_floors)
	go elevio.PollObstructionSwitch(ch_obstr_detected)
	go elevio.PollStopButton(ch_drv_stop)
	go singleElevator.SingleElevatorFSM(
		ch_drv_floors,
		ch_elevator_has_arrived,
		ch_obstr_detected,
		ch_new_order,
		ch_drv_stop,
		ch_req_ID[1],
		ch_req_data[1],
		ch_write_data[1],
		ch_hallCallsTot_updated,
		ch_command_elev,
		ch_take_calls)
	go lift_assigner.Pass_to_network(
		ch_drv_buttons,
		ch_new_order,
		ch_take_calls,
		ch_command_elev,
		ch_new_data,
		ch_req_ID[2],
		ch_req_data[2],
	)
	go networking.Main(
		ch_req_ID,
		ch_new_data,
		ch_take_calls,
		ch_req_data,
		ch_write_data,
		ch_command_elev,
		ch_hallCallsTot_updated)
	select {}
}

/*********************************
		   Hello there
			───▄▄▄
			─▄▀░▄░▀▄
			─█░█▄▀░█
			─█░▀▄▄▀█▄█▄▀
			▄▄█▄▄▄▄███▀

*********************************/
