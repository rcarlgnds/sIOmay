package helpers

import (
	"sync"

	hook "github.com/robotn/gohook"
)

type Coordinates struct {
	X, Y int
}

type MouseWheel struct {
	Event     string
	Amount    int
	Rotation  int
	Direction int
}

type Mouse struct {
	Current   Coordinates
	Last      Coordinates
	Button    int
	Clicks    int
	Wheel     *MouseWheel
	Dragging  bool
	DragStart Coordinates
	DragEnd   Coordinates
	Mu        sync.Mutex
}

func NewMouse() *Mouse {
	return &Mouse{}
}

func (mouse *Mouse) ListenForMouseEvents() {
	mouse.ListenForMouseEventsWithCallback(nil)
}

func (mouse *Mouse) ListenForMouseEventsWithCallback(callback func()) {
	go func() {
		evChan := hook.Start()
		defer hook.End()
		for ev := range evChan {
			mouse.Mu.Lock()
			mouse.Last = mouse.Current
			mouse.Current = Coordinates{X: int(ev.X), Y: int(ev.Y)}
			mouse.Button = int(ev.Button)
			mouse.Clicks = int(ev.Clicks)

			switch ev.Kind {
			case hook.MouseDrag:
				if !mouse.Dragging {
					mouse.Dragging = true
					mouse.DragStart = Coordinates{X: int(ev.X), Y: int(ev.Y)}
				}
				mouse.DragEnd = Coordinates{X: int(ev.X), Y: int(ev.Y)}
			case hook.MouseUp:
				mouse.Clicks = 0
				mouse.Button = 0
				if mouse.Dragging {
					mouse.Dragging = false
				}
			}
			if ev.Kind == hook.MouseWheel {
				mouse.Wheel = &MouseWheel{
					Event:     determineWheelEvent(int(ev.Rotation)),
					Amount:    int(ev.Amount),
					Rotation:  int(ev.Rotation),
					Direction: int(ev.Direction),
				}
			} else {
				mouse.Wheel = nil
			}
			mouse.Mu.Unlock()
			
			// Call callback if provided
			if callback != nil {
				callback()
			}
		}
	}()
}

func determineWheelEvent(rotation int) string {
	if rotation < 0 {
		return "scroll_up"
	} else if rotation > 0 {
		return "scroll_down"
	}
	return "scroll_unknown"
}

func (mouse *Mouse) HasMouseChanged() bool {
	mouse.Mu.Lock()
	defer mouse.Mu.Unlock()

	if mouse.Current != mouse.Last ||
		(mouse.Wheel != nil && (mouse.Wheel.Amount != 0 || mouse.Wheel.Rotation != 0)) {
		return true
	}

	return false
}
