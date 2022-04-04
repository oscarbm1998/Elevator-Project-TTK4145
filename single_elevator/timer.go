package singleElevator

import (
	config "PROJECT-GROUP-10/config"
	"time"
)

func OpenAndCloseDoorsTimer(ch_door_timer_out chan<- bool, ch_door_timer_reset <-chan bool) {
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
			ch_elev_stuck_timer_out <- true
		case <-ch_elev_stuck_timer_start:
			timer.Stop()
			timer.Reset(config.ELEVATOR_STUCK_TIMOUT)
		case <-ch_elev_stuck_timer_stop:
			timer.Stop()
		}
	}
}
