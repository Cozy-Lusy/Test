package main

import (
	"database/sql"
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
	ID      int64
	Image   string
	Preview string
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

	img = resize.Resize(100, 100, img, resize.Bilinear)
	imgPreview, err := os.Create("./images/preview/" + handle.Filename)
	if err != nil {
		log.Fatal(err)
	}
	defer imgPreview.Close()

	jpeg.Encode(imgPreview, img, nil)
	imgPreview.Close()

	saveImg := Image{}

	saveImg.Image = "./img/" + handle.Filename
	saveImg.Preview = "./img/preview/" + handle.Filename

	result, err := db.Exec("insert into images (image, preview) values ($1 , $2)", saveImg.Image, saveImg.Preview)
	if err != nil {
		log.Fatal(err)
	}

	saveImg.ID, _ = result.LastInsertId()
}

//Добавим картинки в бд
func uploadImage(w http.ResponseWriter, r *http.Request) {

}

func main() {

	http.Handle("/img/", http.StripPrefix("/images", http.FileServer(http.Dir("./images"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/upload", uploadHandler)

	http.ListenAndServe(":4000", nil)

}
