package main

import "time"

const (
	ELEVATOR_ID            int    = 1
	SIMULATION             bool   = false
	SIMULATION_IP_AND_PORT string = ""
	NUMBER_OF_FLOORS              = 4
	NUMBER_OF_BUTTONS             = 3

	UDP_S_R_PORT   = 20007
	UDP_PEERS_PORT = 30007

	ELEVATOR_STUCK_TIMOUT   = time.Second * 10
	ELEVATOR_DOOR_OPEN_TIME = time.Second * 10

	REMOVE_OLD_ORDER_TIME = time.Second * 2
)
