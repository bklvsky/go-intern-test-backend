package handlers

import (
	"avito-user-balance/models"
	"avito-user-balance/repositories/postgres"
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	// "fmt"
	"database/sql"
	"log"
	"net/http"

	"encoding/json"

	"github.com/gorilla/mux"
)

type AppHandler struct {
	l  *log.Logger
	hu *UserHandler
	tr *postgres.TransactionsRepository
}

func NewAppHandler(lg *log.Logger, hu *UserHandler, d *sql.DB) *AppHandler {
	return &(AppHandler{lg, hu,
		postgres.NewTransactionsRepository(d)})
}

func (ha *AppHandler) PostTransaction(rw http.ResponseWriter, rq *http.Request) {
	ha.l.Println("Handle POST TRANSACTION")
	ha.l.Println("POST", rq.URL.Path)

	newTr := rq.Context().Value(KeyTransactionPost{}).(*models.Transaction)
	usrUpd := &models.User{newTr.UserId, newTr.Value, newTr.ReserveValue}
	
	err := ha.hu.updateUserData(usrUpd)
	if err == nil {
		err = ha.tr.AddTransaction(newTr)
	}
	if err != nil {
		ha.l.Println("Error:", err.Error())
		rw.WriteHeader(http.StatusInternalServerError)
		enc := json.NewEncoder(rw)
		enc.Encode(PostResponseJSON{false})
		return
	}
	enc := json.NewEncoder(rw)
	enc.Encode(PostResponseJSON{true})
}

func (ha *AppHandler) GetTransactions(rw http.ResponseWriter, rq *http.Request) {
	ha.l.Println("GET/", rq.URL.Path)
	// vars := mux.Vars(rq)

	// if _, ok := vars["id"]; ok {
	// 	ha.l.Println("Looking for models.USER IN DB")
	// 	ha.getTransaction(rw, rq)
	// } else {
	trs, err := ha.tr.FindAllTransactions()
	if err != nil {
		ha.l.Println("Database error:", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	err = transactionsToJSON(trs, rw)

	if err != nil {
		ha.l.Println("Error encoding json", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	// }
}

func (ha *AppHandler) GetTransaction(rw http.ResponseWriter, rq *http.Request) {
	vars := mux.Vars(rq)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		ha.l.Println("URL error, invalid models.userID")
		http.Error(rw, "Invalid Request", http.StatusBadRequest)
		return
	}

	tr, error := ha.tr.FindTransactionByID(id)
	if tr == nil {
		ha.l.Println(error.Error())
		http.Error(rw, error.Error(), http.StatusNotFound)
		return
	}

	err = transactionToJSON(*tr, rw)
	if err != nil {
		ha.l.Println("Error encoding json", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func transactionFromJson(tr *models.Transaction, rd io.Reader) error {
	decoder := json.NewDecoder(rd)
	return decoder.Decode(tr)
}

func transactionsToJSON(trs models.Transactions, wr io.Writer) error {
	encoder := json.NewEncoder(wr)
	return encoder.Encode(trs)
}

func transactionToJSON(tr models.Transaction, wr http.ResponseWriter) error {
	encoder := json.NewEncoder(wr)
	return encoder.Encode(tr)
}

// validate user prerequisites
func (ha *AppHandler) ValidateUser(tr *models.Transaction, err *error) {
	if *err != nil {
		return
	}
	ha.hu.ValidateUserID(tr.UserId, err)
	userDB := ha.hu.ValidateUserInDb(tr.UserId, err)
	if *err != nil {
		return
	}
	// validate User's balance and reserve values for the transaction
	userUpd := &models.User {userDB.ID, tr.Value, tr.ReserveValue}
	fmt.Println("USER UPDATE =", userUpd)
	ha.hu.ValidateUserData(userUpd, userDB, err)
}

type KeyTransactionPost struct{}

func (ha *AppHandler) MiddlewareAdditional(next http.Handler) http.Handler {

	ha.l.Println("In ADDITIONAL MIDDLEWARE")
	return ha.MiddlewareValidateNewTransaction(next)
}

func (ha *AppHandler) MiddlewareValidateNewTransaction(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		ha.l.Println("In TRANSACTION MIDDLEWARE")
		var tr = &models.Transaction{}
		err := transactionFromJson(tr, rq.Body)

		if err != nil {
			ha.l.Println("Error:", err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
			enc := json.NewEncoder(rw)
			enc.Encode(PostResponseJSON{false})
			return
		}
		tr.Timesp = time.Now()

		// How transaction changes User's balance
		tr.ReserveValue = tr.Value
		tr.Value *= -1

		ha.ValidateUser(tr, &err)
		if err != nil {
			ha.l.Println("Error:", err.Error())
			rw.WriteHeader(http.StatusBadRequest)
			enc := json.NewEncoder(rw)
			enc.Encode(PostResponseJSON{false})
			return
		}
		fmt.Println("USER VALIDATED")

		ctx := context.WithValue(rq.Context(), KeyTransactionPost{}, tr)
		rq = rq.WithContext(ctx)

		next.ServeHTTP(rw, rq)
	})
}
