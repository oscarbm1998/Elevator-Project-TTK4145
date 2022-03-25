package networking

import (
	config "PROJECT-GROUP-10/config"
	"PROJECT-GROUP-10/elevio"
	"fmt"
	"strconv"
)

type HallCall struct {
	Up   bool
	Down bool
}

type Elevator_node struct {
	Last_seen   string
	ID          int
	Destination int
	Direction   int
	Floor       int
	Status      int
	HallCalls   [config.NUMBER_OF_FLOORS]HallCall
}

var Elevator_nodes [config.NUMBER_OF_ELEVATORS]Elevator_node

func Main(
	ch_req_ID [3]chan int,
	ch_new_data, ch_ext_dead, ch_take_calls chan int,
	ch_req_data, ch_write_data [3]chan Elevator_node,
	ch_net_command chan elevio.ButtonEvent,
	ch_hallCallsTot_updated chan [config.NUMBER_OF_FLOORS]HallCall) {

	Elevator_nodes[config.ELEVATOR_ID-1].ID = config.ELEVATOR_ID
	go node_data_handler(ch_req_ID, ch_req_data, ch_write_data)
	go heartBeathandler(ch_req_ID[0], ch_ext_dead, ch_new_data, ch_take_calls, ch_req_data[0], ch_write_data[0], ch_hallCallsTot_updated)
	go heartBeatTransmitter(ch_req_ID[0], ch_req_data[0])
	go command_listener(ch_net_command)
}

//Function responsible for node data. Works as a mutex for the resource
func node_data_handler(
	ch_req_ID [3]chan int,
	ch_req_data, ch_write_data [3]chan Elevator_node) {
	for {
		select {
		/*Handle data requests*/
		case ID := <-ch_req_ID[0]:
			ch_req_data[0] <- Elevator_nodes[ID-1]
		case ID := <-ch_req_ID[1]:
			ch_req_data[1] <- Elevator_nodes[ID-1]
		case ID := <-ch_req_ID[2]:
			ch_req_data[2] <- Elevator_nodes[ID-1]
		/*Write incomming data*/
		case data := <-ch_write_data[0]:
			if data.ID != 0 {
				Elevator_nodes[data.ID-1] = data
			}
		case data := <-ch_write_data[1]:
			if data.ID != 0 {
				Elevator_nodes[data.ID-1] = data
			}
		case data := <-ch_write_data[2]:
			if data.ID != 0 {
				Elevator_nodes[data.ID-1] = data
			}
		}
	}
}

func Node_get_data(ID int, ch_req_ID chan int, ch_req_data chan Elevator_node) (nodeData Elevator_node) {
	ch_req_ID <- ID
	nodeData = <-ch_req_data
	for nodeData.ID != ID {
		fmt.Println("Networking: SOMEONE TOOK MY DATA, I WANT " + strconv.Itoa(ID) + " BUT GOT " + strconv.Itoa(nodeData.ID))
		ch_req_ID <- ID
		nodeData = <-ch_req_data
	}
	return nodeData
}

func printError(str string, err error) {
	if err != nil {
		fmt.Print(str)
		fmt.Println(err)
	}
}
