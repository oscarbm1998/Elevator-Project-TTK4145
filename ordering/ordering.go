package ordering

import (
	"PROJECT-GROUP-10/config"
	elevio "PROJECT-GROUP-10/elevio"
	networking "PROJECT-GROUP-10/networking"
	"math"
)

type scoreboard struct {
	elevator  int
	placement int
}

var score [config.NUMBER_OF_ELEVATORS]scoreboard

var elev_overview [config.NUMBER_OF_ELEVATORS]networking.Elevator_node

var button_calls elevio.ButtonEvent

func heartbeat_monitor( //checks and alerts the system whenever a heartbeat ping occurs
	ch_new_data chan int,
	ch_req_ID chan int,
	ch_req_data chan networking.Elevator_node,
) {
	for {
		select {
		case id := <-ch_new_data:
			for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ {
				lighthouse := networking.Node_get_data(id, ch_req_ID, ch_req_data)
				//fmt.Println("Lighthouse: " + strconv.Itoa(id))
				elev_overview[i] = lighthouse
			}
		}
	}
}

func sorting() {
	//sorting
	for p := 0; p < config.NUMBER_OF_ELEVATORS; p++ { //runs thrice
		var roundbest_index int                           //the strongest placement for this round
		var bestscore int                                 //the strongest placement for this round
		for i := p; i < config.NUMBER_OF_ELEVATORS; i++ { //ignores the stuff that has already been positioned
			if score[i].elevator > bestscore { //if the score surpasses the others
				roundbest_index = i           //sets the new index
				bestscore = score[i].elevator //sets the new best score
			}
		}
		score[p].placement = roundbest_index
	}
}

//meldigen som infoer victors modul om hva som skal sendes (main)
func Pass_to_network(
	ch_drv_buttons chan elevio.ButtonEvent,
	ch_new_order chan bool,
	ch_take_calls chan int,
	ch_self_command chan elevio.ButtonEvent,
	ch_new_data chan int,
	ch_req_ID chan int,
	ch_req_data chan networking.Elevator_node,
) {
	go heartbeat_monitor(
		ch_new_data,
		ch_req_ID,
		ch_req_data,
	)

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
			sorting()
			for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ { //will automatically cycle the scoreboard and attempt to send from best to worst
				if elev_overview[score[i].placement].ID == config.ELEVATOR_ID { //if the winning ID is the elevators own
					ch_self_command <- <-ch_drv_buttons
					break
				} else {
					if networking.Send_command(elev_overview[score[i].placement].ID, a.Floor, dir) {
						break
					}
				}
			}

		case death_id := <-ch_take_calls:
			for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ { //finds the elevator that has died
				if elev_overview[i].ID == death_id { //found the elevator
					for e := 0; e < 6; e++ { //checks all calls
						var dir int
						var floor int
						if elev_overview[i].HallCalls[e] == 1 {
							if e%2 == 0 { //the number is even so the dir is up
								dir = 1
								floor = e / 2
								master_tournament(floor, 1)
								//this shit may cause errors as i am unshure if everyone is cool with floors starting at 0
							} else { //the number is odd so the dir is down
								dir = -1
								floor = (e - 1) / 2
								master_tournament(floor, -1)
							}
							sorting()
							for c := 0; c < config.NUMBER_OF_ELEVATORS; c++ { //will automatically cycle the scoreboard and attempt to send from best to worst
								if elev_overview[score[c].placement].ID == config.ELEVATOR_ID { //if the winning ID is the elevators own
									ch_self_command <- <-ch_drv_buttons
									break
								} else {
									if networking.Send_command(elev_overview[score[c].placement].ID, floor, dir) {
										break
										/*********************************
										*		Welcome to Hell			 *
										*********************************/
									}
								}
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
	for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ {
		score[i].elevator = 0
		score[i].placement = 0
	}
	//filters out the nonworking and scores them
	for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ { //cycles shafts
		if !(elev_overview[i].Status == 404) {
			//direction scoring
			if direction == elev_overview[i].Direction {
				score[i].elevator += 3
			}
			//placement scoring (with alot of conversion)
			score[i].elevator += int(math.Abs(float64(elev_overview[i].Floor) - float64(floor)))
		}
	}
}
