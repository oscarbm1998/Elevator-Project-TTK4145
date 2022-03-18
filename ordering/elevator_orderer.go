package ordering

import (
	elevio "PROJECT-GROUP-10/elevio"
	networking "PROJECT-GROUP-10/networking"
	"math"
)

const shaft_ammount = 3

type scoreboard struct {
	elevator [shaft_ammount]int
}

var score scoreboard

var elev_overview [shaft_ammount]networking.Elevator_node

func heartbeat_monitor( //checks and alerts the system whenever a heartbeat ping occurs
	ch_new_data chan int,
	ch_req_ID chan int,
	ch_req_data chan networking.Elevator_node,
) {
	for {
		select {
		case id := <-ch_new_data:
			for i := 0; i < shaft_ammount; i++ {
				lighthouse := networking.Node_get_data(id, ch_req_ID, ch_req_data)
				elev_overview[i] = lighthouse
			}
		}
	}
}

func sorting(ignoreval int) int {
	//sorting
	var best_elevator_index int = 0
	var bestscore int = 0
	for i := 0; i < shaft_ammount; i++ {
		if score.elevator[i] > bestscore {
			if i != ignoreval {
				best_elevator_index = i
			}
		}
	}
	if networking.Send_command(elev_overview[best_elevator_index].ID, a.Floor, dir) {
	} else {
		best_elevator_index := sorting(best_elevator_index)
		networking.Send_command(elev_overview[best_elevator_index].ID, a.Floor, dir)
	}
	return best_elevator_index
}

//meldigen som infoer victors modul om hva som skal sendes
func pass_to_network(
	ch_drv_buttons chan elevio.ButtonEvent,
	ch_new_order chan scoreboard,
	ch_take_calls chan int,
) {
	for {
		select {
		case a := <-ch_drv_buttons:
			var dir int
			switch a.Button {
			case 0: //up
				dir = 1
				master_tournament(a.Floor, 1)
			case 1: //down
				dir = -1
				master_tournament(a.Floor, -1)
			case 2: //cab
				dir = 0
				master_tournament(a.Floor, 0)
			}
			//sends a command to the highest scoring elevator currently only works twice :|

		case death_id := <-ch_take_calls:
			for i := 0; i < shaft_ammount; i++ { //finds the elevator that has died
				if elev_overview[i].ID == death_id { //found the elevator
					for e := 0; e < 6; e++ { //checks all calls
						if elev_overview[i].HallCalls[e] == 1 {
							if e%2 == 0 { //the number is even so the dir is up
								master_tournament(e/2, 1)
								//this shit may cause errors as i am unshure if everyone is cool with floors starting at 0
							} else { //the number is odd so the dir is down
								master_tournament((e-1)/2, -1)
							}

						}
					}
				}
			}
		}
	}
}

func master_tournament(floor int, direction int) { //finds the most lucrative elevator
	//resets scoring
	for i := 0; i < shaft_ammount; i++ {
		score.elevator[i] = 0
	}
	//filters out the nonworking and scores them
	for i := 0; i < shaft_ammount; i++ { //cycles shafts
		if !(elev_overview[i].Status == 404) {
			//direction scoring
			if direction == elev_overview[i].Direction {
				score.elevator[i] += 5
			}
			//placement scoring (with alot of conversion)
			score.elevator[i] += int(math.Abs(float64(elev_overview[i].Floor) - float64(floor)))
		}
	}
}
