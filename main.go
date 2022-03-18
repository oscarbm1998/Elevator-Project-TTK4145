package main

import (
	//"fmt"
	networking "PROJECT-GROUP-10/networking"
	//"net"
)

//singleElevator "PROJECT-GROUP-10/single_elevator"

//"PROJECT-GROUP-10/elevio"

//var current_floor int

func main() {
	/*
		numFloors := 4

		elevio.Init("localhost:15657", numFloors)

		//var d elevio.MotorDirection = elevio.MD_Up
		//elevio.SetMotorDirection(d)

		ch_drv_buttons := make(chan elevio.ButtonEvent)
		ch_drv_floors := make(chan int)
		ch_obstr_detected := make(chan bool)
		ch_drv_stop := make(chan bool)
		ch_elevator_has_arrived := make(chan bool)
		ch_new_order := make(chan bool)

		go elevio.PollButtons(ch_drv_buttons)
		go elevio.PollFloorSensor(ch_drv_floors)
		go elevio.PollObstructionSwitch(ch_obstr_detected)
		go elevio.PollStopButton(ch_drv_stop)
		go singleElevator.SingleElevatorFSM(ch_drv_floors, ch_elevator_has_arrived, ch_obstr_detected, ch_new_order, ch_drv_stop)
		go singleElevator.Hall_order(ch_drv_buttons, ch_new_order)
	*/
	//Networking
	ch_req_ID := make(chan int)                          //Send the ID of the elevator you want data from here
	ch_req_data := make(chan networking.Elevator_node)   //... the data will be returned on this channel
	ch_write_data := make(chan networking.Elevator_node) //Write data on this channel
	ch_new_data := make(chan int)                        //The data handler will send the ID here if new data from HB to cost function
	ch_ext_dead := make(chan int)                        //Returns ID of a dead elevator
	go networking.Networking_main(ch_req_ID, ch_new_data, ch_ext_dead, ch_req_data, ch_write_data)
}
