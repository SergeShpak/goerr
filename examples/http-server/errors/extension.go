package errors

import "fmt"

func (e *Unauthorized) Hint() string {
	return fmt.Sprintf("operation in the namespace %s is not authorized", e.Payload.Namespace)
}
