package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// 神器能力位元,對齊 C bounty.h:311-318 POWER_*(值 = artifactPowerFlags map 的位元)。
const (
	PowerUnknown           = 0x01
	PowerQuarterProtection = 0x02
	PowerDoubleLeadership  = 0x04
	PowerDoubleSpellPower  = 0x08
	PowerIncreaseCommission = 0x10
	PowerCheaperBoatRental = 0x20
	PowerDoubleMaxSpells   = 0x40
	PowerIncreasedDamage   = 0x80
)

// artifactAtOrdered 回傳「地圖排序位置 ordered(= continent*2 + num)」上的神器 id,
// 對齊 C artifact_inversion[ordered]:找 InvertID == ordered 的神器(artifacts.ini 的
// invert_id 正是此神器出現在地圖上的排序位置)。找不到回 -1。
func artifactAtOrdered(arts []Artifact, ordered int) int {
	for i := range arts {
		if arts[i].InvertID == ordered {
			return arts[i].ID
		}
	}
	return -1
}

// HasPower 回報玩家是否已擁有帶指定能力位元的神器(對齊 C has_power:掃 artifact_found[]
// 看有沒有哪個已拾得的神器帶該 power)。
func (gs *GameState) HasPower(a *kbdata.Assets, power int) bool {
	arts := LoadArtifacts(a)
	for i := range arts {
		if arts[i].ID < len(gs.ArtifactFound) && gs.ArtifactFound[arts[i].ID] != 0 && arts[i].PowerFlag()&power != 0 {
			return true
		}
	}
	return false
}

// ArtifactPowers 回傳玩家所有已拾得神器能力位元的 OR 總和,對齊 C prepare_units_player
// (play.c:1138-1141)的 `war->powers[side] |= artifact_powers[i]`。戰鬥佈陣時填入
// combat.Powers,讓 INCREASED_DAMAGE/QUARTER_PROTECTION 等生效。
func (gs *GameState) ArtifactPowers(a *kbdata.Assets) int {
	arts := LoadArtifacts(a)
	power := 0
	for i := range arts {
		if arts[i].ID < len(gs.ArtifactFound) && gs.ArtifactFound[arts[i].ID] != 0 {
			power |= arts[i].PowerFlag()
		}
	}
	return power
}

// ArtifactPickup 描述拾取神器的結果(供 UI 顯示名稱/描述)。
type ArtifactPickup struct {
	Found bool
	Name  string
	Desc  string
}

// TakeArtifact 對齊 C take_artifact(game.c:3332):玩家踩到本洲第 num(0/1)個神器格
// (num = tile - TileArtifact1),標記已拾得、套用「拾取即生效」的能力(領導力/法力/
// 俸祿/法術上限加倍或增額),清掉該格 tile,回傳名稱與描述供顯示。
func (gs *GameState) TakeArtifact(a *kbdata.Assets, num int) ArtifactPickup {
	arts := LoadArtifacts(a)
	ordered := gs.Continent*2 + num
	id := artifactAtOrdered(arts, ordered)
	if id < 0 || id >= len(arts) {
		return ArtifactPickup{}
	}
	art := arts[id]
	if id < len(gs.ArtifactFound) {
		gs.ArtifactFound[id] = 1
	}
	// 拾取即生效的能力(對齊 C take_artifact 的 instant effects,game.c:3345-3357)。
	power := art.PowerFlag()
	if power&PowerDoubleLeadership != 0 {
		gs.BaseLeadership *= 2
		gs.Leadership = gs.BaseLeadership
	}
	if power&PowerDoubleSpellPower != 0 {
		gs.SpellPower *= 2
	}
	if power&PowerIncreaseCommission != 0 {
		gs.Commission += 2000
	}
	if power&PowerDoubleMaxSpells != 0 {
		gs.MaxSpells *= 2
	}
	if gs.WorldMap != nil {
		gs.WorldMap.Set(gs.Continent, gs.X, gs.Y, 0)
	}
	return ArtifactPickup{Found: true, Name: art.Name, Desc: art.Description}
}
