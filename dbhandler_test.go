package gopsql

import (
	"os"
	"testing"

	"github.com/Tebro/logger"
)

//Setup
func TestMain(m *testing.M) {
	logger.SetDebug()

	dbHost := "localhost"

	if v, ok := os.LookupEnv("POSTGRES_HOST"); ok {
		dbHost = v
	}

	dbUser := "postgres"
	dbPass := "postgres"
	dbName := "postgres"
	sslMode := "disable"

	models := []Saveable{
		Book{},
		Page{},
	}

	err := Setup(dbHost, dbUser, dbPass, dbName, sslMode, models)

	if err != nil {
		logger.Error(err)
		logger.Fatal("Failed to connect to database.")
	}

	os.Exit(m.Run())
}

const bookTitle = "Foobar"
const bookAuthor = "Mr. Foo Bar"

func getTestBook() Book {
	b := Book{}
	b.Title = bookTitle
	b.Author = bookAuthor
	return b
}

const pageContent = "Hello World!"

func getTestPage() Page {
	p := Page{}
	p.Content = pageContent
	return p
}

func getTestPageForBook(b Book) Page {
	p := getTestPage()
	p.BookID = b.GetID()
	return p
}

func TestInsertion(t *testing.T) {
	b := getTestBook()

	err := b.Save()
	if err != nil {
		t.Error(err)
	}
}

func TestInsertionAndRetrieval(t *testing.T) {
	b := getTestBook()

	err := b.Save()
	if err != nil {
		t.Error(err)
	}

	b2 := Book{}
	b2.ID = b.GetID()
	err = b2.Find()
	if err != nil {
		t.Error(err)
	}

	if b2.Author != bookAuthor {
		t.Errorf("Author does not match, got: %s", b2.Author)
	}

	if b2.Title != bookTitle {
		t.Errorf("Title does not match, got: %s", b2.Title)
	}
}

func TestUpdating(t *testing.T) {
	const updatedTitle = "BamBazBat"

	b := getTestBook()

	err := b.Save()
	if err != nil {
		t.Error(err)
	}

	b.Title = updatedTitle
	err = b.Save()
	if err != nil {
		t.Error(err)
	}

	b2 := Book{}
	b2.ID = b.GetID()
	err = b2.Find()
	if err != nil {
		t.Error(err)
	}
	if b2.Title != updatedTitle {
		t.Errorf("Title does not match, got: %s", b2.Title)
	}
}

func TestDeleting(t *testing.T) {
	b := getTestBook()

	err := b.Save()
	if err != nil {
		t.Error(err)
	}

	err = b.Delete()
	if err != nil {
		t.Error(err)
	}
}
