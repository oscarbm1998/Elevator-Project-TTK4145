package networking

import (
	config "PROJECT-GROUP-10/config"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var localIP string

func GetLocalIp() (string, error) {
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}

func heartBeatTransmitter(ch_req_ID chan int, ch_req_data chan Elevator_node) (err error) {
	timer := time.NewTimer(config.HEARTBEAT_TIME)
	var msg, date, clock, broadcast string
	var ID int = config.ELEVATOR_ID
	var node Elevator_node
	//Resolve transmit connection (broadcast)
	broadcast = "255.255.255.255:" + strconv.Itoa(config.HEARTBEAT_PORT)
	network, _ := net.ResolveUDPAddr("udp", broadcast)
	con, _ := net.DialUDP("udp", nil, network)

	fmt.Println("Heartbeat: starting transmit")

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
			//ch_req_ID <- ID
			node = Node_get_data(ID, ch_req_ID, ch_req_data)
			/*
				for node.ID != ID { //If i somehow get the wrong data
					ch_req_ID <- config.ELEVATOR_ID
					node = <-ch_req_data
				}
			*/
			//Compressing the elevator data to a message
			msg = msg + strconv.Itoa(ID) + "_"
			msg = msg + strconv.Itoa(node.Direction) + "_"
			msg = msg + strconv.Itoa(node.Destination) + "_"
			msg = msg + strconv.Itoa(node.Floor) + "_"
			msg = msg + strconv.Itoa(node.Status)
			fmt.Println("Sending: " + msg)
			//Sending to all nodes
			/*
				for i := 0; i < config.NUMBER_OF_ELEVATORS-1; i++ {
					if i != ID-1 {
						fmt.Println("sending")
						con.Write([]byte(msg))
					}

				}*/
			con.Write([]byte(msg))
			timer.Reset(config.HEARTBEAT_TIME)
		}
	}
}

func heartBeathandler(ch_write_data chan Elevator_node) {
	//Initiate the UDP listener
	fmt.Println("Networking: HB starting listening thread")
	ch_heartbeatmsg := make(chan string)
	go heartbeat_UDPListener(ch_heartbeatmsg)
	//Initiate heartbeat timers as go routines for each elevator
	ch_timerReset := make(chan int)
	var ch_foundDead, ch_timerStop chan int
	var ID int = config.ELEVATOR_ID
	fmt.Println("Networking: HB starting timers")
	for i := 1; i <= config.NUMBER_OF_ELEVATORS; i++ {
		if i != ID {
			go heartbeatTimer(i, ch_foundDead, ch_timerReset, ch_timerStop)
		}
	}

	var node_data Elevator_node
	for {
		select {
		case msg := <-ch_heartbeatmsg:
			//Parsing the received heartbeat message
			data := strings.Split(msg, "_")
			ID, _ = strconv.Atoi(data[1])
			node_data.Last_seen = data[0]
			node_data.ID = ID
			node_data.Direction, _ = strconv.Atoi(data[2])
			node_data.Destination, _ = strconv.Atoi(data[3])
			node_data.Floor, _ = strconv.Atoi(data[4])
			node_data.Status, _ = strconv.Atoi(data[5])
			fmt.Println("Got heartbeat msg from elevator " + strconv.Itoa(ID) + ": " + msg)

			ch_write_data <- node_data
			//Reset the appropriate timer
			ch_timerReset <- ID

		case msg_id := <-ch_foundDead:
			//Timer has run out,
			fmt.Printf("found " + strconv.Itoa(msg_id) + " dead")
			Elevator_nodes[<-ch_foundDead-1].Status = 404
		}
	}
}

//Timer, waiting for something to timeout. Run as a go routine, accessed through channels
func heartbeatTimer(ID int, ch_foundDead, ch_timerReset, ch_timer_stop chan int) {
	timer := time.NewTimer(config.HEARTBEAT_TIME_OUT)
	timer.Stop()
	for {
		select {
		case <-timer.C:
			fmt.Println("Elevator " + strconv.Itoa(ID) + " is dead")
			ch_foundDead <- ID
		case cmd_id := <-ch_timerReset:
			if cmd_id == ID {
				timer.Reset(config.HEARTBEAT_TIME_OUT)
			}
		case cmd_id := <-ch_timer_stop:
			if cmd_id == ID {
				timer.Stop()
			}
		}
	}
}

func heartbeat_UDPListener(ch_heartbeatmsg chan string) {
	buf := make([]byte, 1024)
	var msg string
	var port string = ":" + strconv.Itoa(config.HEARTBEAT_PORT)
	fmt.Println("Networking: Listening for HB-messages on port " + port)
	network, _ := net.ResolveUDPAddr("udp", port)
	conn, _ := net.ListenUDP("udp", network)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		msg = string(buf[0:n])
		printError("Error: ", err)
		data := strings.Split(msg, "_")
		ID, _ := strconv.Atoi(data[1])

		//Checking weather the message is of the correct format and sending to Heartbeat Handler
		if len(data) == 6 && ID <= config.NUMBER_OF_ELEVATORS && ID != config.ELEVATOR_ID {
			ch_heartbeatmsg <- msg
		}
		defer conn.Close()
	}
}

func printError(str string, err error) {
	if err != nil {
		fmt.Print(str)
		fmt.Println(err)
	}
}
