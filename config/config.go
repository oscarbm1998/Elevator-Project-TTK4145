package config

import "time"

var ELEVATOR_ID int = 1
var ELEVATOR_LOCAL_HOST string = "localhost:15657"

const (
	SIMULATION             bool   = false
	SIMULATION_IP_AND_PORT string = ""
	NUMBER_OF_FLOORS              = 4
	NUMBER_OF_BUTTONS             = 3

	//Timers
	ELEVATOR_STUCK_TIMOUT   = time.Second * 10
	ELEVATOR_DOOR_OPEN_TIME = time.Second * 3

	//Networking
	HEARTBEAT_TIME      = time.Second * 1
	HEARTBEAT_TIMEOUT   = time.Second * 3
	HEARTBEAT_PORT      = 7171
	COMMAND_PORT        = 7272
	COMMAND_RBC_PORT    = 7373
	REVIVE_PORT         = 7474
	NUMBER_OF_ELEVATORS = 3
)
