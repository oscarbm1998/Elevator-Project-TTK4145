Sanntidsprogrammering
=====================

### Hva ønsker vi å implementere:
Vi ønsker å implementere en peer-to-peer-nettverksløsning for heissystemet hvor den som får sitt knappepanel trykket på vil fungere som master og enten ta heisen selv, eller delegere videre hvis den ikke har mulighet. 

### Hvordan skal boot-up fungere?
Boot up vil fungere ved at noden som blir booted først vil fungere som master i boot-up prosessen og boote resten. 

### Hvordan skal vi sjekke om heiser faller ut av nettet?
Vi ønsker en UDP-nettverkstopologi som benytter en slags hjerterytme for å sjekke om samtlige heiser er koblet til nettet. Dette gjøres med en UDP-broadcast funksjon som kjører en timer hver gang den mottar en Heart-beat og hvis den ikke har mottatt info fra en spesifikk heis i så, så lang tid kan heisen anta at den har falt ut av nettet. Her er det viktig å tenke på at man ikke overloader nettet samtidig som man sender ofte nok til at packet loss ikke nødvendigvis blir et problem. 

### Hva skal heisen gjøre hvis den faller ut?
Vi ønsker at heisen skal kjøre en slags single elevator mode ettersom hvis en heis faller ut vil ikke knappepanelet til den heisen klare å sende info til resten om at det eksisterer en ordre der. Det er derfor greit at den heisen kan respondere på Calls som den får trykket inn. Denne single elevator mode kan for eksempel aktiveres når heisen har registrert at den har falt ut. 

### Hvordan skal vi implementere at den heisen som får trykket på knappen ikke kommer frem?
Barnevaktsystem: Heisen som tar orderen sender et par UDP-broadcast om at «jeg tar den», dette skal aktivere en watchdog timer hos de to andre heisene. Hvis denne går ut, skal en av de andre heisene basert på optimal tildelingsalgorimte, sende en til. 

Hvordan skal vi implementere at den he
