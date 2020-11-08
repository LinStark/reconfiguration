package reconfiguration

import (
	"fmt"
	"testing"
)

func TestReconfiguration_BelongShard(t *testing.T) {
	re := NewReconfiguration()
	re.ReadNode()
	re.ReadChain()
	for i:=0;i<200;i++{
		if i == 100{
			fmt.Println("pause")
		}
		fmt.Printf("第%d次调整\n",i+1)
		re.GenerateCrediblePlan()
		re.FillReconfiguration()
		re.checkNodes()
		re.SendToAdjust()
	}

}



