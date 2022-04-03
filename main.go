package main

import (
	//"fmt"
	config "PROJECT-GROUP-10/config"
	elevio "PROJECT-GROUP-10/elevio"
	networking "PROJECT-GROUP-10/networking"
	ordering "PROJECT-GROUP-10/ordering"
	singleElevator "PROJECT-GROUP-10/single_elevator"
	"flag"
	//"net"
)

//var current_floor int

func main() {
	flag.IntVar(&config.ELEVATOR_ID, "id", 1, "id of this peer")
	flag.StringVar(&config.ELEVATOR_LOCAL_HOST, "host", "localhost:15657", "host")
	flag.Parse()

	if config.ELEVATOR_ID > config.NUMBER_OF_ELEVATORS {
		panic("Illegal ID, must be within the range of defined number of elevators")
	}
	numFloors := 4
	elevio.Init(config.ELEVATOR_LOCAL_HOST, numFloors)

	//var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	ch_drv_buttons := make(chan elevio.ButtonEvent) //Give
	ch_drv_floors := make(chan int)                 //Some
	ch_obstr_detected := make(chan bool)            //Explenation
	ch_drv_stop := make(chan bool)                  //of what
	ch_elevator_has_arrived := make(chan bool)      //do

	ch_net_command := make(chan elevio.ButtonEvent, 6)   //pls
	ch_self_command := make(chan elevio.ButtonEvent, 10) //ppl

	//Networking
	ch_new_order := make(chan bool)                                                    //Message here if there is a new order
	ch_hallCallsTot_updated := make(chan [config.NUMBER_OF_FLOORS]networking.HallCall) //Serviced hallcalls for light panel
	ch_take_calls := make(chan int)                                                    //Returns ID of a dead elevator to be handed over
	ch_new_data := make(chan int)                                                      //The data handler will send the ID here if new data from HB to cost function
	//Multiple data modueles to avoid a deadlock
	var ch_req_ID [3]chan int
	var ch_req_data, ch_write_data [3]chan networking.Elevator_node
	for i := range ch_req_ID {
		ch_req_ID[i] = make(chan int)                          //Send the ID of the elevator you want data from here
		ch_req_data[i] = make(chan networking.Elevator_node)   //... the data will be returned on this channel
		ch_write_data[i] = make(chan networking.Elevator_node) //Write data on this channel
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
		ch_net_command,
		ch_self_command,
		ch_take_calls)
	go ordering.Pass_to_network(
		ch_drv_buttons,
		ch_new_order,
		ch_take_calls,
		ch_self_command,
		ch_new_data,
		ch_req_ID[2],
		ch_req_data[2],
	)
	go networking.Main(ch_req_ID, ch_new_data, ch_take_calls, ch_req_data, ch_write_data, ch_net_command, ch_hallCallsTot_updated)
	select {}
}
