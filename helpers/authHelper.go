package helpers

import (
	"errors"
	"net/http"
)

func CheckUserType(w http.ResponseWriter, r *http.Request, role string) (err error) {
	userType := r.Header.Get("user_type")
	if userType != role {
		err = errors.New(userType)
		return err
	}
	return nil
}

func MatchUserTypeToUid(w http.ResponseWriter, r *http.Request, userId string) (err error) {
	userType := r.Header.Get("user_type")
	uid := r.Header.Get("uid")
	if userType == "USER" && uid != userId {
		err = errors.New("This Id is unauthorized to access this resource")
		return err
	}

	err = CheckUserType(w, r, userType)
	return err
}
