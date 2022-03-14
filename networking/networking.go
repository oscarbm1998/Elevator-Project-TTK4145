package networking

import (
	config "PROJECT-GROUP-10/config"
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
	HallCalls   [8]int
}

var Elevator_nodes [config.NUMBER_OF_ELEVATORS]Elevator_node

func Networking_main() {
	var ID int = config.ELEVATOR_ID

	Elevator_nodes[ID-1].ID = config.ELEVATOR_ID
	Elevator_nodes[ID-1].Floor = 4
	ch_req_ID := make(chan int)
	ch_req_data := make(chan Elevator_node)
	ch_write_data := make(chan Elevator_node)
	ch_ext_dead := make(chan int)
	go Node_data_handler(ch_req_ID, ch_req_data, ch_write_data)
	go heartBeathandler(ch_req_ID, ch_ext_dead, ch_req_data, ch_write_data)
	go heartBeatTransmitter(ch_req_ID, ch_req_data)

}

//Function responsible for node data
func Node_data_handler(
	ch_req_ID chan int,
	ch_req_data, ch_write_data chan Elevator_node) {
	for {
		select {
		case ID := <-ch_req_ID: //Sending node data
			ch_req_data <- Elevator_nodes[ID-1]
		case data := <-ch_write_data: //Writing node data
			Elevator_nodes[data.ID-1] = data
		}
	}
}

func Node_get_data(ID int, ch_req_ID chan int, ch_req_data chan Elevator_node) (nodeData Elevator_node) {
	ch_req_ID <- ID
	nodeData = <-ch_req_data
	for nodeData.ID != ID {
		fmt.Println("Loop " + strconv.Itoa(ID) + " " + strconv.Itoa(nodeData.ID))
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

func revive_calls(ID int) {
	var msg, broadcast string
	fmt.Println("Networking: Reviving elevator " + strconv.Itoa(ID) + ", taking his/her hall calls")
	msg = "98_" + strconv.Itoa(ID) + "_DEAD_" + strconv.Itoa(config.ELEVATOR_ID)
	broadcast = "255.255.255.255:" + strconv.Itoa(config.COMMAND_PORT)
	network, _ := net.ResolveUDPAddr("udp", broadcast)
	con, _ := net.DialUDP("udp", nil, network)
	con.Write([]byte(msg))
}
