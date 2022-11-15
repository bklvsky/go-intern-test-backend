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

func setupRouter() *mux.Router {
	db := postgres.InitDB()
	l := log.New(os.Stdout, "", log.LstdFlags)

	mx := mux.NewRouter()
	
	handlerUser := handlers.NewUserHandler(l, db)
	setupRouterUsers(handlerUser, mx)

	handlerApp := handlers.NewAppHandler(l, handlerUser, db)
	setupRouterApp(handlerApp, mx)
	return mx
}

func setupRouterApp(handlerApp *handlers.AppHandler, mx *mux.Router) {
	getTrRouter := mx.Methods(http.MethodGet).Subrouter()
	getTrRouter.HandleFunc("/orders/{id:[0-9]+}", handlerApp.GetTransaction)
	getTrRouter.HandleFunc("/orders{_dummy:/?$}", handlerApp.GetTransactions)

	
	postTrRouter := mx.Methods(http.MethodPost).Subrouter()
	postTrRouter.HandleFunc("/orders{_dummy:/?$}", handlerApp.PostTransaction)
	postTrRouter.Use(handlerApp.MiddlewareAdditional)
}

func setupRouterUsers(handlerUser *handlers.UserHandler, mx *mux.Router) {

	getUserRouter := mx.Methods(http.MethodGet).Subrouter()
	getUserRouter.HandleFunc("/users{_dummy:/?$}", handlerUser.GetUsers)
	// later
	// getRouter.Use(handlerUser.MiddlewareValidateUser)
	getUserRouter.HandleFunc("/users/{id:[0-9]+}", handlerUser.GetUser)

	postUserRouter := mx.Methods(http.MethodPost).Subrouter()
	postUserRouter.HandleFunc("/users{_dummy:/?$}", handlerUser.PostUsers)
	postUserRouter.Use(handlerUser.MiddlewareValidateUser)
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
	
	mx := setupRouter()
	server := setupServer(mx)

	go func() {
		fmt.Println("Server Is Runnning at", server.Addr)
		err := server.ListenAndServe()

		if err != nil {
			log.Println("Error starting server", err)
			os.Exit(1)
		}
	}()

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, os.Kill)

	sig := <-ch
	log.Println("SERVER GOT SIGNAL:", sig)

	cntx := context.Background()
	server.Shutdown(cntx)
	log.Println("SERVER STOPPED")
}
