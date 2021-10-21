package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"gocv.io/x/gocv"
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

type data struct {
	Image []byte `json:"image"`
}

func takePicture(open chan int, take chan []byte) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		open <- 0
		fmt.Println("拍照")
		img := <-take

		jsonImg, err := json.Marshal(img)
		if err != nil {
			fmt.Println(err)
		}

		rw.Write(jsonImg)
	}
}

type IndexData struct {
	Title string
}

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/ws", wsHandler)

	open := make(chan int)
	take := make(chan []byte)

	go func() {
		http.HandleFunc("/takePicture", takePicture(open, take))
		http.ListenAndServe(":3000", nil) // stuck
	}()

	// 使用channel的方式傳送開啟相機資訊
	for {
		func() {
			deviceID := <-open
			// 啟動相機
			webcam, err := gocv.OpenVideoCapture(deviceID)
			if err != nil {
				fmt.Println(err)
			}
			defer webcam.Close()

			img := gocv.NewMat()
			defer img.Close()

			if ok := webcam.Read(&img); !ok {
				fmt.Printf("cannot read device %v\n", deviceID)
				return
			}

			if img.Empty() {
				fmt.Printf("no image on device %v\n", deviceID)
				return
			}

			// 將原圖取得圖片byte做傳輸
			img2, _ := gocv.IMEncode(".jpg", img)
			// 將圖片用channel 傳回
			take <- img2.GetBytes()
			img2.Close()
			// 除吋檔案(ModePerm 預設權限)
			os.WriteFile("test", img2.GetBytes(), os.ModePerm)
		}()
	}
}
