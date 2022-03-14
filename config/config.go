package config

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
	ELEVATOR_DOOR_OPEN_TIME = time.Second * 3

	REMOVE_OLD_ORDER_TIME = time.Second * 2

	//Networking
	HEARTBEAT_TIME        = time.Second * 1
	HEARTBEAT_TIME_OUT    = time.Second * 10
	HEARTBEAT_PORT        = 7171
	COMMAND_SEND_PORT     = 1030
	COMMAND_REC_PORT      = 1040 + ELEVATOR_ID - 1
	COMMAND_READBACK_PORT = 1050 + ELEVATOR_ID - 1
	NUMBER_OF_ELEVATORS   = 2
)
