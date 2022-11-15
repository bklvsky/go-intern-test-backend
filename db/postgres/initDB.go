package postgres

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type dbConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBname   string
}

func InitDB() *sql.DB{

	var dbClient *sql.DB
	cf := dbConfig{
		os.Getenv("POSTGRES_HOST"),
		// "127.0.0.1",
		os.Getenv("POSTGRES_PORT"),
		// "5432",
		os.Getenv("POSTGRES_USER"),
		// "avito_balance",
		os.Getenv("POSTGRES_PASSWORD"),
		// "avito_balance",
		os.Getenv("POSTGRES_NAME")}
		// "avito_balance"}


	url := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
	cf.User, cf.Password, cf.Host, cf.Port, cf.DBname)
	fmt.Println("Trying to connect to", url)
	var err error
	dbClient, err = sql.Open("postgres", url)
	if err != nil {
		fmt.Println("Connection to the database failed")
		panic(err)
	}

	err = dbClient.Ping()
	if err != nil {
		fmt.Println("Connection to the database failed")
		panic(err)
	}

	fmt.Println("Successfuly connected to the database!")
	return dbClient
}
