package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nfnt/resize"
)

var db *sql.DB

type Image struct {
	ID       int64
	Original string
	Preview  string
}

func init() {
	var err error
	db, err = sql.Open("sqlite3", "images.db")
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func main() {

	http.Handle("/images/", http.StripPrefix("/images", http.FileServer(http.Dir("./images"))))

	http.HandleFunc("/", sendClient)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/delete", deleteImage)

	http.ListenAndServe(":4000", nil)

}

//Загрузка изображения с клиента
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	file, handle, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}

	dstImg, err := os.Create("./images/" + handle.Filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dstImg.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	jpeg.Encode(dstImg, img, &jpeg.Options{jpeg.DefaultQuality})

	img = resize.Resize(150, 100, img, resize.Bilinear)
	imgPreview, err := os.Create("./images/preview/" + handle.Filename)
	if err != nil {
		log.Fatal(err)
	}
	defer imgPreview.Close()

	jpeg.Encode(imgPreview, img, nil)
	imgPreview.Close()

	saveImg := Image{}

	saveImg.Original = "./image/" + handle.Filename
	saveImg.Preview = "./image/preview/" + handle.Filename

	result, err := db.Exec("INSERT INTO images (image, preview) VALUES ($1 , $2)", saveImg.Original, saveImg.Preview)
	if err != nil {
		log.Fatal(err)
	}

	saveImg.ID, _ = result.LastInsertId()
}

//Добавим картинки в бд
func uploadImage(w http.ResponseWriter, r *http.Request) {

}

//Json отправка на клиент содержимого бд
func sendClient(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	listImg := []Image{}

	rows, err := db.Query("SELECT * FROM images")
	if err != nil {
		fmt.Println(err)
		return
	}

	for rows.Next() {
		img := Image{}
		err := rows.Scan(&img.ID, &img.Original, &img.Preview)
		if err != nil {
			fmt.Println(err)
			return
		}

		listImg = append(listImg, img)
	}

	js, err := json.Marshal(listImg)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

//Удаление из бд по Id
func deleteImage(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodDelete {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	id := r.URL.Query().Get("id")

	if id == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	_, err := db.Exec("DELETE FROM images WHERE id = $1", id)
	if err != nil {
		log.Println(err)
	}

	http.Redirect(w, r, "/", 301)
	fmt.Fprintf(w, "Image %s deleted successfully\n", id)
}
