package concentrator

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type msg struct {
	Num int
}

type Visualization map[string]map[string][]string

var receiver <-chan Visualization

func VisualizationStart(data <-chan Visualization) {
	http.HandleFunc("/ws", wsHandler)
	// http.HandleFunc("/", rootHandler)
	http.Handle("/", http.FileServer(http.Dir("./fleta/samutil/concentrator/html/")))

	receiver = data

	panic(http.ListenAndServe(":8080", nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	pwd, _ := os.Getwd()
	content, err := ioutil.ReadFile(pwd + "/fleta/samutil/concentrator/html/index.html")
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

	go echo(conn)
}

func echo(conn *websocket.Conn) {
	for {
		data := <-receiver
		if err := conn.WriteJSON(data); err != nil {
			fmt.Println(err)
		}
	}
}
