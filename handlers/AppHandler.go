package handlers

import (
	"avito-user-balance/models"
	"avito-user-balance/repositories/postgres"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

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

func (ha *AppHandler) PostTransfer(rw http.ResponseWriter, rq *http.Request) {
	tf := rq.Context().Value(KeyTransfer{}).(*models.Transfer)
	senderChanges := &models.User{tf.Sender, -tf.Value, 0}
	recipientChanges := &models.User{tf.Recipient, tf.Value, 0}

	err := ha.hu.updateUserData(senderChanges)
	if err == nil {
		err = ha.hu.updateUserData(recipientChanges)
		if err != nil {
			senderChanges.Balance = recipientChanges.Balance
			err = ha.hu.updateUserData(senderChanges)
		}
	}
	if err != nil {
		ha.l.Println("Error:", err.Error())
		rw.WriteHeader(http.StatusInternalServerError)
		enc := json.NewEncoder(rw)
		enc.Encode(PostResponseJSON{false})
		return
	}
	senderTr, recipTr := transactionsFromTransfer(tf)
	err = ha.tr.AddTransaction(senderTr)
	err = ha.tr.AddTransaction(recipTr)
	enc := json.NewEncoder(rw)
	enc.Encode(PostResponseJSON{true})
}

func transactionsFromTransfer(tf *models.Transfer) (*models.Transaction, *models.Transaction) {
	senderTr := &models.Transaction{0, 0, tf.Sender, 0,
		-tf.Value, 0, time.Now(), "approved", "Transfer to another user"}
	recTr := &models.Transaction{0, 0, tf.Recipient, 0,
		tf.Value, 0, time.Now(), "approved", "Transfer from another user"}
	return senderTr, recTr
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

	tr, error := ha.tr.FindLastTransactionByOrder(id)
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

func transferFromJSON(tf *models.Transfer, rd io.Reader) error {
	decoder := json.NewDecoder(rd)
	return decoder.Decode(tf)
}

func transactionFromJSON(tr *models.Transaction, rd io.Reader) error {
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
	userUpd := &models.User{userDB.ID, tr.Value, tr.ReserveValue}
	ha.hu.ValidateUserData(userUpd, userDB, err)
}

func (ha *AppHandler) ValidateTransactionJSON(tr *models.Transaction, err *error) {
	if *err != nil {
		return
	}
	if tr.OrderId <= 0 || tr.ServiceId <= 0 {
		*err = errors.New("Bad Order JSON")
	} else if tr.Status != "" && tr.Status != "in process" &&
		tr.Status != "approved" && tr.Status != "canceled" {
		*err = errors.New(`Bad order status`)
	}
}

type KeyTransactionPost struct{}

func (ha *AppHandler) MiddlewareAdditional(next http.Handler) http.Handler {

	ha.l.Println("In ADDITIONAL MIDDLEWARE")
	return ha.MiddlewareValidateNewTransaction(next)
}

func (ha *AppHandler) ValidateTransactionStatus(tr *models.Transaction, err *error) {
	status := tr.Status
	trDB, _ := ha.tr.FindLastTransactionByOrder(tr.OrderId)

	switch status {
	case "", "in process":
		if trDB != nil {
			*err = errors.New("Order already exists. State its new status")
		}
	case "approved", "canceled":
		if trDB == nil {
			*err = errors.New("Order doesn't exist. It can't be created with 'approved' or 'canceled' status.")
		} else if trDB.Status == "canceled" || trDB.Status == "approved" {
			*err = errors.New("Order is finished and can't be modified")
		}
	}
}

func (ha *AppHandler) PrepareTransactionValue(tr *models.Transaction) {
	switch tr.Status {
	case "", "in process":
		tr.ReserveValue = tr.Value
		tr.Value *= -1
	case "approved":
		tr.ReserveValue = tr.Value * -1
		tr.Value = 0
	case "canceled":
		tr.ReserveValue = tr.Value * -1
	}
}

func (ha *AppHandler) MiddlewareValidateNewTransaction(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		ha.l.Println("In TRANSACTION MIDDLEWARE")
		var tr = &models.Transaction{}
		err := transactionFromJSON(tr, rq.Body)

		if err != nil {
			ha.l.Println("Error:", err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
			enc := json.NewEncoder(rw)
			enc.Encode(PostResponseJSON{false})
			return
		}
		tr.Timesp = time.Now()

		ha.ValidateTransactionStatus(tr, &err)
		ha.ValidateTransactionJSON(tr, &err)
		//set how it changes the user's balance according to order status
		ha.PrepareTransactionValue(tr)
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

func (ha *AppHandler) ValidateTransfer(tf *models.Transfer, err *error) {
	// 1. both users exist
	// sender has enough money for transcation
	sender := ha.hu.ValidateUserInDb(tf.Sender, err)
	_ = ha.hu.ValidateUserInDb(tf.Recipient, err)
	if *err != nil {
		return
	}
	if sender.Balance < tf.Value {
		*err = ErrNotEnoughCredit
	}
}

type KeyTransfer struct{}

func (ha *AppHandler) MiddleWareValidateTransfer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		tf := &models.Transfer{}
		err := transferFromJSON(tf, rq.Body)
		if err != nil {
			ha.l.Println("Error:", err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
			enc := json.NewEncoder(rw)
			enc.Encode(PostResponseJSON{false})
			return
		}

		ha.ValidateTransfer(tf, &err)
		if err != nil {
			ha.l.Println(("Error:"), err.Error())
			rw.WriteHeader((http.StatusBadRequest))
			enc := json.NewEncoder(rw)
			enc.Encode(PostResponseJSON{false})
			return
		}

		ctx := context.WithValue(rq.Context(), KeyTransfer{}, tf)
		rq = rq.WithContext(ctx)

		next.ServeHTTP(rw, rq)
	})
}
