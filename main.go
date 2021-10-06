package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func wsHandler(w http.ResponseWriter, r *http.Request) {

	con, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer con.Close()

	for {
		messageType, p, err := con.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf(string(p))

		image, err := ioutil.ReadFile("./plants.jpeg")
		if err != nil {
			fmt.Println(err)
		}

		base64Img, err := json.Marshal(image)
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// base64Img := base64.StdEncoding.EncodeToString(image)

		if err := con.WriteMessage(messageType, base64Img); err != nil {
			fmt.Println(err)
			return
		}
	}
}

type IndexData struct {
	Title string
}

func showPicHandle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	tmpl := template.Must(template.ParseFiles("./index.html"))
	data := new(IndexData)
	data.Title = "Hi"

	tmpl.Execute(w, data)
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", showPicHandle)
	http.ListenAndServe(":3000", nil)
}
