package auth_test

import (
	"testing"
	"time"

	"github.com/Norrun/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func TestBasic(t *testing.T) {
	id := uuid.New()
	secret := "soem sercet"
	token, err := auth.MakeJWT(id, secret, time.Hour)
	if err != nil {
		t.Error(err)
	}
	id2, err := auth.ValidateJWT(token, secret)
	if err != nil {
		t.Error(err)
	}
	if id != id2 {
		t.Errorf("%s is not %s", id.String(), id2.String())
	}

}
