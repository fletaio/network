package simulations

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

//Msg is the structure that defines the command.
type Msg struct {
	Command string
	Num     int
}

var dataFuncs map[string]map[string]func() []string

func init() {
	dataFuncs = map[string]map[string]func() []string{}
}

//Visualization is a map that stores data to be visualized.
type Visualization map[string]map[string][]string

var sender <-chan Visualization
var receiver chan<- Msg

//VisualizationStart is start the visualization
func VisualizationStart(nodeAdder chan<- Msg) {
	http.HandleFunc("/ws", wsHandler)
	// http.HandleFunc("/", rootHandler)
	http.Handle("/", http.FileServer(http.Dir("./git.fleta.io/fleta/mocknet/simulations/html/")))

	data := make(chan Visualization)
	sender = data
	receiver = nodeAdder

	go func() {
		for {
			sendData := make(map[string]map[string][]string)
			for nodeID, node := range dataFuncs {
				nodeData := map[string][]string{}
				for dataName, f := range node {
					nodeData[dataName] = f()
				}
				sendData[nodeID] = nodeData

			}
			data <- sendData
			time.Sleep(time.Second)
		}
	}()

	panic(http.ListenAndServe(":8080", nil))
}

//AddVisualizationData adds a function that returns visualization data.
func AddVisualizationData(nodeID string, dataName string, dataFunc func() []string) {
	node, has := dataFuncs[nodeID]
	if !has {
		node = map[string]func() []string{}
		dataFuncs[nodeID] = node
	}
	node[dataName] = dataFunc
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	pwd, _ := os.Getwd()
	content, err := ioutil.ReadFile(pwd + "/git.fleta.io/fleta/mocknet/simulations/html/index.html")
	if err != nil {
		fmt.Println("Could not open file.", err)
	}
	fmt.Fprintf(w, "%s", content)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	go read(conn)
	go write(conn)
}

func read(conn *websocket.Conn) {
	m := Msg{}

	for {
		err := conn.ReadJSON(&m)
		if err != nil {
			fmt.Println("Error reading json.", err)
			break
		}
		receiver <- m
	}
}

func write(conn *websocket.Conn) {
	for {
		data := <-sender
		if err := conn.WriteJSON(data); err != nil {
			fmt.Println(err)
			break
		}
	}
}
