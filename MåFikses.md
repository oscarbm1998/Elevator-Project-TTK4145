Heisen kan stoppe på en etasje som blir lagt til etter den er i bevegelse så lenge det stemmer over rens med nåværende retning
Eks heisen kan stå i 4. få bedskjed om 2.ned og stoppe på 3.ned hvis den blir lagt til i mellomtiden
Den kan likevel ikke begynne i 1. etasje, 2, 3 og 4 etajse og så gå rett til 4 etasje som ville vært logisk

Problem at hvis vi har to calls i samme etasje kan den svare ene, sjekke om det finnes flere, finne en i samme etasjem ta den men får ikke kommet ut grunnet at
den ikke får til å få melding om at den har kommet frem.

type Elevator_node struct {
	Last_seen   string
	ID          int
	Destination int
	Direction   int
	Floor       int
	Status      int
	HallCalls   [6]int Endre denne til å være bool og lik single elevator varianten
}