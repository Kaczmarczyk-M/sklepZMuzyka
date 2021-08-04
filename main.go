package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Album struct {
	ID     int32   `json:"ID"`
	Title  string  `json:"Title"`
	Artist string  `json:"Artist"`
	Price  float64 `json:"Price"`
}

var db *sql.DB

// var albums []Album

// albumsByArtist queries for albums that have the specified artist name.
func albumsByArtist(name string) ([]Album, error) {
	// An albums slice to hold data from returned rows.
	var albums []Album
	rows, err := db.Query("SELECT * FROM album WHERE artist = ?", name)
	if err != nil {
		return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb Album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
		}
		albums = append(albums, alb)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
	}
	return albums, nil
}

// albumByID queries for the album with the specified ID.
func albumByID(id int64) (Album, error) {
	// An album to hold data from the returned row.
	var alb Album

	row := db.QueryRow("SELECT * FROM album WHERE id = ?", id)
	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
		if err == sql.ErrNoRows {
			return alb, fmt.Errorf("albumsById %d: no such album", id)
		}
		return alb, fmt.Errorf("albumsById %d: %v", id, err)
	}
	return alb, nil
}

// addAlbum adds the specified album to the database,
// returning the album ID of the new entry
func addAlbum(alb Album) (int64, error) {
	result, err := db.Exec("INSERT INTO album (title, artist, price) VALUES (?, ?, ?)", alb.Title, alb.Artist, alb.Price)
	if err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}
	return id, nil
}

//connecting to DB
func connectWithDB(nameOfDB string) {
	// Capture connection properties.
	cfg := mysql.Config{
		User:   "newuser",
		Passwd: "password",
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: nameOfDB,
	}
	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")
}

func handleRequests() {
	var dir string
	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/albums", handlerReturnAllAlbums)
	myRouter.HandleFunc("/album/{id}", returnSingleAlbum)
	myRouter.HandleFunc("/album", createNewAlbum).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	Body, err := ioutil.ReadFile("/home/michal/Code/projekt1/static/index.html")
	if err != nil {
		panic(err)
	}
	w.Write(Body)
	fmt.Println("Endpoint: Home Page")

}

//returns slice of all albums in DB
func returnAllAlbums() []Album {
	// An albums slice to hold data from returned rows.
	var albums []Album
	rows, err := db.Query("SELECT * FROM album;")
	if err != nil {
		_ = fmt.Errorf("returnAllAlbums: %v", err)
		return nil
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb Album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			_ = fmt.Errorf("returnAllAlbums: %v", err)
			return nil
		}
		albums = append(albums, alb)
	}
	if err := rows.Err(); err != nil {
		_ = fmt.Errorf("returnAllAlbums: %v", err)
		return nil
	}
	return albums
}

func handlerReturnAllAlbums(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(returnAllAlbums())
}

func returnSingleAlbum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, _ := strconv.ParseInt(vars["id"], 10, 8)
	key32 := int32(key)
	allAlbums := returnAllAlbums()
	for _, singleAlbum := range allAlbums {
		if singleAlbum.ID == key32 {
			json.NewEncoder(w).Encode(singleAlbum)
			return
		}
	}
	fmt.Fprintf(w, "Album not found")

}

func createNewAlbum(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var album Album
	json.Unmarshal(reqBody, &album)
	fmt.Printf(" Title: %s;\n Artist: %s;\n Price: %.2f\n", album.Title, album.Artist, album.Price)
	if album.Title != "" && album.Artist != "" && album.Price != 0 {
		idOfAlbum, err := addAlbum(album)
		if err != nil {
			fmt.Printf("CreateNewAlbum: %v\n", err)
		}
		fmt.Printf("ID of new added record: %d", idOfAlbum)
	} else {
		fmt.Println("Error input not filled")
	}
}
func main() {
	//connecting to DB
	connectWithDB("recordings")

	//starting server
	handleRequests()

}
