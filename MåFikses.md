Problem at hvis vi har to calls i samme etasje kan den svare ene, sjekke om det finnes flere, finne en i samme etasjem ta den men får ikke kommet ut grunnet at
den ikke får til å få melding om at den har kommet frem.

Fiks slik at cab calls ikke alltid sjekker fra bunn men ser basert på rettning og nåværene plassering (Jobber med)

Reset direction når heisen ikke har flere ordre å service (Fikset)

Fiks slik at heisen ved første kjøring klarer å stoppe på eks 1 etasje når den er på vei til 0. Funker ikke per nå pga at start direction er 0 og logikken vår bygger
på at heisen har en retning fra sist når 

Legg til buffer size 1 on channels so if 2 commands arrive at the same time it does not deadlock

Forslag til endring på fredag, gjøre self-command og net-command til samme tingen siden de gjør 100% det samme (Penere ifølge anders)
