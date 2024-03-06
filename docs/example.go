package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/maragudk/goqite"
)

const schema = `
create table goqite (
  id text primary key default ('m_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  queue text not null,
  body blob not null,
  timeout text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  received integer not null default 0
) strict;

create trigger goqite_updated_timestamp after update on goqite begin
  update goqite set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = old.id;
end;
`

func main() {
	db, err := sql.Open("sqlite3", ":memory:?_journal=WAL&_timeout=5000&_fk=true")
	if err != nil {
		log.Fatalln(err)
	}
	if _, err := db.Exec(schema); err != nil {
		log.Fatalln(err)
	}

	q := goqite.New(goqite.NewOpts{
		DB:   db,
		Name: "jobs",
	})

	err = q.Send(context.Background(), goqite.Message{
		Body: []byte("yo"),
	})
	if err != nil {
		log.Fatalln(err)
	}

	m, err := q.Receive(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(string(m.Body))

	if err := q.Extend(context.Background(), m.ID, time.Second); err != nil {
		log.Fatalln(err)
	}

	if err := q.Delete(context.Background(), m.ID); err != nil {
		log.Fatalln(err)
	}
}