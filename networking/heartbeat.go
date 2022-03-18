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

func heartBeatTransmitter(ch_req_ID chan int, ch_req_data chan Elevator_node) (err error) {
	var msg, date, clock, broadcast string
	var ID int = config.ELEVATOR_ID
	var node Elevator_node
	//Resolve transmit connection (broadcast)
	broadcast = "255.255.255.255:" + strconv.Itoa(config.HEARTBEAT_PORT)
	network, _ := net.ResolveUDPAddr("udp", broadcast)
	con, _ := net.DialUDP("udp", nil, network)

	fmt.Println("Networking: starting heartbeat transmision")
	//Timer to define when to broadcast heartbeat data
	timer := time.NewTimer(config.HEARTBEAT_TIME)
	//Routine
	for {
		select {
		case <-timer.C:
			//Sampling date and time, and making it nice european style
			year, month, day := time.Now().Date()
			date = strconv.Itoa(day) + "/" + month.String() + "/" + strconv.Itoa(year)
			hour, minute, second := time.Now().Clock()
			clock = strconv.Itoa(hour) + ":" + strconv.Itoa(minute) + ":" + strconv.Itoa(second)
			msg = date + " " + clock + "_"
			//Requesting and getting elevator data
			node = Node_get_data(ID, ch_req_ID, ch_req_data)

			//Generating the heartbeat message
			msg = msg + strconv.Itoa(ID) + "_"
			msg = msg + strconv.Itoa(node.Direction) + "_"
			msg = msg + strconv.Itoa(node.Destination) + "_"
			msg = msg + strconv.Itoa(node.Floor) + "_"
			msg = msg + strconv.Itoa(node.Status) + "_"
			for i := range node.HallCalls {
				msg = msg + strconv.Itoa(node.HallCalls[i]) + "_"
			}
			//Sending the message
			//fmt.Println("Networking: sending HB message: " + msg)
			con.Write([]byte(msg))
			timer.Reset(config.HEARTBEAT_TIME)
		}
	}
}

func heartBeathandler(ch_req_ID, ch_ext_dead chan int, ch_req_data, ch_write_data chan Elevator_node) {
	//Initiate the UDP listener
	fmt.Println("Networking: HB starting listening thread")
	ch_heartbeatmsg := make(chan string)
	go heartbeat_UDPListener(ch_heartbeatmsg)

	//Initiate heartbeat timers and channels for each elevator except myself
	var ch_timerReset, ch_timerStop [config.NUMBER_OF_ELEVATORS]chan int
	ch_foundDead := make(chan int)
	fmt.Println("Networking: HB starting timers")
	for i := 1; i <= config.NUMBER_OF_ELEVATORS; i++ {
		if i != config.ELEVATOR_ID {
			ch_timerReset[i-1] = make(chan int)
			ch_timerStop[i-1] = make(chan int)
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
			for i := range node_data.HallCalls {
				node_data.HallCalls[i], _ = strconv.Atoi(data[6+i])
			}
			//fmt.Println("Networking: Got heartbeat msg from elevator " + strconv.Itoa(ID) + ": " + msg)
			fmt.Println("Elevator " + strconv.Itoa(ID) + " at floor: " + strconv.Itoa(node_data.Floor))
			//Write the node data
			ch_write_data <- node_data
			//Reset the appropriate timer
			ch_timerReset[ID-1] <- ID
			//Allert cost function that there is new data on this ID

		case msg_ID := <-ch_foundDead:
			//Timer has run out,
			fmt.Println("Networking: Elevator " + strconv.Itoa(msg_ID) + " is dead")
			node_data = Node_get_data(msg_ID, ch_req_ID, ch_req_data)
			node_data.Status = 404
			ch_write_data <- node_data
			go revive_calls(msg_ID)

		case msg_ID := <-ch_ext_dead: //Set status to 404 and stop the timer
			node_data = Node_get_data(msg_ID, ch_req_ID, ch_req_data)
			node_data.Status = 404
			ch_write_data <- node_data
			ch_timerStop[msg_ID-1] <- msg_ID
		}
	}
}

//Timer, waiting for something to timeout. Run as a go routine, accessed through channels
func heartbeatTimer(ID int, ch_foundDead, ch_timerReset, ch_timerStop chan int) {
	timer := time.NewTimer(config.HEARTBEAT_TIME_OUT)
	timer.Stop()
	for {
		select {
		case <-timer.C:
			ch_foundDead <- ID
			timer.Stop()
		case cmd_id := <-ch_timerReset:
			if cmd_id == ID {
				timer.Reset(config.HEARTBEAT_TIME_OUT)
			}
		case cmd_id := <-ch_timerStop:
			if cmd_id == ID {
				timer.Stop()
			}
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
		if ID <= config.NUMBER_OF_ELEVATORS && ID != config.ELEVATOR_ID {
			ch_heartbeatmsg <- msg
		}

	}
}

func printError(str string, err error) {
	if err != nil {
		fmt.Print(str)
		fmt.Println(err)
	}
}
