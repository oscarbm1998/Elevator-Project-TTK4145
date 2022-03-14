package networking

import (
	config "PROJECT-GROUP-10/config"
	//"fmt"
	//"strconv"
	"time"
)

type Elevator_node struct {
	Last_seen   string
	ID          int
	Destination int
	Direction   int
	Floor       int
	Status      int
	IP          string
}

var Elevator_nodes [config.NUMBER_OF_ELEVATORS]Elevator_node

func Networking_main() {
	//Initialize heartbeat
	/*
		err := resolveHBConn()
		if err != nil {
			panic(err)
		}
	*/
	//go heartBeathandler()

	ch_req_ID := make(chan int)
	ch_req_data := make(chan Elevator_node)
	ch_write_data := make(chan Elevator_node)
	go Node_data_handler(ch_req_ID, ch_req_data, ch_write_data)
	go heartBeathandler(ch_write_data)
	//go heartBeatTransmitter(ch_req_ID, ch_req_data)

	/*
				//initiate command transmit connections
				for i := 0; i < config.NUMBER_OF_ELEVATORS-1; i++ {
					if i != config.ELEVATOR_ID+1 {
						network, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(config.COMMAND_SEND_PORT+i))
						printError("Command transmit network resolve error: ", err)
						conn, err := net.DialUDP("udp", nil, network)
						printError("Command transmit network dial error: ", err)
						command_cons[i] = conn
						//Readback cons
						network, err = net.ResolveUDPAddr("udp", ":"+strconv.Itoa(config.COMMAND_READBACK_PORT+i))
						printError("Command transmit network resolve error: ", err)
						conn, err = net.DialUDP("udp", nil, network)
						printError("Command transmit network dial error: ", err)
						readback_cons[i] = conn
					}
				}
				//Initiate command readback port connection
				adr, _ := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(config.COMMAND_READBACK_PORT+config.ELEVATOR_ID-1))
				readback_con, _ = net.ListenUDP("udp", adr)

				//Initiate command listener
				ch_command := make(chan string)

				go command_listener(ch_command)
				//Listen for commands
				for {
					select {
					case <-ch_comma
		}
					}
				}
	*/
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
			ID := data.ID
			Elevator_nodes[ID-1] = data
		}
	}
}

func Node_get_data(ID int, ch_req_ID chan int, ch_req_data chan Elevator_node) (nodeData Elevator_node) {
	ch_req_ID <- ID
	nodeData = <-ch_req_data
	for nodeData.ID != ID {
		ch_req_ID <- ID
		nodeData = <-ch_req_data
	}
	return nodeData
}

/*
func send_command(ID, floor, direction int) (success bool) {
	var attempts int = 1
	var cmd string
	if ID == config.ELEVATOR_ID {
		panic("Networking: I do not need networking to command myself")
	}

	cmd = strconv.Itoa(ID) + "_" + strconv.Itoa(floor) + "_" + strconv.Itoa(direction)
	//Send command
	_, err := command_cons[ID-1].Write([]byte(cmd))
	printError("Networking: Error sending command: ", err)
	if err == nil {
		fmt.Println("Networking: command sent to elevator " + strconv.Itoa(ID))
	}
	//Wait for readback
	buf := make([]byte, 1024)
	for {
		n, _, err := readback_con.ReadFromUDP(buf)
		msg := string(buf[0:n])

		//Sending again if the readback is wrong
		if msg != cmd {
			fmt.Println("Networking: bad readback, sending again")
			_, err = command_cons[ID-1].Write([]byte(cmd))
			printError("Networking: Error sending command: ", err)
			attempts++
		} else if msg == "CMD_REJECT" { //Command rejected
			fmt.Printf("Network: elevator rejected the command")
			success = false
			break
		} else {
			fmt.Println("Networking: readback OK")
			_, err = command_cons[ID-1].Write([]byte("CMD_OK"))
			success = true
			break
		}

		if attempts > 2 {
			fmt.Println("Networking: too many command readback attemps")
			success = false
			break
		}
	}

	if err != nil {
		success = false
	}

	return success
}
*/
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
