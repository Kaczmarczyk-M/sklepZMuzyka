package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/srinathgs/mysqlstore"
)

type Album struct {
	ID     int32   `json:"ID"`
	Title  string  `json:"Title"`
	Artist string  `json:"Artist"`
	Price  float64 `json:"Price"`
}

type Customer struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
	Pass  string `json:"pass"`
}

var db *sql.DB

var store *mysqlstore.MySQLStore

// var sessionID string

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// var albums []Album

// albumsByArtist queries for albums that have the specified artist name.
// func albumsByArtist(name string) ([]Album, error) {
// 	// An albums slice to hold data from returned rows.
// 	var albums []Album
// 	rows, err := db.Query("SELECT * FROM album WHERE artist = ?", name)
// 	if err != nil {
// 		return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
// 	}
// 	defer rows.Close()
// 	// Loop through rows, using Scan to assign column data to struct fields.
// 	for rows.Next() {
// 		var alb Album
// 		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
// 			return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
// 		}
// 		albums = append(albums, alb)
// 	}
// 	if err := rows.Err(); err != nil {
// 		return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
// 	}
// 	return albums, nil
// }

// albumByID queries for the album with the specified ID.
// func albumByID(id int64) (Album, error) {
// 	// An album to hold data from the returned row.
// 	var alb Album
// 	row := db.QueryRow("SELECT * FROM album WHERE id = ?", id)
// 	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
// 		if err == sql.ErrNoRows {
// 			return alb, fmt.Errorf("albumsById %d: no such album", id)
// 		}
// 		return alb, fmt.Errorf("albumsById %d: %v", id, err)
// 	}
// 	return alb, nil
// }

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
		User:   "suser",
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
	fmt.Println("DB: Connected!")
}

func handleRequests() {
	myRouter := mux.NewRouter()
	myRouter.HandleFunc("/", loginPage)
	myRouter.HandleFunc("/home", homePage)
	myRouter.HandleFunc("/help", help)
	myRouter.HandleFunc("/albums", handlerReturnAllAlbums)
	myRouter.HandleFunc("/album/{id}", returnSingleAlbum)
	myRouter.HandleFunc("/newalbum", createNewAlbum).Methods("POST")
	myRouter.HandleFunc("/custom", wypiszwszystkich)
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	//returns session called  sessionID
	session, err := store.Get(r, "sessionID")
	if err != nil {
		fmt.Printf("err homepage get sessionID: %v\n", err)
	}
	if session.Values["isLogged"] == "true" {
		fmt.Println("Endpoint: Home Page")
		http.ServeFile(w, r, "./static/index.html")
		return
	} else {
		fmt.Println("Not logged")
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func help(w http.ResponseWriter, r *http.Request) {
	a, _ := store.Get(r, "sessionID")
	fmt.Fprintf(w, "Is logged? : %v", a.Values["isLogged"])

}
func loginPage(w http.ResponseWriter, r *http.Request) {
	//returns session called  sessionID
	session, err := store.Get(r, "sessionID")
	if err != nil {
		fmt.Printf("err loginPage: %v\n", err)
	}
	session.Values["isLogged"] = "false"
	session.Save(r, w)
	if r.Method != "POST" {
		fmt.Println("r.method post 134")
		http.ServeFile(w, r, "./static/log.html")
		return
	}
	var customer Customer
	customer.Email = r.FormValue("Email")
	customer.Pass = r.FormValue("Pass")
	var databaseUsername string
	var databasePassword string
	err = db.QueryRow("SELECT email, pass FROM customer WHERE email=?", customer.Email).Scan(&databaseUsername, &databasePassword)
	if err != nil {
		fmt.Printf("err wrong query: %v\n", err)
		session.Values["isLogged"] = "false"
		err = session.Save(r, w)
		fmt.Println("Session Saved")
		if err != nil {
			fmt.Printf("err: session save %v\n", err)
		}
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if customer.Pass == databasePassword {
		fmt.Println("email and pass are correct")
		session.Values["isLogged"] = "true"
		// fmt.Printf("######## po zmianie session.Values[\"isLogged\"]: %v\n", session.Values["isLogged"])
		err = session.Save(r, w)
		fmt.Println("Session Saved")
		if err != nil {
			fmt.Printf("err: session save %v\n", err)
		}
		http.Redirect(w, r, "/home", http.StatusFound)

	} else {
		fmt.Println("Password is incorect")
		session.Values["isLogged"] = "false"
		err = session.Save(r, w)
		fmt.Println("Session Saved")
		if err != nil {
			fmt.Printf("err: session save %v\n", err)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	}

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

//Colects data from form input and inserts values to DB
func createNewAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var album Album
		album.Title = r.FormValue("Title")
		album.Artist = r.FormValue("Artist")
		album.Price, _ = strconv.ParseFloat(r.FormValue("Price"), 64)
		if album.Title != "" && album.Artist != "" && album.Price != 0 {
			idOfAlbum, err := addAlbum(album)
			if err != nil {
				fmt.Printf("CreateNewAlbum: %v\n", err)
			}
			fmt.Printf("ID of new added record: %d\n", idOfAlbum)
		} else {
			fmt.Println("Error input not filled")
		}
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}
	fmt.Fprintln(w, "Error: Only POST")
	fmt.Println("Error: Only POST")
}

func wypiszwszystkich(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "%v\n", testWypiszWszystkichUzytkownikow())
	json.NewEncoder(w).Encode(testWypiszWszystkichUzytkownikow())
}

func testWypiszWszystkichUzytkownikow() []Customer {
	// An albums slice to hold data from returned rows.
	var Customers []Customer
	rows, err := db.Query("SELECT * FROM customer;")
	if err != nil {
		fmt.Printf("err Querry: %v\n", err)
		return nil
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var cust Customer
		if err := rows.Scan(&cust.Id, &cust.Email, &cust.Pass); err != nil {
			fmt.Printf("error scanning: %v\n", err)
		}
		Customers = append(Customers, cust)
	}
	if err := rows.Err(); err != nil {
		_ = fmt.Errorf("returnAllAlbums: %v", err)
		return nil
	}
	return Customers
}
func main() {
	//connecting to DB
	connectWithDB("shop")
	var err error
	// sessionID, err = GenerateRandomString(32)
	if err != nil {
		panic(err)
	}
	store, err = mysqlstore.NewMySQLStore("suser:password@tcp(127.0.0.1:3306)/shop?parseTime=true&loc=Local", "session", "/", 3600, []byte("SecretKey"))
	if err != nil {
		panic(err)
	}
	defer store.Close()
	//starting server
	handleRequests()

}
