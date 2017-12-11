package db

import "fmt"

type logingError struct {
	account []string

}

func (e *logingError) ErrorCode() int { return -32601 }

func (e *logingError) Error() string {
	return fmt.Sprintf("The account can not loging: ", e.account)
}
