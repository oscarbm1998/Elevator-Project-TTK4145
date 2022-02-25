package singleElevator

var qeue_length int = 7
type elevator_Info struct {
	floor     int //where the elevator should go
	direction int //1 up -1 down
}

type evelvator_Qeue struct {
	floor     [qeue_length]int //the qeue of where the elevators should go
	direction [qeue_length]int //1 up -1 down
	age		  [qeue_length]int //the age of the call
}

func sort(curr_floor int) {
	var new_floor	int //new floor
	var new_direction	int //new direction
	var scoreboard	[qeue_length]int

	for i := 0; i <= qeue_length; i++{
		if
		//her skal algoritmen inn

	}


}

func shift_Qeue() {

}

func pass_Qeue() {

}
