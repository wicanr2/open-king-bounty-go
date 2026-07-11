package screen

// music.go — 各畫面宣告 BGM 音樂場景(對齊 C bgm.c 的 scenes.ini key;由 app 每幀切曲)。
// 實作 bgm.Scener(MusicScene() string)即宣告場景;未列出的畫面(選單/檢視/戰後結算等)
// 不實作此方法 → 沿用目前音樂,不打斷。場景 key 見 data/music/scenes.ini。

// MusicScene:世界地圖 → 野外主題。
func (s *WorldMapScreen) MusicScene() string { return "field1" }

// MusicScene:戰鬥 → 戰鬥主題。
func (s *CombatScreen) MusicScene() string { return "combat" }

// MusicScene:城鎮 → 城鎮主題。
func (s *TownScreen) MusicScene() string { return "town" }

// MusicScene:家鄉城堡 → 城堡主題。
func (s *CastleHomeScreen) MusicScene() string { return "castle" }

// MusicScene:自家城堡 → 城堡主題。
func (s *CastleOwnScreen) MusicScene() string { return "castle" }

// MusicScene:敵方城堡圍攻 → 圍攻主題。
func (s *CastleSiegeScreen) MusicScene() string { return "siege" }

// MusicScene:標題 → 標題主題。
func (s *TitleScreen) MusicScene() string { return "title" }

// MusicScene:開場 logo → 標題主題(沿用到標題,不中斷)。
func (s *LogoScreen) MusicScene() string { return "title" }

// MusicScene:勝利結局 → 勝利主題。
func (s *WinScreen) MusicScene() string { return "win" }
