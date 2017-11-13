package shamir

import "fmt"

/*
*   KeyNeed>KeyNumber = -100
*/
type keyNumberError struct {
	hit string
}

func (e *keyNumberError) ErrorCode() int { return -100 }

func (e *keyNumberError) Error() string {
	return fmt.Sprintf(e.hit)
}


