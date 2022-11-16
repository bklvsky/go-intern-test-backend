package handlers

import (
	"avito-user-balance/models"
	"avito-user-balance/validate"
	"context"
	"net/http"
	"time"
)

func (ha *AppHandler) MiddleWareValidateTransfer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		tf := &models.Transfer{}
		err := transferFromJSON(tf, rq.Body)
		if err != nil {
			SendJSONError(err, "parsing Transfer", rw)
			return
		}

		ha.ValidateTransfer(tf, &err)
		if err != nil {
			SendError(http.StatusBadRequest, err, rw)
			return
		}

		ctx := context.WithValue(rq.Context(), KeyTransfer{}, tf)
		rq = rq.WithContext(ctx)

		next.ServeHTTP(rw, rq)
	})
}

func (ha *AppHandler) MiddleWareHistory(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		hReq := models.HistoryRequest{}
		err := historyFromJSON(&hReq, rq.Body)
		if err != nil {
			SendJSONError(err, "parsing history request", rw)
			return
		}
		validate.ValidateUserID(hReq.UserId, &err)
		validate.ValidateHistoryRequest(&hReq, &err)
		if err != nil {
			SendError(http.StatusBadRequest, err, rw)
			return
		}
		ctx := context.WithValue(rq.Context(), KeyHistory{}, &hReq)
		rq = rq.WithContext(ctx)

		next.ServeHTTP(rw, rq)
	})
}


func (ha *AppHandler) MiddlewareValidateNewTransaction(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		var tr = &models.Transaction{}
		err := transactionFromJSON(tr, rq.Body)

		if err != nil {
			SendJSONError(err, "parsing order", rw)
			return
		}
		tr.Timesp = time.Now()

		trDB, _ := ha.tr.FindLastTransactionByOrder(tr.OrderId)
		validate.ValidateTransactionStatus(tr, trDB, &err)
		validate.ValidateTransactionJSON(tr, &err)
		ha.PrepareTransactionValue(tr)
		ha.ValidateUser(tr, &err)

		if err != nil {
			SendError(http.StatusBadRequest, err, rw)
			return
		}

		ctx := context.WithValue(rq.Context(), KeyTransactionPost{}, tr)
		rq = rq.WithContext(ctx)

		next.ServeHTTP(rw, rq)
	})
}