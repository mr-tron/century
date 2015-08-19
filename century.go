package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"net/http"
	"strings"
)

var session *gocql.Session

func password_hash(password string) string {
	hash := sha512.Sum512([]byte(password))
	return hex.EncodeToString(hash[:])
}

func user(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method == "POST" {
		r.ParseForm()
		login := r.Form["login"][0]
		password := password_hash((r.Form["password"][0]))
		var existed_login string
		if err := session.Query("SELECT login FROM users WHERE login = ?", login).Consistency(gocql.One).Scan(&existed_login); err != nil {
			fmt.Printf("", err)
		}
		if existed_login != "" {
			fmt.Fprintf(w, "{\"error:\":\"User already exist.\"}\n")
			return
		}
		if err := session.Query("INSERT INTO users (login, password) VALUES (?, ?) IF NOT EXISTS",
			login, password).Exec(); err != nil {
			fmt.Println(err)
		}
		fmt.Fprintf(w, "{\"error\":\"\"}\n")
	} else {
		w.WriteHeader(405)
		fmt.Fprintf(w, "{\"error\":\"Allowed methods: POST\"}\n")
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method == "POST" {
		r.ParseForm()
		login := r.Form["login"][0]
		password := password_hash((r.Form["password"][0]))
		var existed_login string
		var existed_cookie gocql.UUID
		var cookie gocql.UUID
		if err := session.Query("SELECT login FROM users WHERE login = ? and password = ?", login, password).Consistency(gocql.One).Scan(&existed_login); err != nil {
			fmt.Println(err)
		}
		if existed_login != "" {
			if err := session.Query("SELECT cookie FROM cookies WHERE login = ?", login).Consistency(gocql.One).Scan(&existed_cookie); err != nil {
				if err == gocql.ErrNotFound {
					cookie, _ = gocql.RandomUUID()
					if err := session.Query("INSERT INTO cookies (login, cookie) VALUES (?, ?)", login, cookie).Exec(); err != nil {
						w.WriteHeader(500)
						return
					}
				} else {
					w.WriteHeader(500)
					return
				}
			} else {
				cookie = existed_cookie
			}
			w.Header().Set("Set-Cookie", "sessionToken="+cookie.String())
			fmt.Fprintf(w, "{\"error\":\"\"}\n")
		} else {
			w.WriteHeader(403)
			fmt.Fprintf(w, "{\"error\":\"User or password is wrong\"}\n")
			return
		}
	} else {
		w.WriteHeader(405)
		fmt.Fprintf(w, "{\"error\":\"Allowed methods: POST\"}\n")
	}
}

func login_check(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var login string
	if r.Method == "GET" {
		cookie, err := r.Cookie("sessionToken")
		if err == nil {
			cookie_uuid, parse_error := gocql.ParseUUID(strings.TrimPrefix(cookie.String(), "sessionToken="))
			if parse_error != nil {
				w.WriteHeader(403)
				fmt.Fprintf(w, "{\"error\":\"unauthorized\"}\n")
				return
			}
			if request_err := session.Query("select login from cookies where cookie = ?", cookie_uuid).Consistency(gocql.One).Scan(&login); request_err != nil {
				if request_err == gocql.ErrNotFound {
					w.WriteHeader(403)
					fmt.Fprintf(w, "{\"error\":\"unauthorized\"}\n")
					return
				} else {
					w.WriteHeader(500)
					return
				}
			}
			fmt.Fprintf(w, "{\"status\":\"You logged in!\"}\n")
		} else {
			w.WriteHeader(403)
			fmt.Fprintf(w, "{\"error\":\"unauthorized\"}\n")
			return
		}
	} else {
		w.WriteHeader(405)
		fmt.Fprintf(w, "{\"error\":\"Allowed methods: GET\"}\n")
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	cookie, err := r.Cookie("sessionToken")
	if r.Method == "POST" {
		if err == nil {
			cookie_uuid, parse_error := gocql.ParseUUID(strings.TrimPrefix(cookie.String(), "sessionToken="))
			if parse_error != nil {
				w.WriteHeader(400)
				fmt.Fprintf(w, "{\"error\":\"bad cookie\"}\n")
				return
			}
			if request_err := session.Query("delete from cookies where cookie = ?", cookie_uuid).Exec(); request_err != nil {
				fmt.Println(request_err)
				w.WriteHeader(500)
				return
			}
		}
		fmt.Fprintf(w, "{\"error\":\"\"}\n")
	} else {
		w.WriteHeader(405)
		fmt.Fprintf(w, "{\"error\":\"Allowed methods: GET\"}\n")
	}
}

func main() {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "century"
	cluster.Consistency = gocql.Quorum
	session, _ = cluster.CreateSession()
	session.Query("CREATE TABLE IF NOT EXIST users (login text PRIMARY KEY, password text)").Exec()
	session.Query("CREATE TABLE IF NOT EXIST cookies (login text, cookie UUID PRIMARY KEY)").Exec()
	session.Query("CREATE INDEX IF NOT EXIST ON cookies(login)").Exec()
	session.Query("CREATE INDEX IF NOT EXIST ON users(password)").Exec()
	http.HandleFunc("/user", user)
	http.HandleFunc("/login", login)
	http.HandleFunc("/login/check", login_check)
	http.HandleFunc("/logout", logout)

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
