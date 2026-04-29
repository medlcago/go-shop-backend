package apperror

import (
	"errors"
	"fmt"
	"testing"
)

func TestWrap_Example(t *testing.T) {
	baseErr := errors.New("sql: no rows in result set")

	err1 := Wrap("repository.GetWishlist", baseErr)
	err2 := Wrap("service.GetWishlist", err1)
	err3 := Wrap("handler.GetWishlist", err2)

	fmt.Println("Message:", err3.Message)
	fmt.Println("Error():", err3.Error())
	fmt.Println("Full error:", err3.Err)
	fmt.Println(errors.Is(err3, baseErr))

	if appErr, ok := errors.AsType[*AppError](err3); ok {
		fmt.Println("Code:", appErr.Code)
	}
}
