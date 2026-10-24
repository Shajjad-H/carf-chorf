package flashes_service

import (
	"encoding/gob"
	"net/http"

	session_service "carf-chorf.com/app/app/services/session"
)

const (
	sessionKey = "flashes"
)

func init() {
	// The value lives in map[interface{}]interface{}; gob must know the concrete type.
	gob.Register(map[string][]string{})
}

func AddFlash(w http.ResponseWriter, r *http.Request, key, message string) {
	sess, _ := session_service.Get(r)

	m, ok := sess.Values[sessionKey].(map[string][]string)
	if !ok {
		m = map[string][]string{}
		sess.Values[sessionKey] = m
	}
	m[key] = append(m[key], message)

	_ = sess.Save(r, w)
}

func GetFlashes(w http.ResponseWriter, r *http.Request, key string) []string {
	sess, _ := session_service.Get(r)

	m, ok := sess.Values[sessionKey].(map[string][]string)
	if !ok {
		return nil
	}

	messages := m[key]
	if len(messages) == 0 {
		return nil
	}

	delete(m, key)
	_ = sess.Save(r, w)
	return messages
}
