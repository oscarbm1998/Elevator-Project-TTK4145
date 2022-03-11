package ordering

import (
	networking "PROJECT-GROUP-10/networking"
)

const floor_ammount = 4
const shaft_ammount = 3

type floor_info struct {
	up   bool
	down bool
	here bool
}

type complete_system struct {
	shaft     [shaft_ammount]floor_info
	placement [shaft_ammount]int
	active    [shaft_ammount]bool
}

type scoreboard struct {
	elevator []int
	winner   int
}

var elevators complete_system //the complete elevator system
var score scoreboard

func master_tournament(floor int, direction int) { //fonds the most lucrative elevator
	//resets scoring
	for i := 0; i < shaft_ammount; i++ {
		score.elevator[i] = 0
	}
	//filters out the nonworking and scores them
	for i := 0; i < shaft_ammount; i++ { //cycles shafts
		if networking.Elevator_nodes[i].ID == 404 {
			//direction scoring
			if direction == networking.Elevator_nodes[i].direction {
				score.elevator[i] += 5
			}
			//placement scoring
			score.elevator[i] += 10 - abs(floor-networking.Elevator_nodes[i].floor)
		}
	}

	//decides wich one is the best suited
	for i := 0; i < shaft_ammount; i++ {

	}

}
