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
	go networking.Networking_main()

	select {}
}
