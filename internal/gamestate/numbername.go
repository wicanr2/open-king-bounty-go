package gamestate

// number_name 對齊 C bounty.c:512-527:把兵力數量轉成模糊的量詞描述
// (gather_information 顯示敵方駐軍時用,不給精確數字)。
var numberNames = [6]string{
	"無數的", // 0  500+
	"一大群", // 1  100-499
	"大量的", // 2  50-99
	"許多",  // 3  20-49
	"一些",  // 4  10-19
	"少數",  // 5  1-9
}
var numberMins = [6]int{500, 100, 50, 20, 10, 1}

// NumberName 回傳數量 num 對應的量詞,逐句對齊 C number_name(bounty.c:522):
// 找第一個 num >= number_mins[i] 的 i 回 numberNames[i];都不滿足(num<10)回
// numberNames[5]=「少數」。
func NumberName(num int) string {
	for i := 0; i < 5; i++ {
		if num >= numberMins[i] {
			return numberNames[i]
		}
	}
	return numberNames[5]
}
