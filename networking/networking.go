package networking

import (
	config "PROJECT-GROUP-10/config"
	"PROJECT-GROUP-10/elevio"
	"fmt"
	"net"
	"strconv"
	"time"
)

type Elevator_node struct {
	Last_seen   string
	ID          int
	Destination int
	Direction   int
	Floor       int
	Status      int
	HallCalls   [6]int
}

var Elevator_nodes [config.NUMBER_OF_ELEVATORS]Elevator_node

func Main(
	ch_req_ID [3]chan int,
	ch_new_data, ch_ext_dead, ch_take_calls chan int,
	ch_req_data, ch_write_data [3]chan Elevator_node,
	ch_net_command chan elevio.ButtonEvent) {
	Elevator_nodes[config.ELEVATOR_ID-1].ID = config.ELEVATOR_ID
	go node_data_handler(ch_req_ID, ch_req_data, ch_write_data)
	go heartBeathandler(ch_req_ID[0], ch_ext_dead, ch_new_data, ch_take_calls, ch_req_data[0], ch_write_data[0])
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

/*
func network_main_observer(ch_main_observer chan string) {
	//Initiate threads that listenes to messages on all ports
	var ch_observers chan string
	var ch_stuck, ch_reset, ch_stop chan int
	for i := 1; i < config.NUMBER_OF_ELEVATORS; i++ {
		go stuck_timer(i, ch_stuck, ch_reset, ch_stop)
	}
	ch_stop <- config.ELEVATOR_ID
	for {
		select {
			case <-ch_stuck:
				ch_main_observer <- strconv.Itoa(i)+ "_" +strconv.Itoa()
			}
		}
	}
}
*/

func stuck_timer(ID int, ch_stuck, ch_reset, ch_stop chan int) {
	timer := time.NewTimer(config.ELEVATOR_STUCK_TIMOUT)
	for {
		select {
		case <-timer.C:
			ch_stuck <- ID
		case <-ch_reset:
			if <-ch_reset == ID {
				timer.Reset(config.ELEVATOR_STUCK_TIMOUT)
			}
		case <-ch_stop:
			if <-ch_stop == ID {
				timer.Stop()
			}
		}

	}
}

func revive_calls(ID int, ch_take_calls chan int) {
	var msg, broadcast string
	fmt.Println("Networking: Reviving elevator " + strconv.Itoa(ID) + ", taking his/her hall calls")
	msg = "98_" + strconv.Itoa(ID) + "_DEAD_" + strconv.Itoa(config.ELEVATOR_ID)
	broadcast = "255.255.255.255:" + strconv.Itoa(config.COMMAND_PORT)
	network, _ := net.ResolveUDPAddr("udp", broadcast)
	con, _ := net.DialUDP("udp", nil, network)
	con.Write([]byte(msg))
	ch_take_calls <- ID
}
