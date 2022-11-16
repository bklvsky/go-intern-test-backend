package validate

import (
	"avito-user-balance/models"
	"errors"
	"fmt"
)

func ValidateUserID(id int, err *error) {
	if *err == nil {
		if id <= 0 {
			*err = fmt.Errorf("Bad User ID[%d] (should be positive number)", id)
		}
	}
}

func ValidateUserData(userUpd *models.User, user *models.User, err *error) {
	if *err != nil {
		return
	}

	if user == nil {
		user = &models.User{}
	}
	if (user.Balance+userUpd.Balance) < 0 ||
		(user.Reserve+userUpd.Reserve) < 0 {
		*err = models.ErrNotEnoughCredit
	}
}

func ValidateTransactionJSON(tr *models.Transaction, err *error) {
	if *err != nil {
		return
	}

	if tr.OrderId <= 0 {
		*err = fmt.Errorf("Bad Order ID[%d] (should be positive number)", tr.OrderId)
	} else if tr.ServiceId <= 0 {
		*err = fmt.Errorf("Bad Service ID[%d] (should be positive number)", tr.ServiceId)
	} else if tr.Status != "" && tr.Status != "in process" &&
		tr.Status != "approved" && tr.Status != "canceled" {
		*err = errors.New(`Bad order status (should be [in process/approved/canceled] or empty])`)
	}
}

func ValidateTransactionStatus(tr *models.Transaction, trDB *models.Transaction, err *error) {
	status := tr.Status

	switch status {
	case "", "in process":
		if trDB != nil {
			*err = fmt.Errorf("Order %d already exists. State its new status", trDB.OrderId)
		}
	case "approved", "canceled":
		if trDB == nil {
			*err = fmt.Errorf("Order doesn't exist. It can't be created with %s status.", status)
		} else if trDB.Status == "canceled" || trDB.Status == "approved" {
			*err = fmt.Errorf("Order is already %s and can't be modified (%s)",
				trDB.Status,
				status)
		}
	}
}
