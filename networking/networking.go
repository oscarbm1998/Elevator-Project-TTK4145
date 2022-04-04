package networking

import (
	config "PROJECT-GROUP-10/config"
	"PROJECT-GROUP-10/elevio"
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"
)

type HallCall struct {
	Up   bool
	Down bool
}

type Elevator_node struct {
	Last_seen   string                            //Date and time of the last heartbeat message
	ID          int                               //Elevator ID
	Destination int                               //Current destination
	Direction   int                               //Current moving direction
	Floor       int                               //Current floor
	Status      int                               //Status variable. 404 = unreachable, 1 or 2 = stuck, 0 = OK
	HallCalls   [config.NUMBER_OF_FLOORS]HallCall //Current hallcalls tasks
}

//Array with information from all the elevators
var Elevator_nodes [config.NUMBER_OF_ELEVATORS]Elevator_node

func Main(
	ch_req_ID [3]chan int,
	ch_new_data, ch_take_calls chan int,
	ch_req_data, ch_write_data [3]chan Elevator_node,
	ch_net_command chan elevio.ButtonEvent,
	ch_hallCallsTot_updated chan [config.NUMBER_OF_FLOORS]HallCall) {
	ch_ext_dead := make(chan int)
	ch_hb_trans := make(chan bool)
	ch_hb_rec := make(chan bool)
	ch_cmd_rec := make(chan bool)
	ch_datahandler := make(chan bool)
	//Initiating the Elevator_nodes data with values
	for i := 1; i <= config.NUMBER_OF_ELEVATORS; i++ {
		if i != config.ELEVATOR_ID {
			Elevator_nodes[i-1].ID = i
			Elevator_nodes[i-1].Status = 2 //Status = 2: have not heard from it yet
		} else {
			Elevator_nodes[i-1].ID = i
		}
	}
	go node_data_handler(ch_req_ID, ch_req_data, ch_write_data, ch_datahandler)
	go heartBeathandler(ch_req_ID[0], ch_ext_dead, ch_new_data, ch_take_calls, ch_req_data[0], ch_write_data[0], ch_hallCallsTot_updated, ch_hb_rec)
	go heartBeatTransmitter(ch_req_ID[0], ch_req_data[0], ch_hallCallsTot_updated, ch_hb_trans)
	go command_listener(ch_net_command, ch_ext_dead, ch_cmd_rec)
	go deadLockDetector(ch_hb_trans, ch_hb_rec, ch_cmd_rec, ch_datahandler)
}

func node_data_handler(
	ch_req_ID [3]chan int,
	ch_req_data, ch_write_data [3]chan Elevator_node,
	ch_datahandler chan<- bool) {
	t := time.NewTimer(time.Second)
	t.Reset(time.Second)
	for {
		select {
		case <-t.C:
			t.Reset(time.Second)
			ch_datahandler <- true
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

//Linux version
func DialBroadcastUDP(port int) net.PacketConn {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		fmt.Println("Error: Socket:", err)
	}
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		fmt.Println("Error: SetSockOpt REUSEADDR:", err)
	}
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	if err != nil {
		fmt.Println("Error: SetSockOpt BROADCAST:", err)
	}
	syscall.Bind(s, &syscall.SockaddrInet4{Port: port})
	if err != nil {
		fmt.Println("Error: Bind:", err)
	}

	f := os.NewFile(uintptr(s), "")
	conn, err := net.FilePacketConn(f)
	if err != nil {
		fmt.Println("Error: FilePacketConn:", err)
	}
	f.Close()

	return conn
}

//Windows version
/*
func DialBroadcastUDP(port int) net.PacketConn {
	config := &net.ListenConfig{Control: func(network, address string, conn syscall.RawConn) error {
		return conn.Control(func(descriptor uintptr) {
			syscall.SetsockoptInt(syscall.Handle(descriptor), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			syscall.SetsockoptInt(syscall.Handle(descriptor), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
		})
	},
	}

	conn, err := config.ListenPacket(context.Background(), "udp4", fmt.Sprintf(":%d", port))
	fmt.Println(err)

	return conn
}*/

func deadLockDetector(ch_hb_trans, ch_hb_rec, ch_cmd_rec, ch_datahandler <-chan bool) {
	var timeOut time.Duration = time.Minute
	var timers [4]*time.Timer
	for i := 0; i < 4; i++ {
		timers[i] = time.NewTimer(timeOut)
		timers[i].Reset(timeOut)
	}
	timers[2].Stop()

	for {
		select {
		case <-ch_hb_trans:
			timers[0].Reset(timeOut)
		case <-timers[0].C:
			panic("Deadlock detected on heartbeat transmitter")
		case <-ch_hb_rec:
			timers[1].Reset(timeOut)
		case <-timers[1].C:
			panic("Deadlock detected on heartbeat receiver")
		case state := <-ch_cmd_rec:
			if state {
				timers[2].Reset(timeOut)
			} else {
				timers[2].Stop()
			}

		case <-timers[2].C:
			panic("Deadlock detected on command receiver")
		case <-ch_datahandler:
			timers[3].Reset(timeOut)
		case <-timers[3].C:
			panic("Deadlock detected on data handler")

		}
	}
}

func printError(str string, err error) {
	if err != nil {
		fmt.Print(str)
		fmt.Println(err)
	}
}
