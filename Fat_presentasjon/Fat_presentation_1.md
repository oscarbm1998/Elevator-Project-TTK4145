Sanntidsprogrammering
=====================

### Hva ønsker vi å implementere:
Vi ønsker å implementere en peer-to-peer-nettverksløsning for heissystemet hvor den som får sitt knappepanel trykket på vil fungere som master og enten ta heisen selv, eller delegere videre hvis den ikke har mulighet. 

### Hvordan skal boot-up fungere?
Boot up vil fungere ved at noden som blir booted først vil fungere som master i boot-up prosessen og boote resten. 

### Hvordan skal vi sjekke om heiser faller ut av nettet?
Vi ønsker en UDP-nettverkstopologi som benytter en UDP-hjerterytme for å sjekke om samtlige heiser er koblet til nettet. Dette gjøres med en UDP-broadcast funksjon som kjører en timer hver gang den mottar en Heart-beat og hvis den ikke har mottatt info fra en spesifikk heis i så, så lang tid kan heisen anta at den har falt ut av nettet. Her er det viktig å tenke på at man ikke overloader nettet samtidig som man sender ofte nok til at packet loss ikke nødvendigvis blir et problem. 

### Hva skal heisen gjøre hvis den faller ut?
Vi ønsker at heisen skal kjøre en slags single elevator mode ettersom hvis en heis faller ut vil ikke knappepanelet til den heisen klare å sende info til resten om at det eksisterer en ordre der. Det er derfor greit at den heisen kan respondere på Calls som den får trykket inn. Denne single elevator mode kan for eksempel aktiveres når heisen har registrert at den har falt ut. 

### Hva skal heisen gjøre når den kommer tilbake?
Heisen vil selv registrere at den har kommet tilbake ettersom den vil motta "hjerterytmen" til de andre heisene. Ettersom at hjerterytmen vil bli brukt til å sende informasjon trenger heisen kun å sitte og vente en liten stund til den har fått nok informasjon fra "hjerterytmene" til de to andre heisene til at den kan begynne å operere igjen.


### Hvordan skal vi implementere at den heisen som får trykket på knappen ikke kommer frem?
Barnevaktsystem: Heisen som tar orderen sender et par UDP-broadcast om at «jeg tar den», dette skal aktivere en watchdog timer hos de to andre heisene. Hvis denne går ut, skal en av de andre heisene basert på optimal tildelingsalgorimte, sende en til. 

Forslag til statemachine
========================

Starter watchdog timer, hvis mottar at en heis beveger seg (skal gjøre uansett)
En slags buffer som holder hvilke etasjer heisen er tildelt. Denne trenger kun å ligge hos heisen det gjelder. Her kan vi bruk minheap og maxheap

State boot:
Boot heartbeat (Heartbeat pakkene inneholder informasjon om state og etasje)
Lagrer informasjon om heisene i eks struct heis 1, 2 og 3 som skal inneholde informasjon om etasje og state for de andre. 

State Idle:
Venter på knappetrykk på heispannelet sitt
Bedskjed om å bevege seg til et call

State Master:
Inntreffer når eksternt heispanelet blir trykket på
Kjør tildelingsalgoritme for orderen (Denne prioriterer retning over etasje)
Send udp-broadcast om hvilken heis som skal ta den
Hvis den ikke tar oppgaven, return to idle
Hvis den tar oppgaven, gå til Movement

State Hall_button_Movement:
Sender UDP-packer med info om at den beveger seg til gitt etasje + at det er en Hall call ()
Skrur på knappelyset på utvendig panel
Faktisk beveger seg

State wait_for_passenger:
Inntreffer når heisen har kommet til ønskelig etasje
Holde døren åpen i en gitt lengde tid
Skru på at dørlys er åpent
Starte en timer for at passasjeren skal trykke på en knapp og hvis den timer ut, skal heisen gå tilbake til idle. 
Hvis person trykker innenfor gitt tid, stoppes watchdog og 

State Cab_button_Movement:
Sender UDP-pakker med info om at den beveger seg til en gitt etajse + at det er en Cab call
Skrur på lyset innvendig
Heisen beveger seg 

State reconnect:
Inntreffer når heisen får kontakt med de andre gjennom heartbeat
Hente info om:
Hvor de andre heisene er
Hvilken state de er i
Send info om hvor heisen er atm
Hvis vi velger å bruke heartbeat til å inneholde etasje og state for hver heis, vil alt dette hentes automatisk ved gjennoppkobling. 

State offline: 
Inntreffer hvis den ikke for informasjon fra begge heiser, over en gitt tidsperiode
Starte single elevator mode
Prøve å gjenopprette connection

Nødvendige moduler
==================
En datatype modul som holder datatypene (Elevator_datatypes)
En timer modul som kan holde funksjoner slik som elevator_stuck og elevator_waiting_for_person_timer
Elevator driveren som inneholder funksjonene til å faktisk sette retning, skru på lys osv.
Request modul som holder funksjoner for å tildele hvilken heis som skal ta hva optimalt
En nettverksmodul som inneholder nettverksfunksjonene
En state machine modul som inneholder informasjon om de forskjellige statesene

Scenario 1: En komplett gjennomføring av en ordre uten feil
===========================================================

1. Systemet booter og alle heisene får startet heartbeat systemene sine så alle kommuniserer
2. En av heispanelene blir trykket på (eks heis 1 sitt)
3. Da blir den heisen master (Gjelder heis 1), kjører optimal tildelingsalgoritme og broadcaster til valgt heis om at den skal ta oppgaven, men mindre seg selv er valgt
4. Heisen finner ut av seg selv er best, og flytter seg selv da over til State Movement.
5. Her sender den et par UDP-pakker om at den skal bevege seg hit. De andre heisene starter watchdog timer i bakgrunnen (Go watchdog) som følger med på at den kommer frem 
6. Heisen kommer frem og sender UDP-pakker om at den er fremme i den etasjen som de andre må lagre. Dette stopper watchdogen hos de to andre ettersom heisen har kommet trykt frem. 
7. Heisen går til state wait_for_passenger
8. Personen trykker så på en etasje som er i samme retning heisen beveger seg.
9. Heisen tilbake til state Movement og gjør alt den skal der.
10. Heisen går til state_wait_for passenger, personen går ut av heisen, timer for idle timer ut og heisen returnerer til idle så den kan bli brukt igjen. 

Scenario 2: En gjennomføring hvor heisen faller ut etter at panelet blir trykket på
===================================================================================

1. Systemet booter og alle heisene får startet heartbeat systemene sine så alle kommuniserer
2. En av heispanelene blir trykket på (eks heis 1 sitt)
3. Da blir den heisen master (Gjelder heis 1), kjører optimal tildelingsalgoritme og broadcaster til valgt heis om at den skal ta oppgaven, men mindre seg selv er valgt
4. Hvis heisen disconnecter før heis 1 får sendt info om at jeg beveger meg, så vil ikke de andre sine watchdog timere starte og den vil reconnecte på et punkt. 
5. Hvis heisen disconnecter etter heis 1 får sendt info om at jeg beveger meg, så vil watchdog timeren til de andre heisene start og vil løpe ut. 
6. Da vil en av de andre heisene også bevege seg til gitt etasje som betyr at to stykker vil dukke opp. Hvilken som går bestemmes av hvilken watchdog timer som går ut først.
7. Begge heisene som ankommer will gå inn i state wait_for_person. Den ene vil få et input fra person og bevege seg videre, den andre vil time ut og gå tilbake til idle. 
8. Person vil bli levert til riktig etasje av heisen han gikk inn i, heisen går til state wait_for_person og så idle. 

Scenario 3: En gjennomføring hvor en heis krasjer med en aktive Cab request (indre panel trykket på), antar at den ikke er mellom etasjer
=========================================================================================================================================

1. Systemet booter og alle heisene får startet heartbeat systemene sine så alle kommuniserer
2. En av heispanelene blir trykket på (eks heis 1 sitt)
3. Da blir den heisen master (Gjelder heis 1), kjører optimal tildelingsalgoritme og broadcaster til valgt heis om at den skal ta oppgaven, men mindre seg selv er valgt
4. Heisen finner ut av seg selv er best, og flytter seg selv da over til State Movement.
5. Her sender den et par UDP-pakker om at den skal bevege seg hit. De andre heisene starter watchdog timer i bakgrunnen (Go watchdog) som følger med på at den kommer frem
6. Heisen kommer frem og sender UDP-pakker om at den er fremme i den etasjen som de andre må lagre. Dette stopper watchdogen hos de to andre ettersom heisen har kommet trykt frem. 
7. Heisen går til state wait_for_passenger.
8. Personen trykker så på en etasje som er i samme retning heisen beveger seg.
9. Heisen slutter å funke i den forstand at av en eller annen grunn får den ikke beveget seg (obstruksjon)
10. Her burde heisen åpne dørene igjen og tilkalle en ny heis
11. Personen må da ut av heisen og benytte en annen 
12. Heisen som ankommer går i state_wait_for_passenger
13. Person går inn i ny heis og velger etasje
14. Heisen går i cab_call_movement og beveger seg til riktig etajse
15. Person går ut, heis går til state_wait_for_passenger, timer ut og går til idle.

Scenario 4: A scenario of packet loss that demonstrates that no lights turn on without order execution (samme som 1 grunnet måten lys blir skurdd på i states)
======================================================================================================

1. Systemet booter og alle heisene får startet heartbeat systemene sine så alle kommuniserer
2. En av heispanelene blir trykket på (eks heis 1 sitt)
3. Da blir den heisen master (Gjelder heis 1), kjører optimal tildelingsalgoritme og broadcaster til valgt heis om at den skal ta oppgaven, men mindre seg selv er valgt
4. Heisen finner ut av seg selv er best, og flytter seg selv da over til State Movement.
5. Her sender den et par UDP-pakker om at den skal bevege seg hit. De andre heisene starter watchdog timer i bakgrunnen (Go watchdog) som følger med på at den kommer frem
6. Heisen kommer frem og sender UDP-pakker om at den er fremme i den etasjen som de andre må lagre. Dette stopper watchdogen hos de to andre ettersom heisen har kommet trykt frem. 
7. Heisen går til state wait_for_passenger
8. Personen trykker så på en etasje som er i samme retning heisen beveger seg.
9. Heisen tilbake til state Movement og gjør alt den skal der.
10. Heisen går til state_wait_for passenger, personen går ut av heisen, timer for idle timer ut og heisen returnerer til idle så den kan bli brukt igjen. 

