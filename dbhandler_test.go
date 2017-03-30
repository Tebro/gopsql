package gopsql

import (
	"os"
	"reflect"
	"testing"

	"github.com/Tebro/logger"
	_ "github.com/lib/pq"
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

func Test_parseFilterString(t *testing.T) {
	type args struct {
		filters []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   []string
		wantErr bool
	}{
		{
			name:    "TestWithNoArgs",
			wantErr: true,
		},
		{
			name:    "TestWithOneArg",
			wantErr: true,
			args: args{
				filters: []string{
					"Foo",
				},
			},
		},
		{
			name:    "TestWithTwoArgs",
			wantErr: false,
			want:    "key=$1",
			want1: []string{
				"value",
			},
			args: args{
				filters: []string{
					"key",
					"value",
				},
			},
		},
		{
			name:    "TestWithTwoPairs",
			wantErr: false,
			want:    "key1=$1 AND key2=$2",
			want1: []string{
				"value1",
				"value2",
			},
			args: args{
				filters: []string{
					"key1",
					"value1",
					"AND",
					"key2",
					"value2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseFilterString(tt.args.filters...)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFilterString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseFilterString() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseFilterString() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getSelectAllQuery(t *testing.T) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name:  "StructWithoutOrderByField",
			want:  "SELECT ID,Title,Author FROM Book;",
			want1: false,
			args: args{
				obj: Book{},
			},
		},
		{
			name:  "StructWithOrderByField",
			want:  "SELECT ID,BookID,Content FROM Page ORDER BY ID;",
			want1: true,
			args: args{
				obj: Page{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getSelectAllQuery(tt.args.obj)
			if got != tt.want {
				t.Errorf("getSelectAllQuery() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getSelectAllQuery() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
