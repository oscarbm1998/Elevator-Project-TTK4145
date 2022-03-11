package networking

import (
	config "PROJECT-GROUP-10/config"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var HB_con_Out [config.NUMBER_OF_ELEVATORS - 1]*net.UDPConn
var elevatorsIPs [config.NUMBER_OF_ELEVATORS - 1]string

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

func heartBeatTransmitter() (err error) {
	timer := time.NewTimer(config.HEARTBEAT_TIME)
	var msg, date, clock string
	var ID int = config.ELEVATOR_ID
	fmt.Println("Heartbeat: starting transmit")
	for {
		select {
		case <-timer.C:
			//Sampling date and time, and making it nice european style
			year, month, day := time.Now().Date()
			date = strconv.Itoa(day) + "/" + month.String() + "/" + strconv.Itoa(year)
			hour, minute, second := time.Now().Clock()
			clock = strconv.Itoa(hour) + ":" + strconv.Itoa(minute) + ":" + strconv.Itoa(second)
			msg = date + " " + clock + "_"

			//Adding elevator data
			msg = msg + strconv.Itoa(ID) + "_"
			msg = msg + strconv.Itoa(Elevator_nodes[ID-1].Direction) + "_"
			msg = msg + strconv.Itoa(Elevator_nodes[ID-1].Destination) + "_"
			msg = msg + strconv.Itoa(Elevator_nodes[ID-1].Floor) + "_"
			msg = msg + strconv.Itoa(Elevator_nodes[ID-1].Status)
			fmt.Println("Sending: " + msg)
			//Sending to all nodes
			//network, _ := net.ResolveUDPAddr("udp", "10.100.23.179:6969")
			//con, _ := net.DialUDP("udp", nil, network)
			for i := 0; i < config.NUMBER_OF_ELEVATORS-1; i++ {
				if i != ID-1 {
					fmt.Println("sending")
					HB_con_Out[i].Write([]byte(msg))
				}

			}
			fmt.Println("restarting timer")
			timer.Reset(config.HEARTBEAT_TIME)
		}

	}
}

func resolveHBConn() (err error) {
	for i := 0; i < config.NUMBER_OF_ELEVATORS-1; i++ {
		if i != config.ELEVATOR_ID-1 {
			//Outgoing
			network, err := net.ResolveUDPAddr("udp", Elevator_nodes[i].IP+":"+strconv.Itoa(config.HEARTBEAT_TRANS_PORT))
			printError("resolveHBconn setup error: ", err)
			con, err := net.DialUDP("udp", nil, network)
			HB_con_Out[i] = con
			printError("resolveHBconn dial error: ", err)
		}
	}
	return err
}

func heartBeathandler() {
	//Initiate connections

	//Initiate the UDP heartbeat listener
	ch_heartbeatmsg := make(chan string)
	fmt.Println("Heartbead UDP listener thread starting")
	go heartbeat_UDPListener(ch_heartbeatmsg)
	fmt.Println("Heartbead UDP listener thread started")
	//Initiate heartbeat timers as go routines for each elevator
	var ch_foundDead, ch_timerStop, ch_timerReset chan int
	var ID int
	fmt.Println("Starting HB listener timers")
	for i := 0; i < config.NUMBER_OF_ELEVATORS-1; i++ {
		if i != ID-1 {
			go heartbeatTimer(i+1, ch_foundDead, ch_timerReset, ch_timerStop)
		}
	}
	fmt.Println("HB listener timers started")

	for {
		fmt.Println("Started for routine")
		select {
		case <-ch_heartbeatmsg:
			//Parsing the received heartbeat message
			fmt.Println("Message")
			data := strings.Split(<-ch_heartbeatmsg, "_")
			fmt.Println("message len: " + strconv.Itoa(len(data)))
			if len(data) == 6 {
				ID, _ = strconv.Atoi(data[1])
				Elevator_nodes[ID-1].Last_seen = data[0]
				Elevator_nodes[ID-1].ID = ID
				Elevator_nodes[ID-1].Direction, _ = strconv.Atoi(data[2])
				Elevator_nodes[ID-1].Destination, _ = strconv.Atoi(data[3])
				Elevator_nodes[ID-1].Floor, _ = strconv.Atoi(data[4])
				Elevator_nodes[ID-1].Status, _ = strconv.Atoi(data[5])

				//Reset the appropriate timer
				ch_timerReset <- ID
				fmt.Println("Got heartbeat msg from elevator " + strconv.Itoa(ID))
			}
		case <-ch_foundDead:
			//Timer has run out,
			fmt.Printf("found dead")
			Elevator_nodes[<-ch_foundDead-1].Status = 404
		}
	}
}

//Timer, waiting for something to timeout. Run as a go routine, accessed through channels
func heartbeatTimer(ID int, ch_foundDead, ch_timerReset, ch_timer_stop chan int) {
	timer := time.NewTimer(config.HEARTBEAT_TIME)
	for {
		select {
		case <-timer.C:
			ch_foundDead <- ID
		case <-ch_timerReset:
			if <-ch_timerReset == ID {
				timer.Reset(config.HEARTBEAT_TIME)
			}
		case <-ch_timer_stop:
			if <-ch_timer_stop == ID {
				timer.Stop()
			}
		}
	}
}

func heartbeat_UDPListener(ch_heartbeatmsg chan<- string) {
	fmt.Println("Hello from UDP listener")
	buf := make([]byte, 1024)
	var msg string
	var port string = ":" + strconv.Itoa(config.HEARTBEAT_REC_PORT)
	fmt.Println(port)
	network, _ := net.ResolveUDPAddr("udp", port)
	conn, _ := net.ListenUDP("udp", network)
	for {
		fmt.Println("Waiting for message")
		n, adr, _ := conn.ReadFromUDP(buf)
		fmt.Println(adr)
		msg = string(buf[0:n])
		fmt.Println("Got message:" + msg)
		ch_heartbeatmsg <- msg
	}
}

func printError(str string, err error) {
	if err != nil {
		fmt.Print(str)
		fmt.Println(err)
	}
}
