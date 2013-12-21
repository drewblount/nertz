package main

import (
    websocket "code.google.com/p/go.net/websocket"
    "fmt"
    "log"
    "os"
    "strconv"
    "bufio"
    "strings"
    "nertz"
	"errors"
)

func Credentials() (string, string) {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter Username: ")
    username, _ := reader.ReadString('\n')

    fmt.Print("Enter Password: ")
    password, _ := reader.ReadString('\n')

    return strings.TrimSpace(username), strings.TrimSpace(password) // ReadString() leaves a trailing newline character
}

var faceMap = map[int]string {
    1  : "A"
    10 : "T"
    11 : "J"
    12 : "Q"
    13 : "K"
}

var invFaceMap = map[string]int {
    "A" : 1
    "T" : 10
    "J" : 11
    "Q" : 12
    "K" : 13
}

var suitMap = map[int]string {
    1 : "h"
    2 : "s"
    3 : "d"
    4 : "c"
}

var invSuitMap = map[string]int {
    "h" : 1
    "s" : 2
    "d" : 3
    "c" : 4
} 


func CardToString(card *Card) (string) {
    face := card.Value
    suit := card.Suit
    str := "%v"
    if face > 1 && face < 10 {
        str = fmt.Sprintf("%v", face)
    }
    else {
    	str = fmt.Sprintf(faceMap[face], "%v")
    }
    str = fmt.Sprintf(str, suitMap[suit])
    return str
}

// Takes a string card-name like Ah, 5c, Td and returns the ints
// describing the card, here (1, 1), (5, 4), (10, 3)
func StringToCard(string str) (int int) {
	if len(str) != 2 {
		return errors.New("Card name is an incorrect length. Try something of the like Kh, Tc, 2s")
	}
	faceS := str[0:1] // the first character
	suitS := str[1:]  // the second
		
	faceI := invFaceMap(faceS)
	// if the face isn't a royal card, ace or 10, faceI is still 0
	var faceCheck string	
	for i := 2; i < 10 && faceI == 0; i++ {
		faceCheck = fmt.Sprintf("%d", i)
		if faceCheck == faceS {
			faceI = i
		}
	}
	if faceI == 0 {
		return errors.New("Invalid face specification. Valid faces are 2, 3, 4, 5, 6, 7, 8, 9, T, J, Q, K")
	}
	
	suitI := invSuitMap(suitS)
	if suitI == 0 {
		return errors.New("Invalid suit specification. Valid suits are h, s, c, d")
	}
		
	return(faceI, suitI)
}

// Prints a card stack as shown in the mockup. 
// toShow = 1:  [[[[ 4h ]
// toShow = 3:  [[ 6d[ 5c[ 4h ]
func PrintCardStack(cs *list.List, toShow int) {
    len := cs.Len()
    stack := "%v" // the %v allows for easy appending later
    for e := cs.Back() ; e != nil ; e = e.Prev() {
        stack = fmt.Sprintf(stack,"[%v")
		if len - toShow <= 0 {
            stack = fmt.Sprintf(stack, " " + CardString(e.Value) + "%v")
        }
        len--
    }
    stack = fmt.Sprintf(stack, " ]")
    if stack == " ]" {
       fmt.Println("[ -empty- ]")
    } else {
       fmt.Println(stack)
    }
}

func PrintHand(hand *Hand) {
    fmt.Println("River:")
    for pile := range hand.River {
    	PrintCardStack(pile, pile.Len())    }
    // to print the stream and streampile, first print |stream|*"["
    
    fmt.Println("Stream:")
    for i := 0; i < hand.Stream.Len ; i++ {
    	fmt.Print("[")
    }
    PrintCardStack(hand.Streampile, hand.Streampile.Len())

    fmt.Println("Nertz Pile:")
    PrintCardStack(hand.NertzPile, 1)
}

func PrintLake(lake Lake) {
	fmt.Println("Lake (only nonempty piles shown):")
	for i, pile := range lake {
		fmt.Printf("  %d:", i)
		PrintCardStack(pile, 1)
	} 
}

func PrintGame(p *Player) {
	fmt.Println("THE WONDERFUL WORLD OF NERTZ")
	PrintLake(p.Lake)
	PrintHand(p.Hand)
}

// returns not the index, but the pile's length - index, i.e.
// the numcards in the sub-pile with this one at the top
// This is for compatability with AJ's Transaction etc
func FindCardInPile(face, suit int, pile *list.List) int {
    count := 0
    for e := pile.Front(); e != nil; e = e.next {
        if e.Value.Value == face && e.Value.Suit == suit {
	    return count
	}
        count ++
    }
    return -1
}

// Looks for a card in the given hand and returns (pilename, pilenum, 
// numcards), with numcards being the size of the subpile starting
// on the sought-after card
func FindCardInHand(cardname string, hand *Hand) (string, int, int) {
    face, suit := StringToCard(cardname)
	pile := "error"
    position := -1
     
    for i, RivPile := range(hand.River) {
		position = FindCardInPile(face, suit, RivPile)
		if position != -1 {
			pile = "River"
			return(pile, i, position)
		}
    }
    // note: as I've written this, if only one card if face-up, that's the last card
    c := hand.Nertzpile.Back().Value
    if c.Value == face && c.Suit == suit {
        return("Nertzpile", 0, 0)
    }
    c = hand.Streampile.Back().Value
    if c.Value == face && c.Suit == suit {
        return("Streampile", 0, 0)
    }
     	 
    // if the card is not to be found, returns ("error", -1, -1)
    return(pile, -1, position)
}

// This is the main function used by a player during a game.
// it would be used like mv(3h, 4s), and only concerns the piles
// in your hand (i.e. the lake is off-limits). You don't have to
// specify each card's location; the function finds them as long
// as they represent a legal move

func (h *Hand) mv(this, underThat string) {
	from, fpilenum, numcards := FindCardInHand(this, h)
	toInfo := FindCardInHand(underThat, h)
	to, tpilenum := toInfo[0], toInfo[1]
	
	h.Transaction(from, fpilenum, to, tpilenum, numcards)
}

// Like the above function, but puts a card on an indexed Lake
// pile

func (p *Player) lake(cardname string, lakepile int) error {
	from, fpilenum, numcards := FindCardInHand(this, h)
	if numcards != 1 {
		return errors.New("You can only move uncovered cards from your hand into the lake")
	}
	p.Hand.Transaction(from, fpilenum, "Lake", lakepile, 1)
}

// Fills an empty space in the river with the top card from the nertz pile
func (h *Hand) fish() error {
	emptyRivPile := -1
	for i := range(Hand.River) {
		if len(Hand.River[i]) == 0 {
			emptyRivPile = i
		}
	}
	if emptyRivPile == -1 {
		return errors.New("You need an empty river to fish. Weird. The bodies-of-water metaphor sorta breaks down here.")
	}
	// the two 1s are unused arguments here, but I
	// filled them with logical values anyway
	h.Transaction("Nertzpile", 1, "River", emptyRivPile, 1)
}

// Moves the Streampile to the bottom of the Stream and makes a new
// Streampile (this is the basic "draw" in the game)
func (h *Hand) draw() {
	h.Transaction("Streampile", 1, "Stream", 1, len(h.Streampile))
	h.Transaction("Stream", 1, "Streampile", 1, 3)
}

func main() {

    if len(os.Args) != 3 {
        fmt.Fprintf(os.Stderr,"usage: %v <host> <port>\n", os.Args[0])
        return
    }

    host := os.Args[1]
    port, err := strconv.Atoi(os.Args[2])
    if err != nil {
        log.Fatal(err)
    }

    origin := "http://localhost/"
    wsurl := fmt.Sprintf("ws://%v:%v/ws", host, port)
    gameurl := fmt.Sprintf("http://%v:%v/move", host, port)

    ws, err := websocket.Dial(wsurl, "", origin)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    name, password := Credentials()
    err = websocket.JSON.Send(ws, nertz.Credentials{ name, password, })
    if err != nil {
        panic("JSON.Send: " + err.Error())
    }
    player :=  NewPlayer(name, gameurl, ws)

    fmt.Fprintf(os.Stdout, "Client connected to %v:%v...\n", host, port)

    go reader(ws, ch)
    for {
        select {
        case msg := <-ch:
            fmt.Printf(msg)
        default:
            reader := bufio.NewReader(os.Stdin)
            fmt.Print("Server gets: ")
            msg, err := reader.ReadString('\n')
            msg = strings.TrimSpace(msg)
            if err != nil {
                log.Fatal(err)
            }

            err = websocket.Message.Send(ws, msg)
            if err != nil {
                log.Fatal(err)
            }
        }
    }
}
