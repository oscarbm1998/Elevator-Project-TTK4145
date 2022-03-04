package main

import (
	singleElevator "PROJECT-GROUP-10/single_elevator"
	"fmt"

	"PROJECT-GROUP-10/elevio"
)

var current_floor int

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	//var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	ch_drv_buttons := make(chan elevio.ButtonEvent)
	ch_drv_floors := make(chan int)
	ch_drv_obstr := make(chan bool)
	ch_drv_stop := make(chan bool)
	ch_elevator_has_arrived := make(chan bool)
	ch_new_order := make(chan bool)
	ch_clear_order := make(chan bool)

	go elevio.PollButtons(ch_drv_buttons)
	go elevio.PollFloorSensor(ch_drv_floors)
	go elevio.PollObstructionSwitch(ch_drv_obstr)
	go elevio.PollStopButton(ch_drv_stop)
	go singleElevator.SingleElevatorFSM(ch_drv_floors, ch_elevator_has_arrived, ch_drv_obstr, ch_new_order)
	go singleElevator.Hall_order(ch_drv_buttons, ch_new_order, ch_clear_order, ch_drv_floors)

	for {
		select {
		case a := <-ch_drv_floors:
			fmt.Printf("Current floor is %+v\n", a)
		case <-ch_drv_obstr:
			//Lag noe her som sier at hvis den er trykket inn, stop
			//Når knappen ikke er trykket inn lenger, resume direction
		case a := <-ch_drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}

}
