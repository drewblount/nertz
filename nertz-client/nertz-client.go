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


func CardString(card *Card) (string) {
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

// The way I read what used to be below was that
// it'd print the mirror of what we had in our mockup, so I unmirrored it
func PrintCardStack(cs *list.List, toShow int) {
    len := cs.Len()
    stack := "%v"
    for e := cs.Front() ; e != nil ; e = e.Next() {
        stack = fmt.Sprintf(stack,"[%v")
	if len - toShow <= 0 {
	    // on the following line, I'm not sure how to get the card component of e
            stack = fmt.Sprintf(stack, " ", CardString(e.Value), "%v")
        }
        len--
    }
    stack = fmt.Sprintf(stack, " ]")
    if stack == "[ ]" {
       fmt.Println("[ -empty- ]")
    } else {
       fmt.Println(stack)
    }
}

func PrintHand(hand *Hand) {
    fmt.Println("River:")
    for pile := range hand.River {
    	PrintCardStack(pile, pile.Len())
    }
    // to print the stream and streampile, first print |stream|*"["
    
    fmt.Println("Stream:")
    for i := 0; i < hand.Stream.Len ; i++ {
    	fmt.Print("[")
    }
    PrintCardStack(hand.Streampile, hand.Streampile.Len())

    fmt.Println("Nertz Pile:")
    PrintCardStack(hand.NertzPile, 1)
}

FindCardInPile(face, suit int, pile *list.List) int {
    count := 0
    for e := pile.Front(); e != nil; e = e.next {
        if e.Value.Value == face && e.Value.Suit == suit {
	    return count
	}
        count ++
    }
    return -1
}

// returns (pilename, cardnum), with pilename taken from only
// the piles accessible for moving cards. For piles such as the nertz
// pile and the stream, only one card is accessible so only one card
// is checked, and cardnum is 0 if that is the card being sought
func FindCardInHand(face, suit int, hand *Hand) (string, int) {
    pile := "error"
    position := -1
     
    for i, RivPile := range(hand.River) {
   	position = FindCardInPile(face, suit, RivPile)
	if position != -1 {
	    pile = fmt.Sprintf("River[%d]",i)
	    return(pile, position)
	}
    }
    // note: as I've written this, if only one card if face-up, that's the last card
    c := hand.Nertzpile.Back().Value
    if c.Value == face && c.Suit == suit {
        return("Nertzpile", 0)
    }
    c = hand.Streampile.Back().Value
    if c.Value == face && c.Suit == suit {
        return("Streampile", 0)
    }
     	 
    // if the card is not to be found, returns (error, -1)
    return(pile, position)
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
