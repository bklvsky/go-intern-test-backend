package main

import (
	"avito-user-balance/handlers"
	"avito-user-balance/db/postgres"
	"github.com/joho/godotenv"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// // Setting up Router with Gorilla framework
// type App struct {
// 	MX *muxRouter
// 	DB *sql.DB
// }

// func (app App) Initialize() {
// 	if err := godotenv.Load("./.env"); err != nil {
// 		log.Fatalf("error loading env variables :%s", err.Error())
// 	}

// 	// for debug
// 	fmt.Println("POSTGRES_DB", os.Getenv("POSTGRES_DB"))
// 	fmt.Println("POSTGRES_PORT", os.Getenv("POSTGRES_PORT"))
// 	fmt.Println("POSTGRES_USERNAME", os.Getenv("POSTGRES_USERNAME"))
// 	fmt.Println("POSTGRES_PASSWORD", os.Getenv("POSTGRES_PASSWORD"))
// 	fmt.Println("POSTGRES_HOST", os.Getenv("POSTGRES_HOST"))
// 	//for debug

// 	handlerUser := handlers.NewUserHandler(l)
// 	app.MX = setupRouter(handlerUser)
// }

func setupRouter(handlerUser *handlers.UserHandler) *mux.Router {
	mx := mux.NewRouter()

	getRouter := mx.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/users{_dummy:/?$}", handlerUser.GetUsers)
	// later
	// getRouter.Use(handlerUser.MiddlewareValidateUser)
	getRouter.HandleFunc("/users/{id:[0-9]+}", handlerUser.GetUsers)

	postRouter := mx.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/users{_dummy:/?$}", handlerUser.PostUsers)
	postRouter.Use(handlerUser.MiddlewareValidateUser)

	return mx
}

// Setting up http Server

func setupServer(mx *mux.Router) http.Server {
	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		addr = ":8080"
	}

	server := http.Server{
		Addr:        addr,
		Handler:     mx,
		ReadTimeout: time.Second * 5,
		IdleTimeout: time.Second * 30,
	}
	return server
}

func main() {
	if err := godotenv.Load("./.env"); err != nil {
		log.Fatalf("error loading env variables :%s", err.Error())
	}

	//for debug
	fmt.Println("POSTGRES_NAME", os.Getenv("POSTGRES_NAME"))
	fmt.Println("POSTGRES_PORT", os.Getenv("POSTGRES_PORT"))
	fmt.Println("POSTGRES_USER", os.Getenv("POSTGRES_USER"))
	fmt.Println("POSTGRES_PASSWORD", os.Getenv("POSTGRES_PASSWORD"))
	fmt.Println("POSTGRES_HOST", os.Getenv("POSTGRES_HOST"))
	//for debug
	
	db := postgres.InitDB()
	l := log.New(os.Stdout, "", log.LstdFlags)

	handlerUser := handlers.NewUserHandler(l, db)

	mx := setupRouter(handlerUser)
	server := setupServer(mx)

	go func() {
		fmt.Println("Server Is Runnning at", server.Addr)
		err := server.ListenAndServe()

		if err != nil {
			log.Println("Error starting server", err)
			os.Exit(1)
		}
	}()

	const (
		host     = "localhost" // replace to IP address/domain
		port     = 5432
		user     = "postgres" // replace with username
		password = "password" // replace with password
		dbname   = "postgres" // replace with the database name
	)

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, os.Kill)

	sig := <-ch
	log.Println("SERVER GOT SIGNAL:", sig)

	cntx := context.Background()
	server.Shutdown(cntx)
	log.Println("SERVER STOPPED")
}
