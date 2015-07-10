package main

import (
	"encoding/json"
	"github.com/fzzy/sockjs-go/sockjs"
	"io/ioutil"
	"log"
	"net/http"
	"net"
	"os"
	"runtime"
	"path"
	"fmt"
	"bufio"
	"strconv"
)

const staticDir = "../client"
const bind = "0.0.0.0:3000"
const PORT = 3333
var message=make(chan string)    

func events_listner() {
    server, err := net.Listen("tcp", ":" + strconv.Itoa(PORT))
    if server == nil {
        panic("couldn't start listening: " + err.Error())
    }
    conns := clientConns(server)
    for {
        go handleConn(<-conns)
	
	
    }

}

func clientConns(listener net.Listener) chan net.Conn {
    ch := make(chan net.Conn)
    i := 0
    go func() {
        for {
            client, err := listener.Accept()
            if client == nil {
                fmt.Printf("couldn't accept: " + err.Error())
                continue
            }
            i++
            //fmt.Printf("%d: %v <-> %v\n", i, client.LocalAddr(), client.RemoteAddr())
            ch <- client
        }
    }()
    return ch
}

func handleConn(client net.Conn){
    b := bufio.NewReader(client)
    for {
        line, err := b.ReadBytes('}')
        if err != nil { // EOF, or worse
            break
        }
        message <- string(line)
    }
}



type Config struct {
	Host  string
	Port  int
	Ident string
	Auth  string
}

func dirname() string {
	_, myself, _, _ := runtime.Caller(1)
	return path.Dir(myself)
}

func readConfig() Config {
	blob, err := ioutil.ReadFile(dirname() + "/" + "config.json")
	checkFatalError(err)

	var conf Config
	err = json.Unmarshal(blob, &conf)
	checkFatalError(err)

	return conf
}

func checkFatalError(err error) {
	if err != nil {
		log.Printf("%s\n", err)
		os.Exit(-1)
	}
}

var sockjsClients *sockjs.SessionPool = sockjs.NewSessionPool()

func dataHandler(s sockjs.Session) {
	sockjsClients.Add(s)
	defer sockjsClients.Remove(s)
	for {
		m := s.Receive()
		if m == nil {
			break
		}
	}
}







func main() {
	

	http.Handle("/", http.FileServer(http.Dir(dirname() + "/" + staticDir + "/")))
	sockjsMux := sockjs.NewServeMux(http.DefaultServeMux)
	sockjsConf := sockjs.NewConfig()
	sockjsMux.Handle("/data", dataHandler, sockjsConf)
	go events_listner()
	go func() {
		  for{ 
		   var msg=<-message
		   //s:=fmt.Sprintf("{\"latitude\": 32.0617, \"type\": \"blacklist\", \"longitude\": 118.7778,\"countrycode\":\"PT\", \"city\":\"Lisboa\"}")
		   sockjsClients.Broadcast([]byte(msg))
		  }
	}()	  
        log.Printf("Binding Honeymap webserver to %s...", bind)
	err := http.ListenAndServe(bind, sockjsMux)
	checkFatalError(err)
}
