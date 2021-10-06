package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = &websocket.Upgrader{}

func wsHandler(w http.ResponseWriter, r *http.Request) {

	con, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer con.Close()

	for {
		_, p, err := con.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf(string(p))

		image, err := ioutil.ReadFile("./plants.jpeg")
		if err != nil {
			fmt.Println(err)
		}

		err = con.WriteJSON(image)
		// base64Img, err := json.Marshal(image)
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// if err := con.WriteMessage(messageType, base64Img); err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
	}
}

type IndexData struct {
	Title string
}

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe(":3000", nil)
}
