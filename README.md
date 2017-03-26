# Gopsql

This is a library that helps in serializing structs into database queries.

Currently only supports PostgreSQL.

## Usage example

models.go

    package main

    import (
        "fmt"
        "github.com/Tebro/gopsql"
        "errors"
    )

    type Book struct {
        ID        int       `column_type:"SERIAL primary key" column_skip_insert:"yes"`
        Title     string    `column_type:"varchar(255)"`
        Author    string    `column_type:"varchar(255)"`
        Pages     []Page    `column_skip:"yes"`
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
            return gopsql.UpdateExisting(*b)
        }
        // New
        return gopsql.InsertNew(*c).Scan(&b.ID)
    }

    func (b *Book) Find() error {
        rows, err := gopsql.GetFiltered(*b, "ID", string(b.ID))
        if err != nil {
            return err
        }

        defer rows.Close()
        if rows.Next(){
            err := rows.Scan(&b.ID, &b.Title, &b.Author)
            if err != nil {
                return err
            }
            b.Pages, err := (Page{}).GetAllForBook(b.ID)
            return err
        }
        return errors.New("No matching book found")
    }


    type Page struct {
        ID         int      `column_type:"SERIAL primary key" column_skip_insert:"yes" column_order_by:"yes"`
        BookID     int      `column_type:"SERIAL references Book(ID)"`
        Content    string   `column_type:"TEXT"`
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
            return gopsql.UpdateExisting(*p)
        }
        // New
        return gopsql.InsertNew(*p).Scan(&p.ID)
    }

    func (p *Page) GetAllForBook(bookID int) ([]Page, error) {
        var res []Page
        rows, err := gopsql.GetFiltered(p, "BookID", fmt.Sprintf("%d", bookID))
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


main.go

    package main

    import "github.com/Tebro/gopsql"

    func main() {
        dbHost := "localhost"
        dbUser := "root"
        dbPass := "something"
        dbName := "demo"
        sslMode := "disable" // Check: https://godoc.org/github.com/lib/pq

        // Select which models to load into the DB on startup
        models := []gopsql.Saveable{
            Book{},
            Page{},
        }

        gopsql.Setup(dbHost, dbUser, dbPass, dbName, sslMode, models)

        // Let's create a book with one page

        b := Book{}
        b.Title = "Foobar"
        b.Author = "Mr. Foo Bar"

        // Save it to DB
        err := b.Save()

        // Not going to check the error here, but you should

        // Let's add a page to the book
        p := Page{}
        p.BookID = b.ID // The book ID is now there
        p.Content = "Hello World!"
        err := p.Save()

        // Remember the to check the error here as well

        // Let's retrieve it from the DB
        b2 := Book{}
        b2.ID = b.ID
        err := b2.Find()

        // Again, check the error!

        // At this point b2 will have all the fields.
    }


For more information check the [reference](doc/ref.md)
