package handlers

import (
	"avito-user-balance/models"
	"avito-user-balance/repositories/postgres"
	"avito-user-balance/validate"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type UserRepository postgres.UserRepository

type UserHandler struct {
	l  *log.Logger
	ur *postgres.UserRepository
}

func NewUserHandler(lg *log.Logger, d *sql.DB) *UserHandler {
	return &(UserHandler{lg, postgres.NewUserRepository(d)})
}

func userToJSON(user models.User, wr http.ResponseWriter) error {
	wr.Header().Set("Content-type", "application/json")
	encoder := json.NewEncoder(wr)
	return encoder.Encode(user)
}

func userFromJSON(user *models.User, rd io.Reader) error {
	decoder := json.NewDecoder(rd)
	return (decoder.Decode(user))
}

func usersToJSON(users models.Users, wr http.ResponseWriter) error {
	wr.Header().Set("Content-type", "application/json")
	encoder := json.NewEncoder(wr)
	return encoder.Encode(users)
}

func (hu *UserHandler) GetUsers(rw http.ResponseWriter, rq *http.Request) {
	users, err := hu.ur.FindAllUsers()
	if err != nil {
		SendError(http.StatusNotFound, err, rw)
		return
	}

	err = usersToJSON(models.Users(users), rw)

	if err != nil {
		SendJSONError(err, "enconding users", rw)
		return
	}
}

func (hu *UserHandler) GetUser(rw http.ResponseWriter, rq *http.Request) {
	vars := mux.Vars(rq)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		SendError(http.StatusBadRequest, errors.New("Invalid User ID"), rw)
		return
	}

	var user *models.User
	user, err = hu.ur.FindUserByID(id)
	if user == nil {
		SendError(http.StatusNotFound, err, rw)
		return
	}

	err = userToJSON(*user, rw)
	if err != nil {
		SendJSONError(err, "encoding user", rw)
		return
	}
}

func (hu *UserHandler) updateUserData(userUpdate *models.User) error {
	oldValue, err := hu.ur.FindUserByID(userUpdate.ID)
	if err != nil {
		return err
	}
	userUpdate.Balance += oldValue.Balance
	userUpdate.Reserve += oldValue.Reserve
	err = hu.ur.UpdateUser(userUpdate)
	return err
}

func (hu *UserHandler) PostUsers(rw http.ResponseWriter, rq *http.Request) {
	var err error

	userUpdate := rq.Context().Value((KeyUserUpdate{})).(*models.User)
	userInDb := rq.Context().Value((KeyUserInDb{})).(*models.User)

	if userInDb == nil {
		hu.ur.AddUser(userUpdate)
	} else {
		err = hu.updateUserData(userUpdate)
	}
	if err != nil {
		SendError(http.StatusBadRequest, err, rw)
		return
	}
	SendSuccessful(rw)
}

func (hu *UserHandler) UserInDb(id int, err *error) *models.User {
	if *err == nil {
		user, _ := hu.ur.FindUserByID(id)
		if user == nil {
			*err = models.ErrUserNotFound
		}
		return user
	}
	return nil
}

type KeyUserUpdate struct{}
type KeyUserInDb struct{}

func (hu *UserHandler) MiddlewareValidateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		var userJSON = &models.User{}
		err := userFromJSON(userJSON, rq.Body)

		if err != nil {
			SendJSONError(err, "decoding user", rw)
			return
		}

		validate.ValidateUserID(userJSON.ID, &err)
		var userInDb *models.User
		userInDb, _ = hu.ur.FindUserByID(userJSON.ID)
		validate.ValidateUserData(userJSON, userInDb, &err)

		if err != nil {
			SendError(http.StatusBadRequest, err, rw)
			return
		}

		ctx := context.WithValue(context.WithValue(rq.Context(),
			KeyUserUpdate{}, userJSON), KeyUserInDb{}, userInDb)
		rq = rq.WithContext(ctx)

		next.ServeHTTP(rw, rq)
	})
}
