package checkdb
import (
	"crypto/sha256"
	"fmt"
	dbm"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"testing"
)
func TestInitAddDB(t *testing.T) {
	InitAddDB(dbm.NewMemDB(),log.TestingLogger())
	txStr := &TX{Content:"100",Sender:"A"}
	txStr.ID = sha256.Sum256([]byte(txStr.Content))

	Save(txStr.ID,txStr,0)

	fmt.Println("存储完成")
	tx:=Search(txStr.ID,0)
	fmt.Println(tx)
}