package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

func main() {
	contents, err := readFile("examples/http-server/errors.json")
	if err != nil {
		log.Fatal(err)
	}
	errorDefinitions, err := parseErrorDefinitions(contents)
	if err != nil {
		log.Fatal(err)
	}
	errorCodeDefinitions, err := generateErrorDefinitions(errorDefinitions)
	if err != nil {
		log.Fatal(err)
	}
	errorIDs, err := generateErrorIDs(errorDefinitions)
	if err != nil {
		log.Fatal(err)
	}
	o := &outFile{
		errorIDs:         errorIDs,
		imports:          imports,
		predefined:       predefined,
		errorDefinitions: errorCodeDefinitions,
	}
	result := o.compose("errors")
	formattedResult, err := formatCode(result)
	if err != nil {
		log.Fatal(err)
	}
	if err := writeGenerated("examples/http-server/errors/errors.go", formattedResult); err != nil {
		log.Fatal(err)
	}
}

type errorDefinition struct {
	HTTPCode int
	Payload  map[string]string
}

type errID struct {
	ErrName string
	ID      string
}

func (e *errID) String() string {
	return fmt.Sprintf("%s = \"%s\"", e.ErrName, e.ID)
}

type outFile struct {
	errorIDs         []*errID
	imports          string
	predefined       string
	errorDefinitions []*errorCodeDefinition
}

func (f *outFile) compose(packageName string) string {
	parts := make([]string, 0, 4+len(f.errorDefinitions))
	errIDConstants := make([]string, len(f.errorIDs))
	for i, id := range f.errorIDs {
		errIDConstants[i] = id.String()
	}
	parts = append(parts, []string{
		fmt.Sprintf("package %s", packageName),
		f.imports,
		f.predefined,
		fmt.Sprintf("const (\n%s\n)", strings.Join(errIDConstants, "\n")),
	}...)
	for _, def := range f.errorDefinitions {
		parts = append(parts, def.String())
	}
	result := strings.Join(parts, "\n")
	return result
}

func parseErrorDefinitions(defs []byte) (map[string]errorDefinition, error) {
	var jsonDefs map[string]errorDefinition
	if err := json.Unmarshal(defs, &jsonDefs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal errors definitions: %v", err)
	}
	return jsonDefs, nil
}

func generateErrorDefinitions(defs map[string]errorDefinition) ([]*errorCodeDefinition, error) {
	errorDefinitions := make([]*errorCodeDefinition, 0, len(defs))
	for errName, d := range defs {
		errDef, err := generateErrorDefinition(errName, &d)
		if err != nil {
			return nil, fmt.Errorf("failed to generate code for the error %s: %v", errName, err)
		}
		errorDefinitions = append(errorDefinitions, errDef)
	}
	return errorDefinitions, nil
}

func generateErrorIDs(defs map[string]errorDefinition) ([]*errID, error) {
	errorIDs := make([]*errID, 0, len(defs)+2)
	for errName, d := range defs {
		id := generateErrorID(errName, &d)
		constName, err := generateFromTemplate(tmplErrorIDConstName, &tmplErrorIDConstNameIn{
			errorNameIn: errorNameIn{
				ErrorName: errName,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to generate an ID for the error %s from a template: %v", errName, err)
		}
		errID := &errID{
			ErrName: constName,
			ID:      id,
		}
		errorIDs = append(errorIDs, errID)
	}
	return errorIDs, nil
}

func readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file %s: %v", path, err)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read the file %s: %v", path, err)
	}
	return data, nil
}

func writeGenerated(path string, result string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to open the file %s: %v", path, err)
	}
	defer f.Close()

	if _, err := f.Write([]byte(result)); err != nil {
		return fmt.Errorf("failed to write the result to the file %s: %v", path, err)
	}
	return nil
}

type errorCodeDefinition struct {
	Payload     string
	ErrorType   string
	Constructor string
}

func (d *errorCodeDefinition) String() string {
	result := fmt.Sprintf("%s\n%s\n%s", d.Payload, d.ErrorType, d.Constructor)
	return result
}

func generateErrorDefinition(name string, def *errorDefinition) (*errorCodeDefinition, error) {
	codeDef := &errorCodeDefinition{}
	var err error
	hasPayload := len(def.Payload) != 0
	if hasPayload {
		codeDef.Payload, err = generateErrorPayloadDefinition(name, def.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to generate a payload structure for the error %s: %v", name, err)
		}
	}
	codeDef.ErrorType, err = generateErrorTypeDefinition(name, hasPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to generate the error type definition for the error %s: %v", name, err)
	}
	if def.HTTPCode == 0 {
		def.HTTPCode = http.StatusInternalServerError
	}
	codeDef.Constructor, err = generateErrorConstructor(&tmplErrorConstructorIn{
		tmplErrorTypeIn: tmplErrorTypeIn{
			tmplErrorPayloadNameIn: tmplErrorPayloadNameIn{
				ErrorName: name,
			},
			HasPayload: hasPayload,
		},
		HTTPCode: def.HTTPCode,
	})
	return codeDef, nil
}

func generateErrorPayloadDefinition(errorName string, payload map[string]string) (string, error) {
	def, err := generateFromTemplate(tmplErrorPayload, &tmplErrorPayloadIn{
		errorNameIn: errorNameIn{
			ErrorName: errorName,
		},
		Fields: payload,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate the definition from template: %v", err)
	}
	return def, nil
}

func generateErrorTypeDefinition(errorName string, hasPayload bool) (string, error) {
	def, err := generateFromTemplate(tmplErrorType, &tmplErrorTypeIn{
		tmplErrorPayloadNameIn: tmplErrorPayloadNameIn{
			ErrorName: errorName,
		},
		HasPayload: hasPayload,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate the definition from template: %v", err)
	}
	return def, nil
}

func generateErrorConstructor(in *tmplErrorConstructorIn) (string, error) {
	def, err := generateFromTemplate(tmplErrorConstructor, in)
	if err != nil {
		return "", fmt.Errorf("template generation failed: %v", err)
	}
	return def, nil
}

func generateErrorID(errName string, d *errorDefinition) string {
	toHash := make([]string, len(d.Payload)+1)
	for field := range d.Payload {
		toHash = append(toHash, field)
	}
	sort.Slice(toHash, func(i int, j int) bool {
		return toHash[i] < toHash[j]
	})
	toHash = append(toHash, errName)
	toHashString := strings.Join(toHash, ",")
	h := sha256.New()
	h.Write([]byte(toHashString))
	hash := h.Sum(nil)
	id := hex.EncodeToString(hash)
	return id
}

func formatCode(code string) (string, error) {
	formattedCode, err := format.Source([]byte(code))
	if err != nil {
		return "", fmt.Errorf("code formatting failed: %v", err)
	}
	return string(formattedCode), nil
}
