package combat

import (
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// 神器加成位元(對應 C bounty.h POWER_*),傷害相關兩個。
const (
	PowerQuarterProtection = 0x02 // 受方:傷害 /4*3
	PowerIncreasedDamage   = 0x80 // 攻方:傷害 *1.5
)

// unitsKilled / damageRemainder 對應 C play.c:1380 / 1384。
func unitsKilled(damage, hp int) int     { return damage / hp }
func damageRemainder(damage, hp int) int { return damage % hp }

// DealDamage 是 deal_damage(play.c:1433)的忠實移植(unit-vs-unit 與 magic-vs-unit)。
// 回傳整隻擊殺數(免疫魔法回 -1 取消攻擊)。rng 注入以對 seed parity;傷害骰用
// KB_rand(melee/ranged min,max)。參數對齊 C:isRanged/isExternal/externalDamage/retaliation。
//
// ⚠ 士氣倍率需 troop_morale(尚未移植)→ 標 TODO;僅在該側有 Hero 且 under_control 時生效,
// 對敵方(Heroes[side]==nil)無影響,故不影響目前無 Hero 的戰鬥與黃金樣本。
func (c *Combat) DealDamage(a *kbdata.Assets, rng kbrng.Rand, aSide, aID, tSide, tID int, isRanged, isExternal bool, externalDamage int, retaliation bool) int {
	u := &c.Units[aSide][aID] // 攻擊單位
	t := &c.Units[tSide][tID] // 目標單位

	demonKills := 0
	finalDamage := 0

	if isExternal {
		finalDamage = externalDamage
	} else {
		atk := a.Troops[u.TroopID]
		tgt := a.Troops[t.TroopID]

		// 死神鐮(惡魔):10% 機率斬殺半數(向上取整)
		if atk.Abilities&kbdata.AbilScythe != 0 {
			if rng.Between(1, 100) > 89 {
				demonKills = t.Count/2 + t.Count%2
			}
		}

		var dmg int
		if isRanged {
			retaliation = true
			u.Shots--
			if atk.Abilities&kbdata.AbilMagic != 0 {
				if tgt.Abilities&kbdata.AbilImmune != 0 {
					return -1 // 免疫魔法 → 取消攻擊
				}
				dmg = atk.RangedMin // 德魯依 10 / 大法師 25(固定)
			} else {
				dmg = rng.Between(atk.RangedMin, atk.RangedMax)
			}
		} else {
			dmg = rng.Between(atk.MeleeMin, atk.MeleeMax)
		}

		total := dmg * u.TurnCount
		skillDiff := atk.SkillLevel + 5 - tgt.SkillLevel
		finalDamage = (total * skillDiff) / 10

		// 士氣倍率(low:/2, high:*1.5),需 Hero 掌控且 troop_morale
		if c.Heroes[aSide] != nil {
			// TODO(flagship): 移植 troop_morale + unit_under_control → 套用士氣倍率
		}

		if c.Powers[aSide]&PowerIncreasedDamage != 0 {
			finalDamage += finalDamage / 2 // *1.5
		}
		if c.Powers[tSide]&PowerQuarterProtection != 0 {
			finalDamage /= 4
			finalDamage *= 3 // 近似 *0.75(先 /4 可能歸 0,較狠)
		}
	}

	tgtHP := a.Troops[t.TroopID].HP
	finalDamage += t.Injury               // 累積舊傷
	finalDamage += tgtHP * demonKills      // 惡魔斬殺

	kills := unitsKilled(finalDamage, tgtHP)
	t.Injury = damageRemainder(finalDamage, tgtHP)

	if kills < t.Count {
		t.Count -= kills
	} else {
		// 整堆殲滅
		t.Dead = true
		t.Count = 0
		kills = t.TurnCount
		finalDamage = kills * tgtHP
	}

	// TODO(flagship): Heroes[tSide].followers_killed += kills(gamestate 尚無該欄)

	if !isExternal {
		atk := a.Troops[u.TroopID]
		if atk.Abilities&kbdata.AbilAbsorb != 0 { // 鬼:吸收擊殺數補己方
			u.Count += kills
		}
		if atk.Abilities&kbdata.AbilLeech != 0 { // 吸血鬼:依傷害吸血,上限 max_count
			u.Count += unitsKilled(finalDamage, atk.HP)
			if u.Count > u.MaxCount {
				u.Count = u.MaxCount
				u.Injury = 0
			}
		}
		// 反擊(僅非反擊、目標未反擊過):目標回敬一次
		if !retaliation && !t.Retaliated {
			t.Retaliated = true
			c.DealDamage(a, rng, tSide, tID, aSide, aID, false, false, 0, true)
		}
	}

	return kills
}

// UnitHitUnit 近戰攻擊,對應 C unit_hit_unit(play.c:1579):deal_damage melee, retaliation=1。
func (c *Combat) UnitHitUnit(a *kbdata.Assets, rng kbrng.Rand, side, id, otherSide, otherID int) int {
	return c.DealDamage(a, rng, side, id, otherSide, otherID, false, false, 0, true)
}

// UnitRangedShot 遠攻,對應 C unit_ranged_shot(play.c:1590):deal_damage ranged, retaliation=0。
func (c *Combat) UnitRangedShot(a *kbdata.Assets, rng kbrng.Rand, side, id, otherSide, otherID int) int {
	return c.DealDamage(a, rng, side, id, otherSide, otherID, true, false, 0, false)
}
