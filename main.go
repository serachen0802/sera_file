package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

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
		// fmt.Printf(string(p))

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

func savePicture(w http.ResponseWriter, r *http.Request) {
	FileName, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	// FileName := fmt.Sprint("picture/test", time.Now().Unix(), ".jpg")
	// err := os.WriteFile(FileName, img99, os.ModePerm)
	// if err != nil {
	// 	if os.IsNotExist(err) == true {
	// 		os.Mkdir("picture", os.ModePerm)
	// 	}
	// 	os.WriteFile(FileName, img99, os.ModePerm)
	// }
	// 儲存檔案(ModePerm 預設權限)

	DB, err := sql.Open("mysql", "root:qwe123@tcp(127.0.0.1:3306)/sera?charset=utf8")
	if err != nil {
		fmt.Println(err)
	}

	_, err = DB.Exec("Insert INTO picture(name,picture) values(?,?)", FileName, img99)
	if err != nil {
		fmt.Println(err)
	}
	DB.Close()
}

func saveByFront(w http.ResponseWriter, r *http.Request) {
	stringTypePhoto, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	saveImg, err := base64.StdEncoding.DecodeString(string(stringTypePhoto))
	if err != nil {
		fmt.Println(err)
	}
	FileName := fmt.Sprint("picture/test", time.Now().Unix(), ".jpg")
	os.WriteFile(FileName, saveImg, os.ModePerm)
}

func search(w http.ResponseWriter, r *http.Request) {
	fileName, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(fileName))

	DB, err := sql.Open("mysql", "root:qwe123@tcp(127.0.0.1:3306)/sera?charset=utf8")
	if err != nil {
		fmt.Println(err)
	}

	var picture []byte
	data := DB.QueryRow("SELECT picture FROM picture WHERE name =?", fileName)
	err = data.Scan(&picture)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(picture)
	jsonImg, err := json.Marshal(picture)
	if err != nil {
		fmt.Println(err)
	}
	DB.Close()

	w.Write(jsonImg)
}

type IndexData struct {
	Title string
}

var img99 []byte

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/save", savePicture)
	http.HandleFunc("/saveByFront", saveByFront)
	http.HandleFunc("/search", search)

	open := make(chan int)
	take := make(chan []byte)

	DB, err := sql.Open("mysql", "root:qwe123@tcp(127.0.0.1:3306)/sera?charset=utf8")
	if err != nil {
		fmt.Println(err)
	}

	_, tableCheck := DB.Query("select * from picture;")
	if tableCheck == nil {
		fmt.Println("table is there")
	} else {
		fmt.Println("table not there")
		sql := `CREATE TABLE picture (
			id int(100) NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name varchar(20) NOT NULL,
			picture longblob NOT NULL,
			created_time datetime DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
		_, err = DB.Exec(sql)
		if err != nil {
			fmt.Println(err)
		}
		DB.Close()
	}

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

			time.Sleep(time.Millisecond * 100)

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
			img99 = img2.GetBytes()
			img2.Close()
		}()
	}

}
