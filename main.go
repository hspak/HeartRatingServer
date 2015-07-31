package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)
import _ "github.com/go-sql-driver/mysql"
import "github.com/go-martini/martini"
import "github.com/martini-contrib/render"

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

type session struct {
	title    string
	show     string
	heart    int
	duration int
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

	_, err = db.Exec(`DROP DATABASE HeartRating`)
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
	if row.Next() {
		err = row.Scan(&show, &title)
		if err != nil {
			fmt.Println(err)
			return show, title, err
		}
	}
	return show, title, err
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

func db_get_users(db *sql.DB) ([]string, error) {
	query := `SELECT username FROM HeartRating.Users ORDER BY last_update DESC LIMIT 20;`
	row, err := db.Query(query)
	users := make([]string, 10)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	i := 0
	for row.Next() {
		err = row.Scan(&(users[i]))
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		i += 1
	}
	fmt.Println("users", users)
	return users, nil
}

func db_get_program_id(db *sql.DB, show string, title string) (int, error) {
	id := -1
	query := `SELECT id FROM HeartRating.Programs WHERE showname=? and title=?;`
	row, err := db.Query(query, show, title)
	if err != nil {
		fmt.Println(err)
		return id, err
	}
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

func db_get_user_sessions(db *sql.DB, user string) ([]session, error) {
	id, err := db_get_user_id(db, user)
	if id == -1 || err != nil {
		return nil, err
	}

	query := `SELECT program_id, heart, duration FROM HeartRating.Sessions WHERE user_id=? ORDER BY created_at DESC;`
	row, err := db.Query(query, id)
	sessions := make([]session, 0)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
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
		s := session{title, show, heart, duration}
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
				pd = append(pd, userPageData{u, v.title, v.show, v.heart, v.duration})
			}
		}
		dat := indexPageData{pd}
		ren.HTML(200, "index", dat)
	})
	m.Get("/:user", func(params martini.Params, ren render.Render) {
		// user := params["user"]
		ren.HTML(200, "user", nil)
	})
	m.Post("/api/save", func(r *http.Request) string {
		body, _ := ioutil.ReadAll(r.Body)
		fmt.Println(string(body))
		return "success"
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
