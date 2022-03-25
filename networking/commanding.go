package networking

import (
	config "PROJECT-GROUP-10/config"
	elevio "PROJECT-GROUP-10/elevio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func Send_command(ID, floor, direction int) (success bool) {
	//(NB!!) This function does not check if the node is alive before attempting transmission.
	var attempts int = 1
	var cmd, rbc, broadcast string
	if ID == config.ELEVATOR_ID {
		panic("Networking: I do not need networking to command myself")
	}

	//Generate command
	//Format: ToElevatorID_ToFloor_InDirection_FromElevatorID
	cmd = strconv.Itoa(ID) + "_" + strconv.Itoa(floor) + "_" + strconv.Itoa(direction) + "_" + strconv.Itoa(config.ELEVATOR_ID)

	//Expected readback
	//Format: fromElevatorID_ToFloor_InDirection_toElevatorID
	rbc = strconv.Itoa(config.ELEVATOR_ID) + "_" + strconv.Itoa(floor) + "_" + strconv.Itoa(direction) + "_" + strconv.Itoa(ID)

	//Initiate command broadcast connection
	broadcast = "255.255.255.255:" + strconv.Itoa(config.COMMAND_PORT)
	network, _ := net.ResolveUDPAddr("udp", broadcast)
	cmd_con, _ := net.DialUDP("udp", nil, network)

	//Initiate readback connection and timer
	ch_rbc_msg := make(chan string)
	ch_rbc_close := make(chan bool)
	ch_rbc_listen := make(chan bool)
	go command_readback_listener(ch_rbc_msg, ch_rbc_close, ch_rbc_listen)

	//Send command
	fmt.Println("Network: sending command " + cmd)
	_, err := cmd_con.Write([]byte(cmd))
	ch_rbc_listen <- true
	printError("Networking: Error sending command: ", err)

	//Starting a timer for timeout
	timOut := time.Second * 3
	timer := time.NewTimer(timOut)
	for {
		select {
		case msg := <-ch_rbc_msg:
			data := strings.Split(msg, "_")
			fmt.Println("Readback: " + msg + " expected: " + rbc)
			rbc_id, _ := strconv.Atoi(data[0])
			//Sending again if the readback is wrong
			if rbc_id == config.ELEVATOR_ID {
				timer.Reset(timOut)
				if msg == rbc {
					fmt.Println("Networking: readback OK")
					cmd_con.Write([]byte(strconv.Itoa(ID) + "_CMD_OK"))
					ch_rbc_close <- true
					success = true
					goto Exit
				} else if rbc == strconv.Itoa(config.ELEVATOR_ID)+"_CMD_REJECT" { //Command rejected
					fmt.Printf("Network: elevator rejected the command")
					ch_rbc_close <- true
					success = false
					goto Exit
				} else {
					fmt.Println("Networking: bad readback, sending command again")
					_, err = cmd_con.Write([]byte(cmd))
					ch_rbc_listen <- true
					printError("Networking: Error sending command: ", err)
					attempts++
				}
				if attempts > 3 {
					fmt.Println("Networking: too many command readback attemps")
					success = false
					goto Exit
				}
			}
		case <-timer.C:
			fmt.Println("Networking: sending command timed out, no readback")
			timer.Stop()
			success = false
			goto Exit
		}
	}
Exit:
	//Stopping readback listener and returning the results
	ch_rbc_close <- true
	return success
}

func command_readback_listener(ch_msg chan<- string, ch_close, ch_rbc_listen <-chan bool) {
	buf := make([]byte, 1024)
	for {
		select {
		case <-ch_close:
			goto Exit
		case <-ch_rbc_listen:
			con := DialBroadcastUDP(config.COMMAND_RBC_PORT)
			con.SetReadDeadline(time.Now().Add(3 * time.Second)) //Will only wait for a response for 3 seconds
			n, _, err := con.ReadFrom(buf)
			if err != nil {
				if e, ok := err.(net.Error); !ok || e.Timeout() {
					printError("Networking: command readback net error: ", err)
				} else {
					fmt.Println("Networking: Getting nothing on readback channel, so quitting")
				}
				goto Exit
			}
			msg := string(buf[0:n])
			ch_msg <- msg
			con.Close()
		}
	}
Exit:
	fmt.Println("Networking: closing readback listener")
}

func command_listener(ch_netcommand chan elevio.ButtonEvent) {
	var button_command elevio.ButtonEvent
	var rbc string
	buf := make([]byte, 1024)
	//adr, _ := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(config.COMMAND_PORT))
	//cmd_con, _ := net.ListenUDP("udp", adr) //Listening to the command port
	cmd_con := DialBroadcastUDP(config.COMMAND_PORT)
	adr, _ := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.Itoa(config.COMMAND_RBC_PORT))
	rbc_con, _ := net.DialUDP("udp", nil, adr) //Broadcasting on the readback port

	fmt.Println("Networking: command listener listenening on port :" + strconv.Itoa(config.COMMAND_PORT))
	for {
		//Listen for incomming commands on command reception port
		n, _, err := cmd_con.ReadFrom(buf)
		printError("Networking: error from command listener: ", err)
		msg := string(buf[0:n])
		data := strings.Split(msg, "_")
		ID, _ := strconv.Atoi(data[0])

		if ID == config.ELEVATOR_ID { //Command is addressed to me
			floor, _ := strconv.Atoi(data[1])
			direction, _ := strconv.Atoi(data[2])
			from_ID, _ := strconv.Atoi(data[3])
			rbc = data[3] + "_" + data[1] + "_" + data[2] + "_" + strconv.Itoa(ID)

			if reject_command(floor, direction) { //Check if i can perfrom the task
				fmt.Println("Networking: incomming command from elevator " + strconv.Itoa(from_ID) + " rejected")
				rbc_con.Write([]byte(strconv.Itoa(from_ID) + "_CMD_REJECT"))
			} else {

				//Accept the command by reading it back
				rbc_con.Write([]byte(rbc))
				//Wait for OK
				n, _, _ = cmd_con.ReadFrom(buf)
				msg = string(buf[0:n])
				if msg == strconv.Itoa(config.ELEVATOR_ID)+"_CMD_OK" {
					//Pass the command to the elevator
					switch direction {
					case -1:
						button_command.Button = elevio.BT_HallDown
					case 1:
						button_command.Button = elevio.BT_HallUp
					}
					button_command.Floor = floor
					ch_netcommand <- button_command
					fmt.Println("Networking: got a command from elevator " + strconv.Itoa(from_ID))
				}
			}
		} else if ID == 98 { //Announcement to everyone from someone
			code := data[2]
			if code == "DEAD" {
				dead_ID, _ := strconv.Atoi(data[1])
				reportedBy_ID, _ := strconv.Atoi(data[3])
				if reportedBy_ID != config.ELEVATOR_ID {
					fmt.Println("Networking: elevator " + strconv.Itoa(dead_ID) + " was found dead by elevator " + strconv.Itoa(reportedBy_ID))
				}
			}
		}
	}

}

func reject_command(floor, direction int) (reject bool) {
	if Elevator_nodes[config.ELEVATOR_ID-1].Status != 0 {
		fmt.Println("Reason for reject: my status is not 0")
		return true
	} else if floor < 0 || floor > config.NUMBER_OF_FLOORS {
		fmt.Println("Reason for reject: illigal floor, can't go to floor " + strconv.Itoa(floor))
		return true
	} else if floor == Elevator_nodes[config.ELEVATOR_ID-1].Floor {
		fmt.Println("Reason for reject: I'm allready there")
		return true
	} else {
		return false
	}
}
