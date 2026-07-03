package screen

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/input"
)

// fakeScreen 是測試用畫面:記下收到的 Action,並回傳預設轉場。
type fakeScreen struct {
	name string
	next Transition
	got  []input.Action
}

func (f *fakeScreen) Update(a input.Action) Transition {
	f.got = append(f.got, a)
	return f.next
}
func (f *fakeScreen) Draw(*ebiten.Image)   {}
func (f *fakeScreen) Keymap() input.Keymap { return input.Keymap{} }

func TestManager_Replace(t *testing.T) {
	b := &fakeScreen{name: "b", next: Stay()}
	a := &fakeScreen{name: "a", next: Replace(b)}
	m := NewManager(a)

	m.Update(input.Action{Kind: input.ActConfirm})
	if m.Current() != b {
		t.Fatal("Replace 後當前畫面應為 b")
	}
}

func TestManager_PushPop(t *testing.T) {
	root := &fakeScreen{name: "root"}
	pushed := &fakeScreen{name: "pushed", next: Pop()}
	root.next = Push(pushed)
	m := NewManager(root)

	m.Update(input.None) // root push pushed
	if m.Current() != pushed {
		t.Fatal("Push 後當前應為 pushed")
	}
	m.Update(input.None) // pushed pop → 回 root
	if m.Current() != root {
		t.Fatal("Pop 後當前應回 root")
	}
}

func TestManager_PopKeepsRoot(t *testing.T) {
	root := &fakeScreen{name: "root", next: Pop()}
	m := NewManager(root)
	m.Update(input.None) // 已在 root,Pop 不該清空
	if m.Current() != root {
		t.Fatal("Pop 不該把 root 也彈掉")
	}
}
