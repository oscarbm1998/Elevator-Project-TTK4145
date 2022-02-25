package singleElevator

import (
	"time"

	"../config"
)

func OpenAndCloseDoorsTimer(door_timer_out_channel chan<- bool, door_timer_reset_channel <-chan bool) {
	//Initiatie the timer
	timer := time.NewTimer(config.ELEVATOR_DOOR_OPEN_TIME)
	timer.Stop()

	for {
		select {
		case <-timer.C:
			door_timer_out_channel <- true
		case <-door_timer_reset_channel:
			timer.Stop()
			timer.Reset(config.ELEVATOR_DOOR_OPEN_TIME)
		}
	}
}

func ElevatorStuckTimer(elev_stuck_timer_out_ch chan<- bool, elev_stuck_timer_start_ch <-chan bool, elev_stuck_timer_stop_ch <-chan bool) {
	timer := time.NewTimer(config.ELEVATOR_STUCK_TIMOUT)
	timer.Stop()

	for {
		select {
		case <-timer.C:
			elev_stuck_timer_out_ch <- true
		case <-elev_stuck_timer_start_ch:
			timer.Stop()
			timer.Reset(config.ELEVATOR_STUCK_TIMOUT)
		case <-elev_stuck_timer_stop_ch:
			timer.Stop()
		}
	}
}
