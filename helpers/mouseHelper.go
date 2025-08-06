package helpers

import (
	"sIOmay/core"
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
	ByteData  *core.Bytedata
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
			mouse.processMouseEvent(ev)
			
			if callback != nil {
				callback()
			}
		}
	}()
}

func (mouse *Mouse) processMouseEvent(ev hook.Event) {
	mouse.Mu.Lock()
	defer mouse.Mu.Unlock()
	
	mouse.updateMousePosition(ev)
	
	mouse.ByteData = core.NewBytedata()
	
	mouse.handleMouseEventType(ev)
	
	mouse.updateWheelData(ev)
}

func (mouse *Mouse) updateMousePosition(ev hook.Event) {
	mouse.Last = mouse.Current
	mouse.Current = Coordinates{X: int(ev.X), Y: int(ev.Y)}
	mouse.Button = int(ev.Button)
	mouse.Clicks = int(ev.Clicks)
}

func (mouse *Mouse) handleMouseEventType(ev hook.Event) {
	switch ev.Kind {
	case hook.MouseMove:
		mouse.ByteData.MouseMove(int16(ev.X), int16(ev.Y))
		
	case hook.MouseDown:
		mouse.handleMouseClick(ev)
		// Also include position for click events
		mouse.ByteData.MouseMove(int16(ev.X), int16(ev.Y))
		
	case hook.MouseDrag:
		mouse.handleMouseDrag(ev)
		
	case hook.MouseUp:
		mouse.handleMouseUp(ev)
		
	case hook.MouseWheel:
		mouse.ByteData.MouseScroll(int16(ev.Rotation))
		// Include position for scroll events
		mouse.ByteData.MouseMove(int16(ev.X), int16(ev.Y))
	}
}

func (mouse *Mouse) handleMouseClick(ev hook.Event) {
	switch ev.Button {
	case 1: // Left button
		mouse.ByteData.MouseClickLeft()
	case 2: // Right button  
		mouse.ByteData.MouseClickRight()
	case 3: // Middle button
		mouse.ByteData.MouseMiddleClick()
	}
}

func (mouse *Mouse) handleMouseDrag(ev hook.Event) {
	mouse.ByteData.MouseMove(int16(ev.X), int16(ev.Y))
	if !mouse.Dragging {
		mouse.Dragging = true
		mouse.DragStart = Coordinates{X: int(ev.X), Y: int(ev.Y)}
	}
	mouse.DragEnd = Coordinates{X: int(ev.X), Y: int(ev.Y)}
}

func (mouse *Mouse) handleMouseUp(ev hook.Event) {
	// Include position for mouse up events
	mouse.ByteData.MouseMove(int16(ev.X), int16(ev.Y))
	
	if mouse.Dragging {
		mouse.Dragging = false
	}
}

func (mouse *Mouse) updateWheelData(ev hook.Event) {
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

// GetByteData returns the current mouse event as byte data
func (mouse *Mouse) GetByteData() []byte {
	mouse.Mu.Lock()
	defer mouse.Mu.Unlock()
	
	if mouse.ByteData != nil {
		return mouse.ByteData.Bytes()
	}
	return nil
}

// GetCurrentByteData returns a copy of the current ByteData
func (mouse *Mouse) GetCurrentByteData() *core.Bytedata {
	mouse.Mu.Lock()
	defer mouse.Mu.Unlock()
	
	return mouse.ByteData
}

// HasNewByteData checks if there's new byte data available
func (mouse *Mouse) HasNewByteData() bool {
	mouse.Mu.Lock()
	defer mouse.Mu.Unlock()
	
	return mouse.ByteData != nil
}

// ClearByteData clears the current byte data
func (mouse *Mouse) ClearByteData() {
	mouse.Mu.Lock()
	defer mouse.Mu.Unlock()
	
	mouse.ByteData = nil
}