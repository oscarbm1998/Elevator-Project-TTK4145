package networking

import (
	config "PROJECT-GROUP-10/config"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var HeartBeatLogger bool = false

func heartBeatTransmitter(ch_req_ID chan int, ch_req_data chan Elevator_node) (err error) {
	var msg, date, clock, broadcast string
	var ID int = config.ELEVATOR_ID
	var node Elevator_node
	//Resolve transmit connection (broadcast)
	broadcast = "255.255.255.255:" + strconv.Itoa(config.HEARTBEAT_PORT)
	network, _ := net.ResolveUDPAddr("udp", broadcast)
	con, _ := net.DialUDP("udp", nil, network)

	fmt.Println("Networking: starting heartbeat transmision")

	timer := time.NewTimer(config.HEARTBEAT_TIME) //Timer to define when to broadcast heartbeat data
	//Routine
	for {
		<-timer.C //Loop every second
		//Sampling date and time, and making it nice european style
		year, month, day := time.Now().Date()
		date = strconv.Itoa(day) + "/" + month.String() + "/" + strconv.Itoa(year)
		hour, minute, second := time.Now().Clock()
		clock = strconv.Itoa(hour) + ":" + strconv.Itoa(minute) + ":" + strconv.Itoa(second)
		msg = date + " " + clock + "_"

		node = Node_get_data(ID, ch_req_ID, ch_req_data) //Requesting and getting the latest elevator

		//Generating the heartbeat message
		msg = msg + strconv.Itoa(ID) + "_"
		msg = msg + strconv.Itoa(node.Direction) + "_"
		msg = msg + strconv.Itoa(node.Destination) + "_"
		msg = msg + strconv.Itoa(node.Floor) + "_"
		msg = msg + strconv.Itoa(node.Status)

		for i := range node.HallCalls {
			var up, down int = 0, 0
			if node.HallCalls[i].Up {
				up = 1
			}
			if node.HallCalls[i].Down {
				down = 1
			}
			msg = msg + "_" + strconv.Itoa(up) + "_"
			msg = msg + strconv.Itoa(down)
		}

		if HeartBeatLogger {
			fmt.Println("Networking: sending HB message " + msg)
		}

		con.Write([]byte(msg)) //Sending the message
		timer.Reset(config.HEARTBEAT_TIME)

	}
}

func heartBeathandler(
	ch_req_ID, ch_ext_dead, ch_new_data, ch_take_calls chan int,
	ch_req_data, ch_write_data chan Elevator_node,
	ch_hallCallsTot_updated chan [config.NUMBER_OF_FLOORS]HallCall) {

	//Initiate the UDP listener
	fmt.Println("Networking: HB starting listening thread")
	ch_heartbeatmsg := make(chan string)
	go heartbeat_UDPListener(ch_heartbeatmsg)

	//Initiate heartbeat timers and channels for each elevator except for myself
	var ch_timerReset, ch_timerStop [config.NUMBER_OF_ELEVATORS]chan bool
	ch_foundDead := make(chan int)
	fmt.Println("Networking: HB starting timers")
	for i := 1; i <= config.NUMBER_OF_ELEVATORS; i++ {
		if i != config.ELEVATOR_ID {
			ch_timerReset[i-1] = make(chan bool)
			ch_timerStop[i-1] = make(chan bool)
			go heartbeatTimer(i, ch_foundDead, ch_timerReset[i-1], ch_timerStop[i-1])
		}
	}

	var node_data Elevator_node
	for {
		select {
		case msg := <-ch_heartbeatmsg:
			//Parsing the received heartbeat message
			data := strings.Split(msg, "_")
			ID, _ := strconv.Atoi(data[1])
			node_data.Last_seen = data[0]
			node_data.ID = ID
			node_data.Direction, _ = strconv.Atoi(data[2])
			node_data.Destination, _ = strconv.Atoi(data[3])
			node_data.Floor, _ = strconv.Atoi(data[4])
			node_data.Status, _ = strconv.Atoi(data[5])
			var k int = 0
			for i := range node_data.HallCalls {
				up, _ := strconv.Atoi(data[6+k])
				k++
				down, _ := strconv.Atoi(data[6+k])
				k++
				switch up {
				case 1:
					node_data.HallCalls[i].Up = true
				case 0:
					node_data.HallCalls[i].Up = false
				}
				switch down {
				case 1:
					node_data.HallCalls[i].Down = true
				case 0:
					node_data.HallCalls[i].Down = false
				}
			}
			if HeartBeatLogger {
				fmt.Println("Networking: Got heartbeat msg from elevator " + strconv.Itoa(ID) + ": " + msg)
				fmt.Println("Elevator " + strconv.Itoa(ID) + " at floor: " + strconv.Itoa(node_data.Floor))
			}

			ch_write_data <- node_data  //Write the node data
			ch_timerReset[ID-1] <- true //Reset the appropriate timer
			//Update data
			ch_hallCallsTot_updated <- update_HallCallsTot(ch_req_ID, ch_req_data)
			ch_new_data <- ID //Tell cost function that there is new data on this ID
		case msg_ID := <-ch_foundDead:
			var msg, broadcast string

			ch_timerStop[msg_ID-1] <- true //Stop the timer of the dead elevator

			//Timer has run out, update status
			fmt.Println("Networking: Elevator " + strconv.Itoa(msg_ID) + " is dead")
			node_data = Node_get_data(msg_ID, ch_req_ID, ch_req_data)
			node_data.Status = 404 //Unreachable
			ch_write_data <- node_data

			//Tell everyone that an elevator has died and that you are taking responsibility
			fmt.Println("Networking: Reviving elevator " + strconv.Itoa(msg_ID) + ", taking his/her hall calls")
			msg = "98_" + strconv.Itoa(msg_ID) + "_DEAD_" + strconv.Itoa(config.ELEVATOR_ID)
			broadcast = "255.255.255.255:" + strconv.Itoa(config.COMMAND_PORT)
			network, _ := net.ResolveUDPAddr("udp", broadcast)
			con, _ := net.DialUDP("udp", nil, network)
			con.Write([]byte(msg))
			con.Close()

			ch_take_calls <- msg_ID //Tell the ordering package to take the hall calls of the dead elevator

		case msg_ID := <-ch_ext_dead: //Set status to 404 and stop the timer
			node_data = Node_get_data(msg_ID, ch_req_ID, ch_req_data)
			node_data.Status = 404
			ch_write_data <- node_data
			ch_timerStop[msg_ID-1] <- true
		}
	}
}

//Timer, waiting for something to timeout. Run as a go routine, accessed through channels
func heartbeatTimer(ID int, ch_foundDead chan int, ch_timerReset, ch_timerStop chan bool) {
	//Offset timeout based on elevator ID
	var time_TIMEOUT = config.HEARTBEAT_TIMEOUT + 100*time.Millisecond*time.Duration(config.ELEVATOR_ID)

	timer := time.NewTimer(time_TIMEOUT)
	timer.Stop()
	for {
		select {
		case <-timer.C:
			ch_foundDead <- ID
			timer.Stop()
		case <-ch_timerReset:
			timer.Reset(time_TIMEOUT)
		case <-ch_timerStop:
			timer.Stop()

		}
	}
}

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

func heartbeat_UDPListener(ch_heartbeatmsg chan string) {
	buf := make([]byte, 1024)
	var msg string
	var port string = ":" + strconv.Itoa(config.HEARTBEAT_PORT)
	fmt.Println("Networking: Listening for HB-messages on port " + port)
	//network, _ := net.ResolveUDPAddr("udp", port)
	//conn, _ := net.ListenUDP("udp", network)

	conn := DialBroadcastUDP(config.HEARTBEAT_PORT)

	for {
		//n, _, err := conn.ReadFromUDP(buf)
		n, _, _ := conn.ReadFrom(buf)
		msg = string(buf[0:n])
		data := strings.Split(msg, "_")
		ID, _ := strconv.Atoi(data[1])

		//Checking weather the message is of the correct format and sending to Heartbeat Handler
		if ID != config.ELEVATOR_ID && ID <= config.NUMBER_OF_ELEVATORS {
			ch_heartbeatmsg <- msg
		}
	}
}

//Returns an array of all the hallcalls currently being served
func update_HallCallsTot(ch_req_ID chan int, ch_req_data chan Elevator_node) (HallCallsTot [config.NUMBER_OF_FLOORS]HallCall) {
	var Elevator Elevator_node
	for i := 1; i <= config.NUMBER_OF_ELEVATORS; i++ {
		Elevator = Node_get_data(i, ch_req_ID, ch_req_data)
		for k := range Elevator.HallCalls {
			if Elevator.HallCalls[k].Up {
				HallCallsTot[k].Up = true
			}
			if Elevator.HallCalls[k].Down {
				HallCallsTot[k].Down = true
			}
		}
	}
	return HallCallsTot
}
