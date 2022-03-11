package ordering

import (
	elevio "PROJECT-GROUP-10/elevio"
	networking "PROJECT-GROUP-10/networking"
	"math"
)

const shaft_ammount = 3

type scoreboard struct {
	elevator [shaft_ammount]int
	podium   [shaft_ammount]int
}

var score scoreboard

//meldigen som infoer victors modul om hva som skal sendes
func pass_to_network(
	ch_drv_buttons chan elevio.ButtonEvent,
	ch_new_order chan scoreboard,
) {
	for {
		select {
		case a := <-ch_drv_buttons:
			master_tournament(a.Floor, int(a.Button))
			ch_new_order <- score
		}
	}
}

func master_tournament(floor int, direction int) { //finds the most lucrative elevator
	//resets scoring
	for i := 0; i < shaft_ammount; i++ {
		score.elevator[i] = 0
		score.podium[i] = 0
	}
	//filters out the nonworking and scores them
	for i := 0; i < shaft_ammount; i++ { //cycles shafts
		if !(networking.Elevator_nodes[i].Status == 404) {
			//direction scoring
			if direction == networking.Elevator_nodes[i].Direction {
				score.elevator[i] += 5
			}
			//placement scoring (with alot of conversion)
			score.elevator[i] += int(math.Abs(float64(networking.Elevator_nodes[i].Floor) - float64(floor)))
		}
	}
	//decides wich one is the best suited through simple sorting algo
	for i := 0; i < 1; i++ { //runs twice to ensure proper sorting
		for i := 0; i < shaft_ammount-1; i++ {
			if score.elevator[i] < score.elevator[i+1] {
				score.podium[i] = i + 1
				score.podium[i+1] = i
			}
		}
	}
}
