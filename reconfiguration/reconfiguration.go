package reconfiguration

import (
	"encoding/json"
	"fmt"
	"github.com/reconfiguration/SGX"

	//"github.com/tendermint/tendermint/SGX"
	//cs "github.com/tendermint/tendermint/consensus"
	//"github.com/tendermint/tendermint/libs/log"
	//"github.com/tendermint/tendermint/rpc/client"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Reconfiguration struct {
	timestamp  time.Time
	FillArea   [][]int
	Nodesinfo  [][]Nodeinfo
	SendNodes  [][]Nodeinfo //需要调整的节点
	ShardCount int
	NodeCount  int
	MoveCount  int
	Txs        [][]Tx
	IsLeader   bool
	//Cs         *cs.ConsensusState
	//Client     client.HTTP
	//logger     log.Logger
	Chaininfo  []string
	Count      []int
	FlagSend   [][]bool
	FlagCreate [][]bool
}

//type ReConfigurationConfig struct{
//	ShardCount string
//	NodeCount  string
//	MoveCount  string
//}
const (
	sendTimeout = time.Second * 10
)

type Tx []byte
type Nodeinfo struct {
	ShardName  string
	Coordinate string
	NodeName   string
	PeerId     string
	// ChainId    string
	Neighbor []int
}

// type ChainInfo struct {
// 	ChainName string
// 	ChainId   string
// }

var PeriodCount = 0
var Ticker = time.NewTicker(time.Second * 1000)
var tenToAny map[int]string = map[int]string{0: "0", 1: "1", 2: "2", 3: "3", 4: "4", 5: "5", 6: "6", 7: "7", 8: "8", 9: "9", 10: "a", 11: "b", 12: "c", 13: "d", 14: "e", 15: "f", 16: "g", 17: "h", 18: "i", 19: "j", 20: "k", 21: "l", 22: "m", 23: "n", 24: "o", 25: "p", 26: "q", 27: "r", 28: "s", 29: "t", 30: "u", 31: "v", 32: "w", 33: "x", 34: "y", 35: "z", 36: ":", 37: ";", 38: "<", 39: "=", 40: ">", 41: "?", 42: "@", 43: "[", 44: "]", 45: "^", 46: "_", 47: "{", 48: "|", 49: "}", 50: "A", 51: "B", 52: "C", 53: "D", 54: "E", 55: "F", 56: "G", 57: "H", 58: "I", 59: "J", 60: "K", 61: "L", 62: "M", 63: "N", 64: "O", 65: "P", 66: "Q", 67: "R", 68: "S", 69: "T", 70: "U", 71: "V", 72: "W", 73: "X", 74: "Y", 75: "Z"}

func NewReconfiguration() *Reconfiguration {
	ShardCount, _ := strconv.Atoi("4")
	NodeCount, _ := strconv.Atoi("16")
	MoveCount, _ := strconv.Atoi("3")
	// fmt.Println("ShardCount", ShardCount, "NodeCount", NodeCount)
	Re := &Reconfiguration{
		ShardCount: ShardCount,
		NodeCount:  NodeCount,
		MoveCount:  MoveCount,
		Count:      make([]int, ShardCount),
		//Cs:         cs,
		//logger:     l,
	}
	Re.SendNodes = make([][]Nodeinfo, Re.ShardCount)
	//Re.ReadNode()  //读取数据
	//Re.ReadChain() //读取链数据
	return Re
}

type JsonStruct struct {
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
	//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	//读取的数据为json格式，需要进行解码
	err = json.Unmarshal(data, v)
	if err != nil {
		return
	}
}

func (Re *Reconfiguration) ReadNode() {
	Re.Nodesinfo = make([][]Nodeinfo, Re.ShardCount)
	for i := 0; i < len(Re.Nodesinfo); i++ {
		Re.Nodesinfo[i] = make([]Nodeinfo, Re.NodeCount)
	}
	JsonParse := NewJsonStruct()
	v := []Nodeinfo{}
	JsonParse.Load("data.json", &v)
	for i := 0; i < len(v); i++ {
		// fmt.Println("i值：", i)
		ShardIndex, _ := strconv.Atoi(v[i].Coordinate)
		Shard, _ := strconv.Atoi(v[i].ShardName)

		Re.Nodesinfo[Shard][ShardIndex] = v[i]
	}
}

func (Re *Reconfiguration) ReadChain() {
	Re.Chaininfo = make([]string, Re.ShardCount)
	JsonParse := NewJsonStruct()
	v := []string{}
	JsonParse.Load("shard.json", &v)
	Re.Chaininfo = v

}
func (Re *Reconfiguration) PeriodReconfiguration() {
	//定时器实现
	time.Sleep(time.Second * 5)
	//Re.Client = *client.NewHTTP("localhost:26657", "/websocket")
	//if Re.Cs.IsLeader() {
		PeriodCount++
		//Re.logger.Info("Leader is Me")
		//BlockTimeStamp := Re.LatestBlockTime()
		//Re.GenerateReconfiguration(BlockTimeStamp)
		Re.FillReconfiguration()
		Re.SendReconfiguration()
		Re.SendToAdjust()

	//}
	for {
		select {
		case <-Ticker.C:

			//if Re.Cs.IsLeader() {
				PeriodCount++
				//Re.logger.Info("Leader is Me")
				//BlockTimeStamp := Re.LatestBlockTime()
				//Re.GenerateReconfiguration(BlockTimeStamp)
				Re.FillReconfiguration()
				Re.SendReconfiguration()
				Re.SendToAdjust()

			//}
		}
	}
}
//func (Re *Reconfiguration) LatestBlockTime() time.Time {
//	//获取区块的可信时间戳
//	status, err := Re.Client.Status()
//	if err != nil {
//		Re.logger.Error("err:", "Reconfiguration", err)
//	}
//	return status.SyncInfo.LatestBlockTime
//}
func (Re *Reconfiguration) GenerateReconfiguration(BlockStamp time.Time) {
	//生成可信方案
	CredibleTimeStamp := SGX.GetCredibleTimeStamp()

	if CredibleTimeStamp.Sub(BlockStamp) > 1 {
		//若满足时间间隔，则进行生成调整方案

		Re.GenerateCrediblePlan()

	}
}
func (Re *Reconfiguration) GenerateCrediblePlan() {
	//清空原有元素
	Re.FillArea = [][]int{}
	for i := 0; i < Re.ShardCount; i++ {
		Re.FillArea = append(Re.FillArea, Re.GenerateCredibleFillArea(i))
	}
}
func Contain(FillArea []int, data int) bool {
	for i := 0; i < len(FillArea); i++ {
		if FillArea[i] == data {
			return true
		}
	}
	return false
}
func (Re *Reconfiguration) GenerateCredibleFillArea(Shard int) []int {
	//生成单片调整计划
	var FillArea []int
	var Total int
	Total = 0
	for {
		if Total < Re.MoveCount {
			data := SGX.GetCredibleRand(Re.NodeCount)
			if !Contain(FillArea, data) {
				FillArea = append(FillArea, data)
				Total++
			}
		} else {
			break
		}
	}

	return FillArea
}

// func NewChain(Shard int) []*ChainInfo {
// 	var ChainInfos []*ChainInfo
// 	for i := 0; i < Shard; i++ {
// 		rand.Seed(time.Now().UnixNano())
// 		rnd_Chain_Id := rand.Int()
// 		time.Sleep(time.Millisecond)
// 		ChainIn := &ChainInfo{
// 			ChainName: strconv.Itoa(i),
// 			ChainId:   strconv.Itoa(rnd_Chain_Id),
// 		}
// 		ChainInfos = append(ChainInfos, ChainIn)
// 	}
// 	return ChainInfos
// }
// func (Re *Reconfiguration) NewNodes(ShardCount int, NodeCount int) [][]Nodeinfo {
// 	var Nodesinfo [][]Nodeinfo
// 	Nodesinfo = make([][]Nodeinfo, ShardCount)
// 	for i := 0; i < ShardCount; i++ {
// 		for j := 0; j < NodeCount; j++ {
// 			rand.Seed(time.Now().UnixNano())
// 			rnd_PeerId := rand.Int()
// 			time.Sleep(time.Millisecond)
// 			Node := Nodeinfo{
// 				NodeName:   "TT" + strconv.Itoa(i) + "Node" + strconv.Itoa(j),
// 				Coordinate: GenerateCoordinate(decimalToAny(j, 4)),
// 				PeerId:     strconv.Itoa(rnd_PeerId),
// 				ChainId:    Re.GetShardInfo(i),
// 			}
// 			Nodesinfo[i] = append(Nodesinfo[i], Node)
// 		}
// 	}
// 	return Nodesinfo
// }
// func (Re *Reconfiguration) GetShardInfo(Shard int) string {
// 	return Re.ChInfo[Shard].ChainId
// }
func (Re *Reconfiguration) SendReconfiguration() {
	//发送交易到区块
	//Re.logger.Info("Sending Reconfiguration")
	Re.Txs = make([][]Tx, Re.ShardCount)
	//res, _ := json.Marshal(Re.Nodesinfo)

	//go Re.Client.BroadcastTxAsync(res)
	//for i := 0; i < Re.ShardCount; i++ {
	//	Re.Txs[i] = make([]Tx, Re.NodeCount)
	//	for j := 0; j < Re.NodeCount; j++ {
	//		tx, _ := json.Marshal(Re.Nodesinfo[i][j])
	//		//Re.logger.Info("tx:", "Reconfiguration", string(tx))
	//		go Re.Client.BroadcastTxAsync(tx)
	//	}
	//}
}
func decimalToAny(num int, n int) string {
	new_num_str := ""
	var remainder int
	var remainder_string string
	if num == 0 {
		new_num_str = "0"
	}
	for num != 0 {
		remainder = num % n
		if 76 > remainder && remainder > 9 {
			remainder_string = tenToAny[remainder]
		} else {
			remainder_string = strconv.Itoa(remainder)
		}
		new_num_str = remainder_string + new_num_str
		num = num / n
	}

	return new_num_str
}
func GenerateCoordinate(ndecimal string) string {
	if len(ndecimal) < 3 {
		for i := 0; i < 3-len(ndecimal)+1; i++ {
			ndecimal = "0" + ndecimal
		}
	}
	return ndecimal
}

// func (Re *Reconfiguration) GetNodeInfo() [][]Nodeinfo {
// 	var Nodesinfo [][]Nodeinfo
// 	Nodesinfo = make([][]Nodeinfo, Re.ShardCount)
// 	for i := 0; i < Re.ShardCount; i++ {
// 		for j := 0; j < Re.NodeCount; j++ {
// 			rand.Seed(time.Now().UnixNano())
// 			rnd_PeerId := rand.Int()
// 			time.Sleep(time.Millisecond)
// 			Node := Nodeinfo{
// 				NodeName:   "TT" + strconv.Itoa(i) + "Node" + strconv.Itoa(j),
// 				Coordinate: GenerateCoordinate(decimalToAny(i, 4)),
// 				PeerId:     strconv.Itoa(rnd_PeerId),
// 				ChainId:    Re.GetShardInfo(i),
// 			}
// 			Nodesinfo[i] = append(Nodesinfo[i], Node)
// 		}
// 	}
// 	return Nodesinfo
// }
func String2Int(s uint8) int {
	a := string(s)
	s1, _ := strconv.Atoi(a)
	return s1
}
func Modify(s string, index int, modify string) string {
	var data []byte = []byte(s)
	data[index] = modify[0]
	ModifiedString := string(data)
	return ModifiedString
}
func GetNeighbor(x string, m int) []string {
	list := []string{}
	for i := 0; i < 3; i++ {
		if String2Int(x[i])+1 >= 0 && String2Int(x[i])+1 < m {
			Node := Modify(x, i, strconv.Itoa(String2Int(x[i])+1))
			list = append(list, Node)
		}
		if String2Int(x[i])-1 >= 0 {
			Node := Modify(x, i, strconv.Itoa(String2Int(x[i])-1))
			list = append(list, Node)
		}

	}
	return list
}
func reverse(str string) string { //字符串反转
	rs := []rune(str)
	len := len(rs)
	var tt []rune

	tt = make([]rune, 0)
	for i := 0; i < len; i++ {
		tt = append(tt, rs[len-i-1])
	}
	return string(tt[0:])
}
func (Re *Reconfiguration) GetNeighborName(x string, m int) []int {
	Neighborlist := GetNeighbor(x, m)
	var NameList []int
	for i := 0; i < len(Neighborlist); i++ {
		sum := 0
		for j := 0; j < len(Neighborlist[i]); j++ {
			ss := reverse(Neighborlist[i])
			sum = sum + String2Int(ss[j])*int(math.Pow(float64(m), float64(j)))

		}
		if sum < Re.NodeCount {
			NameList = append(NameList, sum)
		}
	}
	return NameList
}

//watchshard的作用是，看看哪一个分片还没被填满并且不能填到自身分片。
func WatchShard(BelongShard int, FullShard []int, Shard int) bool {
	for i := 0; i < len(FullShard); i++ {
		if BelongShard == FullShard[i] {
			return true
		}
	}
	if BelongShard == Shard {
		return true
	}
	return false
}
func (Re *Reconfiguration) BelongShard(ShardCount int, FullShard []int, i int) int {

	rand.Seed(time.Now().UnixNano())
	BelongShard := rand.Intn(ShardCount)
	time.Sleep(time.Millisecond * 5)
	for {

		if WatchShard(BelongShard, FullShard, i) && Re.WatchError() {
			rand.Seed(time.Now().UnixNano())
			BelongShard = rand.Intn(ShardCount)
			time.Sleep(time.Millisecond * 5)
			continue
		} else {

			if !Re.WatchError() {
				return -1
			}
			break
		}
	}
	Re.Count[BelongShard]++
	return BelongShard
}

//看看前面的所有分片是否已经满了
func (Re *Reconfiguration) WatchError() bool {
	for i := 0; i < Re.ShardCount-1; i++ {
		if Re.Count[i] != Re.MoveCount {
			return true
		}
	}
	return false
}

//判断是否可以放入，这里需要斟酌
func (Re *Reconfiguration) JudgeElement(TargetShard int, TargetNodeIndex int, Shard int, NodesInfo [][]Nodeinfo,index int) bool {
	//for i := 0; i < len(NodesInfo[Shard]); i++ {
	//	if Re.Nodesinfo[TargetShard][Re.FillArea[TargetShard][TargetNodeIndex]].PeerId == NodesInfo[Shard][i].PeerId{
	//		//fmt.Println("目标分片index",TargetNodeIndex)
	//		return false
	//	}
	//}
	data:=strings.Split(Re.Nodesinfo[TargetShard][TargetNodeIndex].NodeName,"S")
	shard,_:= strconv.Atoi(data[0])

	if shard == Shard{
		return false
	}
	NewPeerName := strconv.Itoa(Shard)+"S"+strconv.Itoa(index+1)
	if Re.Nodesinfo[TargetShard][TargetNodeIndex].NodeName == NewPeerName{
		fmt.Println("新节点",NewPeerName)//该位置需要移动
		fmt.Println("目标分片index1",TargetNodeIndex)
		return false
	}

	//for i:=0;i < len(Re.SendNodes[TargetShard]);i++{
	//	if Re.SendNodes[TargetShard][i].NodeName == NodesInfo[Shard][Re.FillArea[TargetShard][TargetNodeIndex]].NodeName{
	//		return false
	//	}
	//}
	return true
}
func (Re *Reconfiguration) ReFill(Shard int, NodesInfo [][]Nodeinfo,index int) (int, int) {
	var TargetShard int
	var TargetNodeIndex int
	flag := 0
	num_count:=0
	//选出目标节点进行更换
	for {
		if flag == 1 {
			return TargetShard, Re.FillArea[TargetShard][TargetNodeIndex]
		}
		time.Sleep(time.Millisecond * 5)
		rand.Seed(time.Now().UnixNano())
		TargetShard = rand.Intn(Re.ShardCount)
		if Shard == TargetShard{
			continue
		} else {

			count := 0
			for {
				num_count +=1
				if num_count%100==0{
					break
				}

				if count == len(Re.FillArea[TargetShard]) {
					break
				}
				time.Sleep(time.Millisecond * 5)
				rand.Seed(time.Now().UnixNano())
				TargetNodeIndex = rand.Intn(len(Re.FillArea[TargetShard]))
				if Re.JudgeElement(TargetShard, TargetNodeIndex, Shard, NodesInfo,index) {
					flag = 1
					break
				}

			}
		}
	}

}
func AddFullShard(FillArea [][]int) []int {
	var FullShard []int
	for i := 0; i < len(FillArea); i++ {

		if len(FillArea[i]) == 0 {
			FullShard = append(FullShard, i)
		}
	}

	return FullShard
}
func BelongCoordinate(FillArea []int) (int, int) {
	//返回两个参数：1. 要调整节点的参数。2. 互换节点的参数
	rand.Seed(time.Now().UnixNano())
	CoordinateIndex := rand.Intn(len(FillArea))//任选一个分片
	time.Sleep(time.Millisecond)
	return FillArea[CoordinateIndex], CoordinateIndex
}
func DeleteElement(Fillshard []int, CoordinateIndex int) []int {
	//删除集合元素
	Fillshard = append(Fillshard[:CoordinateIndex], Fillshard[CoordinateIndex+1:]...)
	return Fillshard
}

//删除被替换的元素
func (Re *Reconfiguration) DeleteExistElement(ShardIndex int, NodeIndex int,NodesInfo [][]Nodeinfo)string {
	for i := 0; i < len(Re.SendNodes[ShardIndex]); i++ {
		if Re.SendNodes[ShardIndex][i].PeerId == Re.Nodesinfo[ShardIndex][NodeIndex].PeerId {
			//fmt.Println("删除容器节点",Re.SendNodes[ShardIndex][i])
			ss:= Re.SendNodes[ShardIndex][i].NodeName
			//还原原来的值
			//Re.Nodesinfo[ShardIndex][NodeIndex].NodeName = NodesInfo[ShardIndex][NodeIndex].NodeName
			//Re.Nodesinfo[ShardIndex][NodeIndex].PeerId = NodesInfo[ShardIndex][NodeIndex].PeerId
			//fmt.Println("还原后的节点",Re.Nodesinfo[ShardIndex][NodeIndex])
			Re.SendNodes[ShardIndex] = append(Re.SendNodes[ShardIndex][:i], Re.SendNodes[ShardIndex][i+1:]...)
			return ss
		}
	}
	return ""
}
func (Re *Reconfiguration) FillReconfiguration() {
	Re.SendNodes = make([][]Nodeinfo,Re.ShardCount)
	var FillArea = make([][]int, len(Re.FillArea[:]))
	for i := 0; i < len(Re.FillArea); i++ {
		FillArea[i] = make([]int, len((Re.FillArea[i][:])))
		copy(FillArea[i], Re.FillArea[i][:])
	}
	var NodesInfo = make([][]Nodeinfo, len(Re.Nodesinfo[:]))
	//拿取4个分片信息节点,并做拷贝
	for i := 0; i < len(Re.Nodesinfo); i++ {
		NodesInfo[i] = make([]Nodeinfo, len((Re.Nodesinfo[i][:])))
		copy(NodesInfo[i], Re.Nodesinfo[i][:])
	}
	//填补区域，补充节点
	for i := 0; i < len(Re.FillArea); i++ {
		for j := 0; j < len(Re.FillArea[i]); j++ {
			//每个节点都要算自己属于哪个分片
			//由于随机选择分片，就存在有些分片很快就会被填满了，所以应该避免这种情况的发生。
			//每次检测哪些分片已经调整满了
			FullShard := AddFullShard(FillArea)
			BelongShard := Re.BelongShard(Re.ShardCount, FullShard, i)
			if BelongShard == -1 {//说明其他分片全部都被填满
				//避免哈希碰撞
				//得到替换节点的坐标
				//CoordinateIndex 表示？ BelongCoordinate 表示？
				BelongCoordinate, CoordinateIndex := BelongCoordinate(FillArea[i])//这个是获取还有哪个节点仍然没有被填入
				ShardIndex, NodeIndex := Re.ReFill(i, NodesInfo,BelongCoordinate)

				//得到分片坐标


				//删除集合元素
				FillArea[i] = DeleteElement(FillArea[i], CoordinateIndex)

				//删除被移动的元素
				name :=Re.DeleteExistElement(ShardIndex, NodeIndex,NodesInfo)
				if name ==""{
					continue
				}

				Re.Nodesinfo[i][BelongCoordinate].PeerId = Re.Nodesinfo[ShardIndex][NodeIndex].PeerId//会不会导致被覆盖？
				NodeChange := Nodeinfo{}
				NodeChange = Re.Nodesinfo[i][BelongCoordinate]
				NodeChange.NodeName = name
				//fmt.Println("碰撞节点1",NodeChange)//该节点为要调到原始位置的节点
				//NodeChange.Neighbor = Re.Nodesinfo[ShardIndex][NodeIndex].Neighbor
				Re.Nodesinfo[ShardIndex][NodeIndex].PeerId = NodesInfo[i][BelongCoordinate].PeerId
				NodeMove := Nodeinfo{}
				NodeMove = Re.Nodesinfo[ShardIndex][NodeIndex]
				NodeMove.NodeName = NodesInfo[i][Re.FillArea[i][j]].NodeName
				//fmt.Println("碰撞节点2",NodeMove)//该节点应该是原始位置要调到那个位置的节点
				if NodeChange.NodeName == NodeMove.NodeName{
					fmt.Println("错误")
				}
				Re.Count[i]++//说明该片节点以及有几个被完成调整
				//SendNodes主要是发布给调整服务
				//添加两个元素,被替换的节点和调整到替换节点的节点
				//节点碰撞
				ss,_:=strconv.Atoi(NodeMove.ShardName)
				for t:=0;i<len(Re.SendNodes[ss]);t++{
					if Re.SendNodes[ss][t].NodeName == NodeMove.NodeName || NodeChange.NodeName == Re.SendNodes[ss][t].NodeName{
						fmt.Println("存在重复行为")
					}
				}

				Re.SendNodes[i] = append(Re.SendNodes[i], NodeChange)
				Re.SendNodes[ShardIndex] = append(Re.SendNodes[ShardIndex], NodeMove)
			} else {
				//得到分片坐标

				BelongCoordinate, CoordinateIndex := BelongCoordinate(FillArea[BelongShard])
				//删除集合元素

				FillArea[BelongShard] = DeleteElement(FillArea[BelongShard], CoordinateIndex)
				//修改NodeInfo填写数据
				//fmt.Printf("origin shard= %d index=%d\n",i,Re.FillArea[i][j])
				Re.Nodesinfo[BelongShard][BelongCoordinate].PeerId = NodesInfo[i][Re.FillArea[i][j]].PeerId
				//开辟一个新的节点 节点的名字是旧名字
				//fmt.Printf("new shard= %d index=%d\n",BelongShard,BelongCoordinate)
				NodeMove := Nodeinfo{}

				NodeMove = Re.Nodesinfo[BelongShard][BelongCoordinate]
				NodeMove.NodeName = NodesInfo[i][Re.FillArea[i][j]].NodeName
				//NodeMove.Neighbor = NodesInfo[i][Re.FillArea[i][j]].Neighbor
				//fmt.Printf("new name= %s old name=%s\n",NodeMove.NodeName,Re.Nodesinfo[BelongShard][BelongCoordinate].NodeName)
				//SendNodes主要是发布给调整服务
				//fmt.Println("删除节点",NodeMove)
				ss,_:=strconv.Atoi(NodeMove.ShardName)
				for t:=0;t<len(Re.SendNodes[ss]);t++{

					if Re.SendNodes[ss][t].NodeName == NodeMove.NodeName{
						fmt.Println("存在重复行为")
					}
				}
				Re.SendNodes[BelongShard] = append(Re.SendNodes[BelongShard], NodeMove)

			}

		}
	}
	Re.Count = make([]int, Re.ShardCount)
	//Re.logger.Info("Done Fill")
}

func (Re *Reconfiguration) SendShard(i int) {

	for j := 0; j < len(Re.SendNodes[i]); j++ {
		//利用坐标，获取其邻居坐标
		Re.SendNode(i, j)

	}
}
func (Re *Reconfiguration)checkNodes(){
	for i:=0;i<len(Re.SendNodes);i++{
		//fmt.Println("移动的个数",len(Re.SendNodes[i]))
		for j:=0;j<len(Re.SendNodes[i]);j++{
			for k:=0;k<len(Re.SendNodes[i]);k++{

				if Re.SendNodes[i][j].NodeName==Re.SendNodes[i][k].NodeName && k !=j{
					fmt.Println(Re.SendNodes[i][j])
					fmt.Println(Re.SendNodes[i][k])
				}

			}

		}
	}

}
//循环遍历列表观察是否完全删除
func (Re *Reconfiguration) DeleteFlag() bool {
	for i := 0; i < len(Re.FlagSend); i++ {
		for j := 0; j < len(Re.FlagSend[i]); j++ {
			if Re.FlagSend[i][j] == false {
				return false
			}
		}
	}
	return true
}

//循环遍历列表观察是否完全创建
func (Re *Reconfiguration) CreateFlag() bool {
	for i := 0; i < len(Re.FlagCreate); i++ {
		for j := 0; j < len(Re.FlagCreate[i]); j++ {
			if Re.FlagCreate[i][j] == false {
				return false
			}
		}
	}
	return true
}

//执行调整方案
func (Re *Reconfiguration) SendToAdjust() {
	Re.FlagSend = make([][]bool, len(Re.SendNodes))
	Re.FlagCreate = make([][]bool, len(Re.SendNodes))
	//检测哪些节点还未被删除
	//初始化赋值都是false
	for i := 0; i < len(Re.FlagSend); i++ {
		Re.FlagSend[i] = make([]bool, len(Re.SendNodes[i]))
		Re.FlagCreate[i] = make([]bool, len(Re.SendNodes[i]))
		for j := 0; j < len(Re.FlagSend[i]); j++ {
			Re.FlagSend[i][j] = false
			Re.FlagCreate[i][j] = false
		}
	}

	Re.SendDelete()
	//检测所有节点容器是否被删除
	// for{
	// 	time.Sleep(time.Second)//睡眠1秒钟
	// 	//遍历节点是否都为ture删除
	// 	if Re.DeleteFlag()==false{//判断是否全部被删除了
	// 		Re.SendDelete()//继续删除
	// 	}else{
	// 		break//全部删除
	// 	}
	// }

	//Re.logger.Info("调整方案删除过程已经完成")

	Re.SendCreate() //创建节点，加入网络
	// for{
	// 	//检测所有节点是否完全创建
	// 	time.Sleep(time.Second)//睡眠1秒钟
	// 	//遍历节点是否都为ture创建
	// 	if Re.CreateFlag()==false{//判断是否全部被创建
	// 		Re.SendCreate()//继续创建
	// 	}else{
	// 		break//全部创建
	// 	}
	// }
	//Re.logger.Info("调整方案已经成功执行")
}
func (Re *Reconfiguration) SendCreate() {
	for i := 0; i < len(Re.SendNodes); i++ { //按分片创建
		Re.SendShard(i)
	}
}
func (Re *Reconfiguration) SendDelete() {
	for i := 0; i < len(Re.SendNodes); i++ {
		Re.SendDeleteShard(i)

	}
}
func (Re *Reconfiguration) SendDeleteShard(i int) {
	for j := 0; j < len(Re.SendNodes[i]); j++ {
		//判断该容器是否被删除，如果被删除则跳过
		if Re.FlagSend[i][j] == true {
			continue
		} else {
			Re.DeleteNode(i, j)

		}
	}
}
func (Re *Reconfiguration) DeleteNode(i int, j int) {
	//Re.SendNodes[i][j].NodeName
	//fmt.Println("删除容器",OldPeerName)
	//deleteurl := "http://10.77.70.135:9001/api/v2/tendermint/delete-peer"
	//result := DeletePost(deleteurl, OldPeerName)
	//修改删除列表的bool值
	if true {
		Re.FlagSend[i][j] = true
	}

}

//调用接口完善,还需要记录是否以及起成功的信息
func (Re *Reconfiguration) SendNode(i int, j int) {
	//每个节点进行发送服务
	Neighborstr := ""
	Gensisstr := ""
	//OldPeerName := ""
	Shard, _ := strconv.Atoi(Re.SendNodes[i][j].ShardName)
	for k := 0; k < len(Re.SendNodes[i][j].Neighbor); k++ {

		if Re.SendNodes[i][j].Neighbor[k] > Re.NodeCount-1 {
			continue
		}
		Node := Re.Nodesinfo[Shard][Re.SendNodes[i][j].Neighbor[k]]
		SubStr := Node.PeerId + "@" + Node.NodeName + ":26656"

		if k == len(Re.SendNodes[i][j].Neighbor)-1 {
			Neighborstr = Neighborstr + SubStr
		} else {
			Neighborstr = Neighborstr + SubStr + ","
		}
	}
	//fmt.Println("新容器的邻居",Neighborstr)
	//fmt.Println("Done")
	Gensisstr = Re.Chaininfo[Shard]
	if Gensisstr != "" {
		Gensisstr = Gensisstr + ""
	}
	//OldPeerName = Re.SendNodes[i][j].NodeName //旧节点的名字

	Coordinate, _ := strconv.Atoi(Re.SendNodes[i][j].Coordinate)
	Coordinate = Coordinate + 1
	//Shard1, _ := strconv.Atoi(Re.SendNodes[i][j].ShardName)

	//NewPeerName := Re.SendNodes[i][j].ShardName + "S" + strconv.Itoa(Coordinate)
	//fmt.Println("旧容器名字：", OldPeerName,"旧容器的id",Re.SendNodes[i][j].PeerId, "新容器名字", NewPeerName)
	//fmt.Println("新容器名",NewPeerName,"新容器的邻居",Neighborstr)
	//moveurl := "http://10.77.70.135:9001/api/v2/tendermint/create-peer"
	//result := MovePost(moveurl, OldPeerName, NewPeerName, Neighborstr, Gensisstr)
	if true {
		Re.FlagSend[i][j] = true
	}
}
type createInfo struct {
	PeerName    string `json:"peerName"`
	NewPeerName string `json:"newPeerName"`
	Neighbors   string `json:"neighbors"`
	Genesis     string `json:"genesis"`
}
type deleteinfo struct {
	PeerName string `json:"peerName"`
}
type returnmessage struct {
	Flag bool `json:"flag"`

	Data    string `json:"data"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func DeletePost(deleteurl string, OldPeerName string) bool {

	deleteurl = deleteurl + "?peerName=" + OldPeerName
	fmt.Println(deleteurl)
	req, err := http.Get(deleteurl)
	if err != nil {

		return false
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		// handle error
		return false
	}
	message := returnmessage{}
	json.Unmarshal(body, &message)
	//将返回信息打印出来观察
	fmt.Println("返回信息：", message)
	if message.Code == 200 {
		return true
	} else {
		return false
	}

}

func MovePost(url1 string, OldPeerName string, NewPeerName string, Neighborstr string, Gensisstr string) bool {
	gensis := Gensisstr[1 : len(Gensisstr)-1]
	// fmt.Println(gensis)
	loginInfo := createInfo{

		PeerName:    OldPeerName,
		NewPeerName: NewPeerName,
		Neighbors:   Neighborstr,
		Genesis:     gensis,
	}
	contentType := "application/json"
	info, _ := json.Marshal(loginInfo)
	req, err := http.Post(url1, contentType, strings.NewReader(string(info)))
	// fmt.Println("string:",strings.NewReader(string(info)))
	if err != nil {

		return false
	}
	//  http.PostForm(url1,
	// 	url.Values{"peerName":{PeerName}}
	// 	)
	if err != nil {

		return false
		// handle error
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		// handle error
	}

	message := returnmessage{}
	json.Unmarshal(body, &message)
	//将返回信息打印出来观察
	fmt.Println("返回信息：", message)
	if message.Code == 200 {
		return true
	} else {
		return false
	}

}
func GetCredibleTimeStamp()time.Time{
	CredibleTimeStamp:=time.Now()
	return CredibleTimeStamp
}
func GetCredibleRand(ShardCount int)int{
	r:=rand.New(rand.NewSource(time.Now().UnixNano()))
	time.Sleep(time.Millisecond)
	rnd:=r.Intn(ShardCount)
	return rnd
}