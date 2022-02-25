package main

import (
	"../elevio"
	"fmt"
)

var current_floor int

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	//var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 || a == 0 {
				elevio.SetMotorDirection(elevio.MD_Stop)
			}

			//elevio.SetMotorDirection(d)
			current_floor = a

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			/*
				if a {
					elevio.SetMotorDirection(elevio.MD_Stop)
				} else {
					elevio.SetMotorDirection(d)
				}
			*/
		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
		go elevator_goTo(0, numFloors)
	}

}

func elevator_goTo(iDestination int, numFloors int) (reached bool) {
	//Check for illigal destination
	if iDestination > numFloors-1 {
		iDestination = numFloors - 1
	} else if iDestination < 0 {
		iDestination = 0
	}
	reached = false
	//Go to the destination and stop
	if current_floor < iDestination {
		elevio.SetMotorDirection(elevio.MD_Up)
	} else if current_floor > iDestination {
		elevio.SetMotorDirection(elevio.MD_Down)
	} else {
		elevio.SetMotorDirection(elevio.MD_Stop)
		reached = true
	}
	return reached
}
