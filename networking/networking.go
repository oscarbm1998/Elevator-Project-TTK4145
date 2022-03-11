package networking

import (
	config "PROJECT-GROUP-10/config"
)

type elevator_node struct {
	last_seen string
	ID        int
	direction int
	floor     int
	status    int
}

var elevator_nodes [config.NUMBER_OF_ELEVATORS]elevator_node
