package config

import "time"

var ELEVATOR_ID int = 1

const (
	//ELEVATOR_ID            int    = 1
	SIMULATION             bool   = false
	SIMULATION_IP_AND_PORT string = ""
	NUMBER_OF_FLOORS              = 4
	NUMBER_OF_BUTTONS             = 3
	NUMBER_OF_HALL_BUTTONS        = 6

	UDP_S_R_PORT   = 20007
	UDP_PEERS_PORT = 30007

	ELEVATOR_STUCK_TIMOUT   = time.Second * 10
	ELEVATOR_DOOR_OPEN_TIME = time.Second * 3

	REMOVE_OLD_ORDER_TIME = time.Second * 2

	//Networking
	HEARTBEAT_TIME      = time.Second * 1
	HEARTBEAT_TIMEOUT   = time.Second * 3
	HEARTBEAT_PORT      = 7171
	COMMAND_PORT        = 7272
	COMMAND_RBC_PORT    = 7373
	REVIVE_PORT         = 7474
	NUMBER_OF_ELEVATORS = 2
)
