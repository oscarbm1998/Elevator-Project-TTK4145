package ordering

import (
	"PROJECT-GROUP-10/config"
	elevio "PROJECT-GROUP-10/elevio"
	networking "PROJECT-GROUP-10/networking"
	"fmt"
	"math"
)

type score_tracker struct { //this struct keeps track of the score and what elevator the socre belongs to
	score           int
	elevator_number int
}

//uses the score tracker struct to make a sort of scoreboard the first array is the podium itself whilst the internal "elevator number" keeps track of which elevator is which
var placement [config.NUMBER_OF_ELEVATORS]score_tracker

//a copy of the elevator node struct to keep internal track of the elevators
var elev_overview [config.NUMBER_OF_ELEVATORS]networking.Elevator_node

//a translator for when we need to pass info between two channels
var button_calls elevio.ButtonEvent

//checks and alerts the system whenever a heartbeat ping occurs
func heartbeat_monitor(
	ch_new_data chan int,
	ch_req_ID chan int,
	ch_req_data chan networking.Elevator_node,
) {
	for {
		select {
		case id := <-ch_new_data: //new data arrives
			for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ { //runs a loop
				lighthouse := networking.Node_get_data(id, ch_req_ID, ch_req_data) //sets the "lighthouse" as a middle man as node get data only updates one elevator at a time
				//fmt.Println("Lighthouse: " + strconv.Itoa(id))
				elev_overview[i] = lighthouse //updates elev_overview with the new data
			}
		}
	}
}

//a sorting algorithm responsible for updating the placement struct from highest to lowest score
func sorting() {
	for p := 0; p < config.NUMBER_OF_ELEVATORS; p++ { //runs thrice
		var roundbest_index int                           //the strongest placement for this round
		var bestscore int                                 //the strongest placement for this round
		for i := p; i < config.NUMBER_OF_ELEVATORS; i++ { //ignores the stuff that has already been positioned
			if placement[i].score > bestscore { //if the score surpasses the others
				roundbest_index = i            //sets the new index
				bestscore = placement[i].score //sets the new best score
			}
		}
		placement[p].elevator_number = roundbest_index //sets the index of the highest scorer
	}
	//printing the sorting
	for x := 0; x < config.NUMBER_OF_ELEVATORS; x++ {
		fmt.Printf("Elevator%+v placed %+v with a score of %+v \n", placement[x].elevator_number, x, placement[x].score)
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
		case a := <-ch_drv_buttons: //takes the new data and runs a tournament to determine what the most suitable elevator is
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
			sorting()                                         //calls the sorting algorithm to sort the elevator placements
			for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ { //will automatically cycle the scoreboard and attempt to send from best to worst
				if elev_overview[placement[i].elevator_number].ID == config.ELEVATOR_ID { //if the winning ID is the elevators own
					fmt.Printf("own elevator won\n")
					button_calls := <-ch_drv_buttons //as the message needs to be passed between two channels we need a middle man
					ch_self_command <- button_calls
					break
				} else { //if the call is not going to itself
					if networking.Send_command(elev_overview[placement[i].elevator_number].ID, a.Floor, dir) { //send command to suitable external elevator
						fmt.Printf("external elevator won\n")
						break //if it succeds break the loop
					}
				}
			}
			//if a death or stall occurs
		case death_id := <-ch_take_calls: //id of the elevator in question is transmitted as an event
			for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ { //finds the elevator that has died in the internal overwiew struct
				if elev_overview[i].ID == death_id { //found the elevator
					for e := 0; e < 6; e++ { //checks all calls by running a
						var dir int //creates temp variables
						var floor int
						if elev_overview[i].HallCalls[e] == 1 { //if a hall call is found
							//here we need to run a distinction algorithm to determine direction to be able to run the tournament
							if e%2 == 0 { //the number is even so the dir is up
								dir = 1                     //sets the tempdir to up
								floor = e / 2               //finds the correct floor
								master_tournament(floor, 1) //runs a tournament with the new
								//this shit may cause errors as i am unshure if everyone is cool with floors starting at 0
							} else { //the number is odd so the dir is down
								dir = -1                     //sets the tempdir down
								floor = (e - 1) / 2          //finds the floor
								master_tournament(floor, -1) //sets the dir down
							}
							sorting() //runs the sorting algorithm
							//again tries to send the results to the elevators
							for c := 0; c < config.NUMBER_OF_ELEVATORS; c++ { //will automatically cycle the scoreboard and attempt to send from best to worst
								if elev_overview[placement[c].elevator_number].ID == config.ELEVATOR_ID { //if the winning ID is the elevators own
									button_calls := <-ch_drv_buttons //again the convertion is needed as it is between channels
									ch_self_command <- button_calls
									break
								} else {
									if networking.Send_command(elev_overview[placement[c].elevator_number].ID, floor, dir) {
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

//a function that scores all the elevators based on two inputs: floor and direction
func master_tournament(floor int, direction int) {
	//resets scoring to prepare the tournament
	for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ {
		placement[i].score = 0
		placement[i].elevator_number = 0
	}
	//filters out the nonworking and scores them
	for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ { //cycles shafts
		if !(elev_overview[i].Status == 404) { //if the elevator is nonfunctional it is ignored
			//direction scoring
			if direction == elev_overview[i].Direction { //if the elevators direction matches the input
				placement[i].score += 3 //give 3 good boy points
			}
			//placement scoring (with alot of conversion) basically takes the floor difference of where the elevator is and where it is supposed to go and then subtracts it with 4
			//this means that the closer the elevator is the higher the score
			placement[i].score += (4 - int(math.Abs(float64(elev_overview[i].Floor)-float64(floor))))
		}
	}
}
