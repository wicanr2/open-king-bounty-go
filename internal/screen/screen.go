// Package screen 是畫面流程狀態機:把 input.Action 餵給當前畫面,畫面回報是否轉場,
// 由 Manager 以堆疊管理(title → 選角 → 遊戲中 → 戰鬥…)。畫面只認 Action(不碰 raw key),
// 觸控/鍵盤共用同一套輸入(見 internal/input)。
package screen

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/input"
)

// Screen 是一個遊戲畫面。消費一個 Action、畫自己、宣告當下接受哪些 Action(給觸控層)。
type Screen interface {
	Update(a input.Action) Transition
	Draw(dst *ebiten.Image)
	Keymap() input.Keymap
}

// TransitionKind 是 Update 後的畫面轉場種類。
type TransitionKind uint8

const (
	KindStay    TransitionKind = iota // 留在本畫面
	KindReplace                       // 換成 Next(title→選角)
	KindPush                          // 疊上 Next(進戰鬥,結束後回原畫面)
	KindPop                           // 回上一個
)

// Transition 描述轉場。用建構子 Stay()/Replace()/Push()/Pop() 產生。
type Transition struct {
	Kind TransitionKind
	Next Screen
}

func Stay() Transition            { return Transition{Kind: KindStay} }
func Replace(s Screen) Transition { return Transition{Kind: KindReplace, Next: s} }
func Push(s Screen) Transition    { return Transition{Kind: KindPush, Next: s} }
func Pop() Transition             { return Transition{Kind: KindPop} }

// Manager 以堆疊管理畫面,最上層為當前畫面。
type Manager struct {
	stack []Screen
}

// NewManager 以 root 為起始畫面。
func NewManager(root Screen) *Manager {
	return &Manager{stack: []Screen{root}}
}

// Current 回傳當前(最上層)畫面;空堆疊回 nil。
func (m *Manager) Current() Screen {
	if len(m.stack) == 0 {
		return nil
	}
	return m.stack[len(m.stack)-1]
}

// Update 把 Action 交給當前畫面並套用其轉場。
func (m *Manager) Update(a input.Action) {
	cur := m.Current()
	if cur == nil {
		return
	}
	tr := cur.Update(a)
	switch tr.Kind {
	case KindReplace:
		if tr.Next != nil {
			m.stack[len(m.stack)-1] = tr.Next
		}
	case KindPush:
		if tr.Next != nil {
			m.stack = append(m.stack, tr.Next)
		}
	case KindPop:
		if len(m.stack) > 1 {
			m.stack = m.stack[:len(m.stack)-1]
		}
	}
}

// Draw 畫當前畫面。
func (m *Manager) Draw(dst *ebiten.Image) {
	if cur := m.Current(); cur != nil {
		cur.Draw(dst)
	}
}

// Keymap 回傳當前畫面的 Keymap(給觸控層渲染控制)。
func (m *Manager) Keymap() input.Keymap {
	if cur := m.Current(); cur != nil {
		return cur.Keymap()
	}
	return input.Keymap{}
}
