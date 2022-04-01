package ordering

import (
	"PROJECT-GROUP-10/config"
	elevio "PROJECT-GROUP-10/elevio"
	networking "PROJECT-GROUP-10/networking"
	"fmt"
	"math"
	"sync"
)

type score_tracker struct { //this struct keeps track of the score and what elevator the socre belongs to
	score           int
	elevator_number int
} //uses the score tracker struct to make a sort of scoreboard the first array is the podium itself whilst the internal "elevator number" keeps track of which elevator is which

//a copy of the elevator node struct to keep internal track of the elevators
var elev_overview [config.NUMBER_OF_ELEVATORS]networking.Elevator_node

var number_of_alive_elevs int

//a translator for when we need to pass info between two channels

var m sync.Mutex

//checks and alerts the system whenever a heartbeat ping occurs
func heartbeat_monitor(
	ch_new_data chan int,
	ch_req_ID chan int,
	ch_req_data chan networking.Elevator_node,
) {
	for {
		id := <-ch_new_data                                                                                        //new data arrives
		elev_overview[id-1] = networking.Node_get_data(id, ch_req_ID, ch_req_data)                                 //updates elev_overview with the new data
		elev_overview[config.ELEVATOR_ID-1] = networking.Node_get_data(config.ELEVATOR_ID, ch_req_ID, ch_req_data) //update myself
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
	number_of_alive_elevs = config.NUMBER_OF_ELEVATORS
	for {
		select {
		case a := <-ch_drv_buttons: //takes the new data and runs a tournament to determine what the most suitable elevator is
			go cab_call_hander(ch_self_command, a, elev_overview)
			/*
				switch a.Button {
				case 0: //up
					master_tournament(a.Floor, int(elevio.MD_Up))
					dir := 1
					if number_of_alive_elevs >= 2 {
						Send_to_best_elevator(ch_self_command, a, dir)
					} else {
						ch_self_command <- a

					}
				case 1: //down
					master_tournament(a.Floor, elevio.MD_Down)
					dir := -1
					if number_of_alive_elevs >= 2 {
						Send_to_best_elevator(ch_self_command, a, dir)
					} else {
						ch_self_command <- a
					}
				case 2: //cab
					fmt.Print("Cab call found\n")
					ch_self_command <- a
				}
			*/
			//if a death or stall occurs
		case death_id := <-ch_take_calls: //id of the elevator in question is transmitted as an event
			number_of_alive_elevs--
			fmt.Printf("Number of alive elevators is now: %d", number_of_alive_elevs)
			go death_call_hander(death_id, ch_self_command, elev_overview)
			/*
				number_of_alive_elevs--
				fmt.Printf("Number of alive elevators is now: %d", number_of_alive_elevs)
				for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ { //finds the elevator that has died in the internal overwiew struct
					if elev_overview[i].ID == death_id { //found the elevator
						var temp_button_event elevio.ButtonEvent //defines a temporary button event in order to reuse a command
						for e := 0; e < config.NUMBER_OF_FLOORS; e++ {
							if elev_overview[i].HallCalls[e].Up {
								master_tournament(e, 1) //runs a tournament with the parametres for up
								temp_button_event.Button = 1
								temp_button_event.Floor = e
								Send_to_best_elevator(ch_self_command, temp_button_event, int(temp_button_event.Button))
							}
							//has to be this way otherwise it wont catch both instances
							if elev_overview[i].HallCalls[e].Down {
								master_tournament(e, -1) //runs a tournament with the parametres for up
								temp_button_event.Button = -1
								temp_button_event.Floor = e
								Send_to_best_elevator(ch_self_command, temp_button_event, int(temp_button_event.Button))
							}
						}
					}
				}
			*/
		}
	}
}

/*********************************
*		Welcome to Hell
			───▄▄▄
			─▄▀░▄░▀▄
			─█░█▄▀░█
			─█░▀▄▄▀█▄█▄▀
			▄▄█▄▄▄▄███▀

*********************************/
func cab_call_hander(ch_self_command chan elevio.ButtonEvent, a elevio.ButtonEvent, lighthouse [config.NUMBER_OF_ELEVATORS]networking.Elevator_node) {
	var placement [config.NUMBER_OF_ELEVATORS]score_tracker
	switch a.Button {
	case 0: //up
		placement := master_tournament(a.Floor, int(elevio.MD_Up), placement, lighthouse)
		dir := 1
		if number_of_alive_elevs >= 2 {
			Send_to_best_elevator(ch_self_command, a, dir, lighthouse, placement, &m)
		} else {
			ch_self_command <- a
		}
	case 1: //down
		placement := master_tournament(a.Floor, elevio.MD_Down, placement, lighthouse)
		dir := -1
		if number_of_alive_elevs >= 2 {
			Send_to_best_elevator(ch_self_command, a, dir, lighthouse, placement, &m)
		} else {
			ch_self_command <- a
		}
	case 2: //cab
		fmt.Print("Cab call found\n")
		ch_self_command <- a
	}
}

func death_call_hander(ID int, ch_self_command chan elevio.ButtonEvent, lighthouse [config.NUMBER_OF_ELEVATORS]networking.Elevator_node) {
	m.Lock()
	var placement [config.NUMBER_OF_ELEVATORS]score_tracker
	for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ { //finds the elevator that has died in the internal overwiew struct
		if lighthouse[i].ID == ID { //found the elevator
			var temp_button_event elevio.ButtonEvent //defines a temporary button event in order to reuse a command
			for e := 0; e < config.NUMBER_OF_FLOORS; e++ {
				if lighthouse[i].HallCalls[e].Up {
					placement = master_tournament(e, 1, placement, lighthouse) //runs a tournament with the parametres for up
					temp_button_event.Button = 1
					temp_button_event.Floor = e
					Send_to_best_elevator(ch_self_command, temp_button_event, 1, lighthouse, placement, &m)
				}
				//has to be this way otherwise it wont catch both instances
				if lighthouse[i].HallCalls[e].Down {
					placement = master_tournament(e, -1, placement, lighthouse) //runs a tournament with the parametres for up
					temp_button_event.Button = -1
					temp_button_event.Floor = e
					Send_to_best_elevator(ch_self_command, temp_button_event, -1, lighthouse, placement, &m)
				}
			}
		}
	}
	m.Unlock()
}

/*
			Monke mode
              ██████
            ██▒▒▒▒▒▒██
          ██▓▓▓▓▓▓▒▒██
          ██▓▓▒▒▒▒▒▒▒▒██
        ██░░      ▒▒    ██
        ██░░  ████░░██  ░░██
      ████░░  ████░░██  ░░██
      ██▓▓▒▒▒▒░░░░▒▒░░▓▓▒▒██
      ██▓▓▒▒▒▒▒▒░░░░░░▓▓▒▒██
    ██▓▓████████░░░░░░████
  ██▓▓▓▓██▓▓▓▓████████▓▓██
██▓▓▓▓██▓▓▒▒▒▒██▓▓▓▓▓▓██▓▓██
██▓▓██▓▓▒▒▒▒▒▒██▒▒▒▒░░▓▓▒▒▓▓██
██▓▓██▓▓▒▒▒▒▒▒██▒▒░░██▓▓▒▒▓▓██
██▓▓██▓▓▒▒▒▒▒▒▒▒████▓▓▓▓▒▒▓▓██
██▓▓██▓▓▒▒▒▒▒▒▒▒▒▒██░░▒▒▒▒▓▓██
██░░██░░▒▒▒▒░░░░░░████░░░░░░██
  ██████░░░░░░██░░██  ██████
        ██████████
*/

/*
func send_command_helper(returnval chan bool, ID int, floor int, direction int, m *sync.Mutex) {
	m.Lock()
	if networking.Send_command(ID, floor, direction) {
		returnval <- true
	} else {
		returnval <- false
	}
	m.Unlock()
	return
}
*/

//a function that scores all the elevators based on two inputs: floor and direction
func master_tournament(floor int, direction int, placement [config.NUMBER_OF_ELEVATORS]score_tracker, lighthouse [config.NUMBER_OF_ELEVATORS]networking.Elevator_node) (return_placement [config.NUMBER_OF_ELEVATORS]score_tracker) {
	//resets scoring to prepare the tournament
	for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ {
		fmt.Printf("floor of elev %d is %d \n", i, lighthouse[i].Floor)
		fmt.Printf("direction of elev %d is %d \n", i, lighthouse[i].Direction)
	}

	for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ {
		placement[i].score = 0
		placement[i].elevator_number = 0
	}
	//filters out the nonworking and scores them
	for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ { //cycles shafts
		if !(lighthouse[i].Status != 0) { //if the elevator is nonfunctional it is ignored
			//direction scoring
			if direction == lighthouse[i].Direction { //if the elevators direction matches the input
				placement[i].score += 2 //give 3 good boy points
			}
			//placement scoring (with alot of conversion) basically takes the floor difference of where the elevator is and where it is supposed to go and then subtracts it with 4
			//this means that the closer the elevator is the higher the score
			placement[i].score += (3 - int(math.Abs(float64(lighthouse[i].Floor-floor))))
		}
	}
	return placement
}

func Send_to_best_elevator(ch_self_command chan elevio.ButtonEvent, a elevio.ButtonEvent, dir int, lighthouse [config.NUMBER_OF_ELEVATORS]networking.Elevator_node, placement [config.NUMBER_OF_ELEVATORS]score_tracker, m *sync.Mutex) {

	var temporary_placement [config.NUMBER_OF_ELEVATORS]score_tracker = sorting(placement) //calls the sorting algorithm to sort the elevator placements
	for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ {                                      //will automatically cycle the scoreboard and attempt to send from best to worst
		if lighthouse[temporary_placement[i].elevator_number].ID == config.ELEVATOR_ID { //if the winning ID is the elevators own
			fmt.Printf("own elevator won\n")
			button_calls := a //as the message needs to be passed between two channels we need a middle man
			ch_self_command <- button_calls
			break
		} else if lighthouse[temporary_placement[i].elevator_number].Status == 0 { //if the call is not going to itself
			m.Lock()
			success := networking.Send_command(lighthouse[temporary_placement[i].elevator_number].ID, a.Floor, dir)
			m.Unlock()
			if success {
				break
			}

		}
	}
}

//a sorting algorithm responsible for updating the placement struct from highest to lowest score
func sorting(placement [config.NUMBER_OF_ELEVATORS]score_tracker) (return_placement [config.NUMBER_OF_ELEVATORS]score_tracker) {
	var temp_placement [config.NUMBER_OF_ELEVATORS]score_tracker
	for p := 0; p < config.NUMBER_OF_ELEVATORS; p++ { //runs thrice
		var roundbest_index int = 0                       //the strongest placement for this round
		var bestscore int = 0                             //the strongest placement for this round
		for i := p; i < config.NUMBER_OF_ELEVATORS; i++ { //ignores the stuff that has already been positioned
			if placement[i].score > bestscore { //if the score surpasses the others
				roundbest_index = i            //sets the new index
				bestscore = placement[i].score //sets the new best score
			}
		}
		temp_placement[p].elevator_number = roundbest_index //sets the index of the highest scorer
	}
	//printing the sorting
	for x := 0; x < config.NUMBER_OF_ELEVATORS; x++ {
		fmt.Printf("Elevator%+v placed %+v with a score of %+v \n", temp_placement[x].elevator_number, x, temp_placement[x].score)
	}
	return temp_placement
}

/*
for e := 0; e < 6; e++ { //checks all calls by running a
	var dir int //creates temp variables
	var floor int
	if elev_overview[i].HallCalls[e].Up {
		master_tournament(e, 1) //runs a tournament with the parametres for up
		dir = 1
	} else if elev_overview[i].HallCalls[e].Down {
		master_tournament(e, -1) //runs a tournament with the parametres for up
		dir = -1
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
			} else { //if the call is not going to itself
				returnval := make(chan bool)
				go send_command_helper(returnval ,elev_overview[placement[c].elevator_number].ID, floor, dir)
				if returnval {
					fmt.Printf("external elevator won\n")
					close(returnval)
					break
				} else {
					close(returnval)
				}
			}
		}
	}
}
*/
