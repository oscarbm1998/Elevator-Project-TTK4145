package main

import (
	"fmt"
	"time"

	"../config"
)

func OpenAndCloseDoorsTimer(ch_door_timer_out chan<- bool, ch_door_timer_reset <-chan bool) {
	//Initiatie the timer
	timer := time.NewTimer(config.ELEVATOR_DOOR_OPEN_TIME)
	timer.Stop()

	for {
		select {
		case <-timer.C:
			fmt.Println("Elevator: Doors closed")
			ch_door_timer_out <- true
		case <-ch_door_timer_reset:
			fmt.Println("Elevator: Opening doors")
			timer.Stop()
			timer.Reset(config.ELEVATOR_DOOR_OPEN_TIME)
		}
	}
}

func ElevatorStuckTimer(ch_elev_stuck_timer_out chan<- bool, ch_elev_stuck_timer_start <-chan bool, ch_elev_stuck_timer_stop <-chan bool) {
	timer := time.NewTimer(config.ELEVATOR_STUCK_TIMOUT)
	timer.Stop()

	for {
		select {
		case <-timer.C:
			fmt.Println("Elevator: I'm stuck, please call Vakt & Service")
			ch_elev_stuck_timer_out <- true
		case <-ch_elev_stuck_timer_start:
			timer.Stop()
			timer.Reset(config.ELEVATOR_STUCK_TIMOUT)
		case <-ch_elev_stuck_timer_stop:
			timer.Stop()
		}
	}
}
