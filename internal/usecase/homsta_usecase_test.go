package usecase

import (
	"fmt"
	"testing"

	"github.com/zuxt268/sales/internal/model"
)

func Test_getCompInfo(t *testing.T) {
	result, err := getCompInfo("https://yamanakasougoukenkyujo.com")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(result)
}

func Test_getDiscInfo(t *testing.T) {
	homsta := model.Homsta{
		DiscUsage: "1.3G",
	}
	fmt.Println(homsta.GetDiscUsage())
}

func Test_Homsta(t *testing.T) {
	fmt.Println(getServer("/home/xb439432/hp"))
}
