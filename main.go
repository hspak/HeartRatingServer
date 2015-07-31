package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)
import _ "github.com/go-sql-driver/mysql"
import "github.com/go-martini/martini"
import "github.com/martini-contrib/render"

type ApiSession struct {
	U      User      `json:"user"`
	S      []Session `json:"sessions"`
	Status string    `json:"status"`
}

type ApiUsers struct {
	U      []User `json:"users"`
	Status string `json:"status"`
}

type PostData struct {
	Heart    int    `json:"heart-score"`
	Duration int    `json:"watch-time"`
	Show     string `json:"show"`
	Title    string `json:"title"`
	User     string `json:"user"`
}

type indexPageData struct {
	Fill []userPageData
}

type userPageData struct {
	User     string
	Title    string
	Show     string
	Heart    int
	Duration int
}

type User struct {
	Id   int
	Name string
}

type Session struct {
	Title    string
	Show     string
	Heart    int
	Duration int
}

func setup_db() *sql.DB {
	db, err := sql.Open("mysql", "root@tcp(localhost:3306)/")
	if err != nil {
		panic(err.Error())
	}

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	_, err = db.Exec(`DROP DATABASE IF EXISTS HeartRating`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE DATABASE IF NOT EXISTS HeartRating`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`USE HeartRating;`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Users(
		id INT NOT NULL AUTO_INCREMENT,
		username varchar(255) UNIQUE,
		last_update datetime,
		PRIMARY KEY (id));`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Programs(
		id INT NOT NULL AUTO_INCREMENT,
		showname VARCHAR(255),
		title VARCHAR(255),
		PRIMARY KEY (id));`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Sessions(
		id INT NOT NULL AUTO_INCREMENT,
		program_id INT,
		user_id INT,
		heart INT,
		duration INT,
		created_at datetime,
		PRIMARY KEY (id));`)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func db_get_program(db *sql.DB, pid int) (string, string, error) {
	show := ""
	title := ""
	query := `SELECT showname, title FROM HeartRating.Programs WHERE id=?;`
	row, err := db.Query(query, pid)
	if err != nil {
		fmt.Println(err)
		return show, title, err
	}
	defer row.Close()
	if row.Next() {
		err = row.Scan(&show, &title)
		if err != nil {
			fmt.Println(err)
			return show, title, err
		}
	}
	return show, title, err
}

func db_user_exist(db *sql.DB, user string) bool {
	if len(user) == 0 {
		return false
	}

	query := `SELECT id FROM HeartRating.Users WHERE username=?;`
	row, err := db.Query(query, user)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer row.Close()
	if row.Next() {
		return true
	}
	return false
}

func db_get_user_id(db *sql.DB, user string) (int, error) {
	if len(user) == 0 {
		return -1, nil
	}

	var id int
	query := `SELECT id FROM HeartRating.Users WHERE username=?;`
	row, err := db.Query(query, user)
	if err != nil {
		fmt.Println(err)
		return -1, err
	}
	defer row.Close()
	if row.Next() {
		err = row.Scan(&id)
		if err != nil {
			fmt.Println(err)
			return -1, err
		}
	}
	fmt.Println("user id", user, id)
	return id, nil
}

func db_get_users_all(db *sql.DB) ([]User, error) {
	query := `SELECT id, username FROM HeartRating.Users ORDER BY last_update DESC;`
	row, err := db.Query(query)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer row.Close()
	users := make([]User, 0)
	i := 0
	for row.Next() {
		var u string
		var d int
		err = row.Scan(&d, &u)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		users = append(users, User{d, u})
		i += 1
	}
	fmt.Println("users", users)
	return users, nil
}

func db_get_users(db *sql.DB) ([]string, error) {
	query := `SELECT username FROM HeartRating.Users ORDER BY last_update DESC;`
	row, err := db.Query(query)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer row.Close()
	users := make([]string, 0)
	i := 0
	for row.Next() {
		var u string
		err = row.Scan(&u)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		users = append(users, u)
		i += 1
	}
	fmt.Println("users", users)
	return users, nil
}

func db_program_exist(db *sql.DB, show string, title string) bool {
	query := `SELECT id FROM HeartRating.Programs WHERE showname=? and title=?;`
	row, err := db.Query(query, show, title)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer row.Close()
	if row.Next() {
		return true
	}
	return false
}

func db_get_program_id(db *sql.DB, show string, title string) (int, error) {
	id := -1
	query := `SELECT id FROM HeartRating.Programs WHERE showname=? and title=?;`
	row, err := db.Query(query, show, title)
	if err != nil {
		fmt.Println(err)
		return id, err
	}
	defer row.Close()
	if row.Next() {
		err = row.Scan(&id)
		if err != nil {
			fmt.Println(err)
			return id, err
		}
	}
	fmt.Println("program id", id)
	return id, nil
}

func db_get_user_sessions(db *sql.DB, user string) ([]Session, error) {
	id, err := db_get_user_id(db, user)
	if id == -1 || err != nil {
		return nil, err
	}

	query := `SELECT program_id, heart, duration FROM HeartRating.Sessions WHERE user_id=? ORDER BY created_at DESC;`
	row, err := db.Query(query, id)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer row.Close()
	sessions := make([]Session, 0)
	for row.Next() {
		var pid int
		var heart int
		var duration int
		err = row.Scan(&pid, &heart, &duration)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		show, title, err := db_get_program(db, pid)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		s := Session{title, show, heart, duration}
		sessions = append(sessions, s)
	}
	fmt.Println("sessions", sessions)
	return sessions, nil
}

func db_new_user(db *sql.DB, user string) error {
	cmd := `INSERT INTO HeartRating.Users(username, last_update) values(?, NOW());`
	_, err := db.Exec(cmd, user)
	if err != nil {
		return err
	}
	return nil
}

func db_new_program(db *sql.DB, show string, title string) error {
	cmd := `INSERT INTO HeartRating.Programs(showname, title) values(?, ?);`
	_, err := db.Exec(cmd, show, title)
	if err != nil {
		return err
	}
	return nil
}

func db_new_session(db *sql.DB, pid int, uid int, heart int, duration int) error {
	cmd := `INSERT INTO HeartRating.Sessions(program_id, user_id, heart, duration, created_at)
			values(?, ?, ?, ?, NOW());`
	_, err := db.Exec(cmd, pid, uid, heart, duration)
	if err != nil {
		return err
	}
	cmd = `UPDATE HeartRating.Users SET last_update=NOW() WHERE id=?`
	_, err = db.Exec(cmd, uid)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func db_new_data(db *sql.DB) error {
	return nil
}

func launch_web(db *sql.DB) {
	m := martini.Classic()
	m.Use(render.Renderer())
	m.Get("/", func(ren render.Render) {
		pd := make([]userPageData, 0)
		users, _ := db_get_users(db)
		for _, u := range users {
			s, _ := db_get_user_sessions(db, u)
			for _, v := range s {
				pd = append(pd, userPageData{u, v.Title, v.Show, v.Heart, v.Duration})
			}
		}
		dat := indexPageData{pd}
		ren.HTML(200, "index", dat)
	})
	m.Get("/:user", func(params martini.Params, ren render.Render) {
		pd := make([]userPageData, 0)
		user := params["user"]
		ses, _ := db_get_user_sessions(db, user)
		for _, v := range ses {
			pd = append(pd, userPageData{user, v.Title, v.Show, v.Heart, v.Duration})
		}
		dat := indexPageData{pd}
		ren.HTML(200, "user", dat)
	})

	m.Post("/api/save", func(r *http.Request) string {
		var dat PostData
		err := json.NewDecoder(r.Body).Decode(&dat)
		if err != nil {
			log.Println(err)
			return "failure"
		}

		if !db_user_exist(db, dat.User) {
			db_new_user(db, dat.User)
		}
		uid, _ := db_get_user_id(db, dat.User)

		if !db_program_exist(db, dat.Show, dat.Title) {
			db_new_program(db, dat.Show, dat.Title)
		}
		pid, _ := db_get_program_id(db, dat.Show, dat.Title)

		db_new_session(db, pid, uid, dat.Heart, dat.Duration)

		return "success"
	})
	m.Get("/api/sessions/:user", func(params martini.Params) string {
		user := params["user"]
		ses, _ := db_get_user_sessions(db, user)
		u, _ := db_get_user_id(db, user)
		U := User{u, user}
		fmt.Println("U", U)
		fmt.Println("ses", ses)
		resp := ApiSession{U, ses, "success"}
		out, _ := json.MarshalIndent(resp, "", "  ")
		return string(out)
	})
	m.Get("/api/users", func() string {
		users, _ := db_get_users_all(db)
		resp := ApiUsers{users, "success"}
		out, _ := json.MarshalIndent(resp, "", "  ")
		return string(out)
	})
	m.Run()
}

func test_data(db *sql.DB) {
	db_new_user(db, "alice")
	db_new_user(db, "bob")
	db_new_user(db, "carl")
	db_new_user(db, "dan")
	db_new_user(db, "evan")
	db_new_user(db, "george")
	db_new_program(db, "show", "title")

	pid, _ := db_get_program_id(db, "show", "title")
	uid, _ := db_get_user_id(db, "alice")
	uid2, _ := db_get_user_id(db, "bob")
	uid3, _ := db_get_user_id(db, "carl")

	db_new_session(db, pid, uid, 1, 1)
	db_new_session(db, pid, uid2, 1, 1)
	db_new_session(db, pid, uid3, 1, 1)
	db_new_session(db, pid, uid2, 1, 1)
	db_new_session(db, pid, uid3, 1, 1)
	db_new_session(db, pid, uid, 1, 1)
	db_new_session(db, pid, uid, 1, 1)
}

func main() {
	db := setup_db()
	defer db.Close()

	test_data(db)
	launch_web(db)
}
