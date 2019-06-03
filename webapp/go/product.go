package main

import (
	ca "github.com/patrickmn/go-cache"
	"strconv"
)

// Product Model
type Product struct {
	ID          int
	Name        string
	Description string
	ImagePath   string
	Price       int
	CreatedAt   string
}

// ProductWithComments Model
type ProductWithComments struct {
	ID           int
	Name         string
	Description  string
	ImagePath    string
	Price        int
	CreatedAt    string
	CommentCount int
	Comments     []CommentWriter
}

// CommentWriter Model
type CommentWriter struct {
	Content string
	Writer  string
}

func getProduct(pid int) Product {
	data, found := cache.Get("product_" + strconv.Itoa(pid))

	if found {
		return data.(Product)
	}

	p := Product{}
	row := db.QueryRow("SELECT * FROM products WHERE id = ? LIMIT 1", pid)
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.ImagePath, &p.Price, &p.CreatedAt)
	if err != nil {
		panic(err.Error())
	}

	cache.Set("product_"+strconv.Itoa(pid), p, ca.DefaultExpiration)

	return p
}

func getProductsWithCommentsAt(page int) []ProductWithComments {
	// select 50 products with offset page*50
	products := []ProductWithComments{}
	rows, err := db.Query("SELECT * FROM products ORDER BY id DESC LIMIT 50 OFFSET ?", page*50)
	if err != nil {
		return nil
	}

	defer rows.Close()
	for rows.Next() {
		p := ProductWithComments{}
		err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.ImagePath, &p.Price, &p.CreatedAt)

		var cnt int
		cnt = cachedProductCommentCount(p.ID)
		p.CommentCount = cnt

		if cnt > 0 {
			p.Comments = cachedWriters(p.ID)
		}

		products = append(products, p)
	}

	return products
}

func cachedWriters(pid int) []CommentWriter {
	data, found := cache.Get("cWriters_" + strconv.Itoa(pid))

	var cWriters []CommentWriter

	if !found {
		subrows, suberr := db.Query("SELECT content ,name FROM comments as c INNER JOIN users as u "+
			"ON c.user_id = u.id WHERE c.product_id = ? ORDER BY c.created_at DESC LIMIT 5", pid)
		if suberr != nil {
			subrows = nil
		}

		defer subrows.Close()
		for subrows.Next() {
			var cw CommentWriter
			subrows.Scan(&cw.Content, &cw.Writer)
			cWriters = append(cWriters, cw)
		}
	} else {
		cWriters = data.([]CommentWriter)
	}

	cache.Set("cWriters_"+strconv.Itoa(pid), cWriters, ca.DefaultExpiration)

	return cWriters
}

func cachedProductCommentCount(id int) int {
	var cnt int
	data, found := cache.Get("product_count_" + strconv.Itoa(id))

	if !found {
		cnterr := db.QueryRow("SELECT count(1) as count FROM comments WHERE product_id = ?", id).Scan(&cnt)
		if cnterr != nil {
			cnt = 0
		}
	} else {
		cnt = data.(int)
	}

	cache.Set("product_count_"+strconv.Itoa(id), cnt, ca.DefaultExpiration)

	return cnt
}

func (p *Product) isBought(uid int) bool {
	var count int

	err := db.QueryRow(
		"SELECT count(1) as count FROM histories WHERE product_id = ? AND user_id = ?",
		p.ID, uid,
	).Scan(&count)
	if err != nil {
		panic(err.Error())
	}

	return count > 0
}
