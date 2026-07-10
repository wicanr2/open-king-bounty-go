package gamestate

// PuzzleW/PuzzleH 對齊 C PUZZLEMAP_W/H(bounty.h:75)。
const (
	PuzzleW = 5
	PuzzleH = 5
)

// puzzleMap 對齊 C puzzle_map[5][5](bounty.c):每格由一件神器(負值,artifact_id=-id-1)
// 或一名惡棍(非負,villain id)遮住;找到該神器/捕獲該惡棍即掀開該格,露出權杖周邊
// (5×5,中心 = 權杖所在)的實際地圖。全 8 神器 + 17 惡棍 = 25 格。
var puzzleMap = [PuzzleH][PuzzleW]int{
	{-1, 7, -2, 6, -3},
	{5, 15, 14, 13, 4},
	{-4, 12, 16, 11, -5},
	{3, 10, 9, 8, 2},
	{-6, 1, -7, 0, -8},
}

// PuzzleOpened 回報拼圖第 (x,y) 格是否已掀開(對齊 C view_puzzle 的 opened 判定):
// 神器格看 ArtifactFound、惡棍格看 VillainCaught。
func (gs *GameState) PuzzleOpened(x, y int) bool {
	if x < 0 || x >= PuzzleW || y < 0 || y >= PuzzleH {
		return false
	}
	id := puzzleMap[y][x]
	if id < 0 {
		artID := -id - 1
		return artID >= 0 && artID < len(gs.ArtifactFound) && gs.ArtifactFound[artID] != 0
	}
	return id >= 0 && id < len(gs.VillainCaught) && gs.VillainCaught[id] != 0
}

// PuzzleTile 回傳拼圖第 (x,y) 格對應的地圖 tile(權杖周邊 5×5,中心為權杖):
// 地圖座標 = (scepter_x - PuzzleW/2 + x, scepter_y - PuzzleH/2 + y)。掀開後才有意義。
func (gs *GameState) PuzzleTile(x, y int) byte {
	if gs.WorldMap == nil {
		return 0
	}
	mx := gs.ScepterX - PuzzleW/2 + x
	my := gs.ScepterY - PuzzleH/2 + y
	return gs.WorldMap.Tile(gs.ScepterContinent, mx, my)
}
