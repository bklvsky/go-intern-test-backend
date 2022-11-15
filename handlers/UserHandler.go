package handlers

import (
	"avito-user-balance/models"
	"avito-user-balance/repositories/postgres"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"

	// "encoding/json"
	"log"
	"net/http"

	// "regexp"
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
	encoder := json.NewEncoder(wr)
	return encoder.Encode(user)
}

func userFromJSON(user *models.User, rd io.Reader) error {
	decoder := json.NewDecoder(rd)
	return (decoder.Decode(user))
}

func usersToJSON(users models.Users, wr io.Writer) error {
	encoder := json.NewEncoder(wr)
	return encoder.Encode(users)
}

func (hu *UserHandler) GetUsers(rw http.ResponseWriter, rq *http.Request) {
	hu.l.Println("GET/", rq.URL.Path)
	// vars := mux.Vars(rq)

	// if _, ok := vars["id"]; ok {
	// 	hu.l.Println("Looking for models.USER IN DB")
	// 	hu.getUser(rw, rq)
	// } else {
	users, err := hu.ur.FindAllUsers()
	if err != nil {
		hu.l.Println("Database error:", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	err = usersToJSON(models.Users(users), rw)

	if err != nil {
		hu.l.Println("Error encoding json", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	// }
}

func (hu *UserHandler) GetUser(rw http.ResponseWriter, rq *http.Request) {
	vars := mux.Vars(rq)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		hu.l.Println("URL error, invalid models.userID")
		http.Error(rw, "Invalid Request", http.StatusBadRequest)
		return
	}

	// var models.userServ models.UserService
	user, error := hu.ur.FindUserByID(id)
	if user == nil {
		hu.l.Println(error.Error())
		http.Error(rw, error.Error(), http.StatusNotFound)
		return
	}

	err = userToJSON(*user, rw)
	if err != nil {
		hu.l.Println("Error encoding json", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

var ErrNotEnoughCredit = fmt.Errorf("Not enough money in the account")

type PostResponseJSON struct {
	Status bool `json:"status"`
}

func (hu *UserHandler) updateUserData(userUpdate *models.User) error {
	oldValue, err := hu.ur.FindUserByID(userUpdate.ID)
	if (err != nil) {
		return err
	}
	userUpdate.Balance += oldValue.Balance
	userUpdate.Reserve += oldValue.Reserve
	err = hu.ur.UpdateUser(userUpdate)
	return err
}

func (hu *UserHandler) PostUsers(rw http.ResponseWriter, rq *http.Request) {
	hu.l.Println("Handle POST models.User")
	var err error

	userUpdate := rq.Context().Value((KeyUserUpdate{})).(*models.User)
	userInDb := rq.Context().Value((KeyUserInDb{})).(*models.User)

	if userInDb == nil {
		hu.ur.AddUser(userUpdate)
	} else {
		err = hu.updateUserData(userUpdate)
	}
	if err != nil {
		hu.l.Println("Error:", err.Error())
		rw.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(rw)
		enc.Encode(PostResponseJSON{false})
		return
	}
	enc := json.NewEncoder(rw)
	enc.Encode(PostResponseJSON{true})
}

func (hu *UserHandler) ValidateUserData(userUpd *models.User, user *models.User, err *error) {
	if *err == nil {
		if user == nil {
			user = &models.User{}
		}
		if ((user.Balance + userUpd.Balance) < 0 ||
		(user.Reserve + userUpd.Reserve) < 0) {
			*err = ErrNotEnoughCredit
		}
	}
}

func (hu *UserHandler) ValidateUserInDb(id int, err *error) *models.User {
	if *err == nil {
		user, _ := hu.ur.FindUserByID(id)
		if user == nil {
			*err = models.ErrUserNotFound
		}
		return user
	}
	return nil
}

func (hu *UserHandler) ValidateUserID(id int, err *error) {
	if *err == nil {
		if id <= 0 {
			*err = fmt.Errorf("Bad models.User ID")
		}
	}
}

// I NEED TO MAKE AN ERROR STRUCT IN REPO
// IT WILL HAVE FIELDS:		MESSAGE (ERROR())
//							STATUS
//							DESCRIPTION (OPTIONAL)

// func (hu *UserHandler) MiddleWareValidateUser(next http.Handler) http.Handler {
// 	return http.HandlerFunc( func(rw http.ResponseWriter, rq *http.Request) {

// 	})
// }

type KeyUserUpdate struct{}
type KeyUserInDb struct{}

func (hu *UserHandler) MiddlewareValidateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		hu.l.Println("IN MIDDLEWARE")
		var userJSON = &models.User{}
		err := userFromJSON(userJSON, rq.Body)

		if err != nil {
			hu.l.Println("Error:", err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
			enc := json.NewEncoder(rw)
			enc.Encode(PostResponseJSON{false})
			return
		}

		hu.ValidateUserID(userJSON.ID, &err)
		var userInDb *models.User
		userInDb, _ = hu.ur.FindUserByID(userJSON.ID)
		hu.ValidateUserData(userJSON, userInDb, &err)

		if err != nil {
			hu.l.Println("Error:", err.Error())
			rw.WriteHeader(http.StatusBadRequest)
			enc := json.NewEncoder(rw)
			enc.Encode(PostResponseJSON{false})
			return
		}

		ctx := context.WithValue(context.WithValue(rq.Context(),
		KeyUserUpdate{}, userJSON), KeyUserInDb{}, userInDb)
		rq = rq.WithContext(ctx)

		next.ServeHTTP(rw, rq)
	})
}
