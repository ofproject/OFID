package httpapi

import "fmt"
type signError struct {
	hit string
}

func (e *signError) ErrorCode() int { return -100 }

func (e *signError) Error() string {
	return fmt.Sprintf(e.hit)
}


type pubMatchError struct {
	hit string
}

func (e *pubMatchError) ErrorCode() int { return -200 }

func (e *pubMatchError) Error() string {
	return fmt.Sprintf(e.hit)
}

type whiteListErr struct {
	hit string
}

func (e *whiteListErr) ErrorCode() int { return -200 }

func (e *whiteListErr) Error() string {
	return fmt.Sprintf(e.hit)
}


type backIpError struct {
	hit string
}

func (e *backIpError) ErrorCode() int { return -200 }

func (e *backIpError) Error() string {
	return fmt.Sprintf(e.hit)
}
