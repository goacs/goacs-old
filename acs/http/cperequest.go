package http

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"goacs/acs"
	acsxml "goacs/acs/types"
	"net/http"
)

type CPERequest struct {
	Request      *http.Request
	Response     http.ResponseWriter
	DBConnection *sqlx.DB
	Session      *acs.ACSSession
	Envelope     *acsxml.Envelope
	Body         []byte
	ReqType      string
}

func (r *CPERequest) SendResponse(body string) {
	//log.Println(body)
	_, _ = fmt.Fprint(r.Response, body)
}
