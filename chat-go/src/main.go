package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) //подключенные клиенты
var broadcast = make(chan Message)           //broadcast клиент действует как очередь сообщений

//Настройка обновления
var upgrader = websocket.Upgrader{}

//Message определяет объект для хранения сообщений
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil) // обновим начальный GET запрос на вебсокет
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clients[ws] = true //регестрируем нового клиента

	for {
		var msg Message

		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}

		//отправляем принятое сообщение на broadcast
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast //берем следующее сообщение из канала broadcast

		for client := range clients { //отправляем его каждому подключенному клиенту
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	fs := http.FileServer(http.Dir("../public")) //создаем простой файловый сервер
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections) //настройка маршрута вебсокета

	go handleMessages() //начинаем слушать входящие сообщения чата

	log.Println("Server started on :8000...")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
