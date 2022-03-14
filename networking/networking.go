package networking

import (
	config "PROJECT-GROUP-10/config"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type Elevator_node struct {
	Last_seen   string
	ID          int
	Destination int
	Direction   int
	Floor       int
	Status      int
	Calls       [8]int
}

var Elevator_nodes [config.NUMBER_OF_ELEVATORS]Elevator_node

func Networking_main() {
	var ID int = config.ELEVATOR_ID

	Elevator_nodes[ID-1].ID = config.ELEVATOR_ID
	Elevator_nodes[ID-1].Floor = 4
	ch_req_ID := make(chan int)
	ch_req_data := make(chan Elevator_node)
	ch_write_data := make(chan Elevator_node)
	go Node_data_handler(ch_req_ID, ch_req_data, ch_write_data)
	go heartBeathandler(ch_req_ID, ch_req_data, ch_write_data)
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

func send_command(ID, floor, direction int) (success bool) {
	//(NB!!) This function does not check if the node is alive before attempting transmission.
	var attempts int = 1
	var cmd, rbc, broadcast string
	if ID == config.ELEVATOR_ID {
		panic("Networking: I do not need networking to command myself")
	}

	//Generate command
	//Format: ToElevatorID_ToFloor_InDirection_FromElevatorID
	cmd = strconv.Itoa(ID) + "_" + strconv.Itoa(floor) + "_" + strconv.Itoa(direction) + "_" + strconv.Itoa(config.ELEVATOR_ID)
	rbc = strconv.Itoa(config.ELEVATOR_ID) + "_" + strconv.Itoa(floor) + "_" + strconv.Itoa(direction) + "_" + strconv.Itoa(ID)

	//Initiate command broadcast connection
	broadcast = "255.255.255.255:" + strconv.Itoa(config.COMMAND_PORT)
	network, _ := net.ResolveUDPAddr("udp", broadcast)
	cmd_con, _ := net.DialUDP("udp", nil, network)

	//Initiate readback connection and timer
	ch_rbc_msg := make(chan string)
	ch_rbc_close := make(chan bool)
	go command_readback_listener(ch_rbc_msg, ch_rbc_close)
	//Send command
	_, err := cmd_con.Write([]byte(cmd))

	//Starting a timer for timeout
	timOut := time.Second * 3
	timer := time.NewTimer(timOut)
	for {
		select {
		case msg := <-ch_rbc_msg:
			data := strings.Split(msg, "_")
			rbc_id, _ := strconv.Atoi(data[0])
			//Sending again if the readback is wrong
			if rbc_id == config.ELEVATOR_ID {
				timer.Reset(timOut)
				if msg == rbc {
					fmt.Println("Networking: readback OK")
					_, err = cmd_con.Write([]byte(strconv.Itoa(ID) + "_CMD_OK"))
					success = true
					break
				} else if rbc == strconv.Itoa(config.ELEVATOR_ID)+"_CMD_REJECT" { //Command rejected
					fmt.Printf("Network: elevator rejected the command")
					success = false
					break
				} else {
					fmt.Println("Networking: bad readback, sending command again")
					_, err = cmd_con.Write([]byte(cmd))
					printError("Networking: Error sending command: ", err)
					attempts++
				}
				if attempts > 3 {
					fmt.Println("Networking: too many command readback attemps")
					success = false
					break
				}
			}
		case <-timer.C:
			fmt.Println("Networking: sending command timed out, no readback")
			timer.Stop()
			success = false
			ch_rbc_close <- true
			break
		}
	}

	if err != nil {
		success = false
	}

	return success
}

func command_readback_listener(ch_msg chan string, ch_close chan bool) {
	network, _ := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(config.COMMAND_RBC_PORT))
	con, _ := net.ListenUDP("udp", network)
	buf := make([]byte, 1024)
	for {
		select {
		case <-ch_close:
			con.Close()
			break
		default:
			con.SetReadDeadline(time.Now().Add(3 * time.Second))
			n, _, err := con.ReadFromUDP(buf)
			if err != nil {
				if e, ok := err.(net.Error); !ok || e.Timeout() {
					printError("Networking: command readback net error: ", err)
				} else {
					fmt.Println("Networking: Getting nothing on readback channel, so quitting")
					con.Close()
					break
				}
				break
			}
			msg := string(buf[0:n])
			ch_msg <- msg
		}
		defer con.Close()
	}
}

/*
func command_listener(ch_netcommand chan string) {
	adr, _ := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(config.COMMAND_REC_PORT+config.ELEVATOR_ID-1))
	con, _ := net.ListenUDP("udp", adr)
	buf := make([]byte, 1024)
	for {
		//Listen for incomming commands on command reception port
		n, _, err := con.ReadFromUDP(buf)
		printError("Networking: error from command listener: ", err)
		msg := string(buf[0:n])
		data := strings.Split(msg, "_")
		ID, _ := strconv.Atoi(data[0])
		floor, _ := strconv.Atoi(data[1])
		direction, _ := strconv.Atoi(data[2])

		if reject_command(floor, direction) { //Check if i can perfrom the task
			fmt.Println("Network: incomming command rejected")
			_, err = readback_cons[ID-1].Write([]byte("CMD_REJECT"))
		} else { //Accept the command by reading it back
			_, err = readback_cons[ID-1].Write([]byte(msg))
			//Wait for OK
			n, _, err = con.ReadFromUDP(buf)
			msg = string(buf[0:n])
			if msg == "CMD_OK" {
				ch_netcommand <- msg
			}
		}
	}
}
*/

func reject_command(direction, floor int) (reject bool) {
	if Elevator_nodes[config.ELEVATOR_ID-1].Status == 0 || floor < 0 || floor > config.NUMBER_OF_FLOORS {
		return true
	} else {
		return false
	}
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
