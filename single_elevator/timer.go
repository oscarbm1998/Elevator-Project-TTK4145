package singleElevator

import (
	"fmt"
	"time"
	config "PROJECT-GROUP-10/config"
)

func OpenAndCloseDoorsTimer(ch_door_timer_out chan<- bool, ch_door_timer_reset <-chan bool) {
	//Initiatie the timer
	timer := time.NewTimer(config.ELEVATOR_DOOR_OPEN_TIME)
	timer.Stop()

	for {
		select {
		case <-timer.C:
			ch_door_timer_out <- true
		case <-ch_door_timer_reset:
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

func array_leftshift(in [7]int, size int) (out [7]int) {
	for i := 0; i < 7; i++ {
		out[i] = in[i+1]
	}
	return out
}
