package sectorstorage

import (
	"fmt"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao"
	"testing"
)

func TestRao(t *testing.T){
	err := rao.RunWebServer()
	fmt.Println(err)
}
