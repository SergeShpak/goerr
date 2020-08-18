package handlers

import (
	"log"
	"net/http"

	"github.com/SergeyShpak/goerr/examples/http/errors"
)

func AccessAdminConsole(w http.ResponseWriter, r *http.Request) {
	if err := checkRequest(r); err != nil {
		data := errors.PrepareErrorToSend(err)
		log.Printf("an error occurred: %s\n%s", data.Attrs.Msg, data.Attrs.Stack)
		w.WriteHeader(data.Attrs.HTTPCode)
		w.Write([]byte(data.Hint))
		return
	}
	w.Write([]byte("ok"))
}

func checkRequest(r *http.Request) error {
	err := errors.NewUnauthorized("authorization failed", nil, &errors.UnauthorizedPayload{
		Namespace: "admin",
	})
	return err
}
