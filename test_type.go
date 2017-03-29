package gopsql

import (
	"errors"
	"fmt"
)

type Book struct {
	ID     int    `column_type:"SERIAL primary key" column_skip_insert:"yes"`
	Title  string `column_type:"varchar(255)"`
	Author string `column_type:"varchar(255)"`
	Pages  []Page `column_skip:"yes"`
}

// Implement Saveable interface
func (b Book) GetID() int {
	return b.ID
}

// Implement Saveable interface
func (b Book) SetID(id int) {
	b.ID = id
}

func (b *Book) Save() error {
	if b.ID > 0 { //Exists, update
		return UpdateExisting(*b)
	}
	// New
	return InsertNew(*b).Scan(&b.ID)
}

func (b *Book) Find() error {
	rows, err := GetFiltered(*b, "ID", fmt.Sprintf("%v", b.ID))
	if err != nil {
		return err
	}

	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(&b.ID, &b.Title, &b.Author)
		if err != nil {
			return err
		}
		b.Pages, err = GetAllPagesForBook(b.ID)
		return err
	}
	return errors.New("No matching book found")
}

func (b *Book) Delete() error {
	return Delete(*b)
}

type Page struct {
	ID      int    `column_type:"SERIAL primary key" column_skip_insert:"yes" column_order_by:"yes"`
	BookID  int    `column_type:"SERIAL references Book(ID)"`
	Content string `column_type:"TEXT"`
}

// Implement Saveable
func (p Page) GetID() int {
	return p.ID
}

// Implement Saveable interface
func (p Page) SetID(id int) {
	p.ID = id
}

func (p *Page) Save() error {
	if p.ID > 0 { // Exists, update
		return UpdateExisting(*p)
	}
	// New
	return InsertNew(*p).Scan(&p.ID)
}

func GetAllPagesForBook(bookID int) ([]Page, error) {
	var res []Page
	rows, err := GetFiltered(Page{}, "BookID", fmt.Sprintf("%d", bookID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		page := Page{}
		err := rows.Scan(&page.ID, &page.BookID, &page.Content)

		if err != nil {
			return nil, err
		}

		res = append(res, page)
	}
	return res, nil
}
