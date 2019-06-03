package main

import (
	"strconv"
	"time"

	"github.com/gin-gonic/contrib/sessions"
	ca "github.com/patrickmn/go-cache"
)

// User model
type User struct {
	ID        int
	Name      string
	Email     string
	Password  string
	LastLogin string
}

func authenticate(email string, password string) (User, bool) {
	var u User
	err := db.QueryRow("SELECT * FROM users WHERE email = ? LIMIT 1", email).Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.LastLogin)
	if err != nil {
		return u, false
	}
	result := password == u.Password
	return u, result
}

func notAuthenticated(session sessions.Session) bool {
	uid := session.Get("uid")
	return !(uid.(int) > 0)
}

func getUser(uid int) User {
	u := User{}
	r := db.QueryRow("SELECT * FROM users WHERE id = ? LIMIT 1", uid)
	err := r.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.LastLogin)
	if err != nil {
		return u
	}

	return u
}

func currentUserIDOnly(session sessions.Session) User {
	u := User{}

	id := session.Get("uid")
	name := session.Get("name")

	if id != nil && name != nil {
		u.ID = id.(int)
		u.Name = name.(string)
	}

	return u
}

func currentUser(session sessions.Session) User {
	uid := session.Get("uid")
	u := User{}
	r := db.QueryRow("SELECT * FROM users WHERE id = ? LIMIT 1", uid)
	err := r.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.LastLogin)
	if err != nil {
		return u
	}

	return u
}

// BuyingHistory : products which user had bought
func (u *User) BuyingHistory() []Product {

	data, found := cache.Get("buying_history_" + strconv.Itoa(u.ID))

	if found {
		return data.([]Product)
	}

	rows, err := db.Query(
		"SELECT p.id, p.name, p.description, p.image_path, p.price, h.created_at "+
			"FROM histories as h "+
			"LEFT OUTER JOIN products as p "+
			"ON h.product_id = p.id "+
			"WHERE h.user_id = ? "+
			"ORDER BY h.id DESC", u.ID)
	if err != nil {
		return nil
	}

	var products []Product

	defer rows.Close()
	for rows.Next() {
		p := Product{}
		var cAt string
		fmt := "2006-01-02 15:04:05"
		err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.ImagePath, &p.Price, &cAt)
		tmp, _ := time.Parse(fmt, cAt)
		p.CreatedAt = (tmp.Add(9 * time.Hour)).Format(fmt)
		if err != nil {
			panic(err.Error())
		}
		products = append(products, p)
	}

	cache.Set("buying_history_"+strconv.Itoa(u.ID), products, ca.DefaultExpiration)

	return products
}

// BuyProduct : buy product
func (u *User) BuyProduct(pid string) {
	cache.Delete("buying_history_" + strconv.Itoa(u.ID))

	db.Exec(
		"INSERT INTO histories (product_id, user_id, created_at) VALUES (?, ?, ?)",
		pid, u.ID, time.Now())
}

// CreateComment : create comment to the product
func (u *User) CreateComment(pid string, content string) {
	db.Exec(
		"INSERT INTO comments (product_id, user_id, content, created_at) VALUES (?, ?, ?, ?)",
		pid, u.ID, content, time.Now())

	var cnt int
	data, found := cache.Get("product_count_" + pid)
	if !found {
		cnterr := db.QueryRow("SELECT count(1) as count FROM comments WHERE product_id = ?", pid).Scan(&cnt)
		if cnterr != nil {
			cnt = 0
		}
	} else {
		cnt = data.(int) + 1
	}

	cache.Set("product_count_"+pid, cnt, ca.DefaultExpiration)
	cache.Delete("cWriters_" + pid)
}
