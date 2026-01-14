package usecase

import (
	"fmt"
	"testing"
)

func Test_getCompInfo(t *testing.T) {
	result, err := getCompInfo("https://yamanakasougoukenkyujo.com")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(result)
}
