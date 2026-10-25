package auth_service

import (
	"encoding/gob"
	"net/http"

	session_service "carf-chorf.com/app/app/services/session"
)

const (
	stateKey     = "auth"
	challengeKey = "otp"
)

type State struct {
	UserID uint
	Email  string
}

type OTPChallenge struct {
	Email     string
	CodeHash  string
	Intent    string
	Next      string
	ExpiresAt int64
	SentAt    int64
	Attempts  int
	Send      bool
}

func init() {
	gob.Register(State{})
	gob.Register(OTPChallenge{})
}

func SetAuthenticated(w http.ResponseWriter, r *http.Request, state State) error {
	sess, err := session_service.Get(r)
	if err != nil {
		return err
	}
	sess.Values[stateKey] = state
	delete(sess.Values, challengeKey)
	return sess.Save(r, w)
}

func Current(r *http.Request) (State, bool) {
	sess, err := session_service.Get(r)
	if err != nil {
		return State{}, false
	}
	state, ok := sess.Values[stateKey].(State)
	return state, ok
}

func IsAuthenticated(r *http.Request) bool {
	_, ok := Current(r)
	return ok
}

func ClearCurrent(w http.ResponseWriter, r *http.Request) error {
	sess, err := session_service.Get(r)
	if err != nil {
		return err
	}
	delete(sess.Values, stateKey)
	return sess.Save(r, w)
}

func SetChallenge(w http.ResponseWriter, r *http.Request, challenge OTPChallenge) error {
	sess, err := session_service.Get(r)
	if err != nil {
		return err
	}
	sess.Values[challengeKey] = challenge
	return sess.Save(r, w)
}

func Challenge(r *http.Request) (OTPChallenge, bool) {
	sess, err := session_service.Get(r)
	if err != nil {
		return OTPChallenge{}, false
	}
	challenge, ok := sess.Values[challengeKey].(OTPChallenge)
	return challenge, ok
}

func UpdateChallenge(w http.ResponseWriter, r *http.Request, challenge OTPChallenge) error {
	return SetChallenge(w, r, challenge)
}

func ClearChallenge(w http.ResponseWriter, r *http.Request) error {
	sess, err := session_service.Get(r)
	if err != nil {
		return err
	}
	delete(sess.Values, challengeKey)
	return sess.Save(r, w)
}

func Logout(w http.ResponseWriter, r *http.Request) error {
	sess, err := session_service.Get(r)
	if err != nil {
		return err
	}
	delete(sess.Values, stateKey)
	delete(sess.Values, challengeKey)
	return sess.Save(r, w)
}
