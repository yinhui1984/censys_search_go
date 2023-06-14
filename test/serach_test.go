package test

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/yinhui1984/censys_search_go"
)

func TestHostSearch(t *testing.T) {
	godotenv.Load()

	err := censyssearchgo.Search(censyssearchgo.Host, "services.http.response.body:\"this is a demo\"", true, func(hit censyssearchgo.Hits) {
		//Do your logic here
		t.Log(hit.IP)
	}, os.Getenv("API_ID"), os.Getenv("SECRET"))
	if err != nil {
		t.Error(err)
	}
}
