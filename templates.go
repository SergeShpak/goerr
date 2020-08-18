package main

import (
	"bytes"
	"text/template"
)

type errorNameIn struct {
	ErrorName string
}

type tmplErrorPayloadNameIn errorNameIn

var tmplErrorPayloadName = template.Must(template.New("ErrorPayloadName").
	Parse(`{{.ErrorName}}Payload`))

type tmplErrorPayloadIn struct {
	errorNameIn
	Fields map[string]string
}

var tmplErrorPayload = template.Must(tmplErrorPayloadName.New("ErrorPayload").Parse(
	`type {{template "ErrorPayloadName" .}} struct {
		{{- range $name, $type := .Fields }}
		{{ $name }} {{ $type }}
	 	{{- end }}
	}`))

type tmplErrorTypeIn struct {
	tmplErrorPayloadNameIn
	HasPayload bool
}

var tmplErrorType = template.Must(tmplErrorPayloadName.New("ErrorPayloadType").Parse(
	`type {{.ErrorName}} struct {
		BaseError
		{{if .HasPayload}}Payload *{{template "ErrorPayloadName" .}}{{end}}
	}`))

type tmplErrorIDConstNameIn struct {
	errorNameIn
}

var tmplErrorIDConstName = template.Must(tmplErrorPayloadName.New("ErrorIDConstName").
	Parse(`ErrID{{.ErrorName}}`))

type tmplErrorConstructorIn struct {
	tmplErrorTypeIn
	HTTPCode int
}

var tmplErrorConstructor = template.Must(tmplErrorIDConstName.New("ErrorConstructor").Parse(
	`func New{{.ErrorName}}(msg string, err error{{if .HasPayload}}, payload *{{template "ErrorPayloadName" .}}{{end}}) error {
		eMsg := newErrorMessage(msg, err)
		e := &{{.ErrorName}}{
			BaseError: BaseError{
				ErrorAttributes: ErrorAttributes{
					ID:            {{template "ErrorIDConstName" .}},
					HTTPCode:      {{.HTTPCode}},
					Msg:           eMsg,
					Stack:         getStack(),
					OriginalError: err,
				},
			},
			{{if .HasPayload}}Payload: payload,{{end}}
		}
		return e
	}`))

func generateFromTemplate(tmpl *template.Template, in interface{}) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, in); err != nil {
		return "", err
	}
	return buf.String(), nil
}
