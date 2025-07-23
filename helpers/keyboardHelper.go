package helpers

import (
	"fmt"
	"github.com/eiannone/keyboard"
	hook "github.com/robotn/gohook"
	"strconv"
	"sync"
)

type Keyboard struct {
	isRunning bool
	Char      string
	Key       string
	Shift     bool
	Ctrl      bool
	Alt       bool
	Mu        sync.Mutex
}

func NewKeyboardEvent() *Keyboard {
	return &Keyboard{}
}

func (k *Keyboard) ListenForGlobalKeyboardEvents(events chan<- *Keyboard) error {
	fmt.Println("Starting global keyboard listener...")
	go func() {
		for ev := range hook.Start() {
			if ev.Kind == hook.KeyDown || ev.Kind == hook.KeyUp {
				isPressed := ev.Kind == hook.KeyDown
				k.Mu.Lock()
				switch ev.Rawcode {
				case 42, 54:
					k.Shift = isPressed
				case 29, 157:
					k.Ctrl = isPressed
				case 56, 184:
					k.Alt = isPressed
				default:
					k.Char = string(rune(ev.Rawcode))
					if k.Char == "" {
						k.Char = "Unknown"
					}
					k.Key = strconv.Itoa(int(ev.Rawcode))
				}
				k.Mu.Unlock()
				events <- &Keyboard{
					Char:  k.Char,
					Key:   k.Key,
					Shift: k.Shift,
					Ctrl:  k.Ctrl,
					Alt:   k.Alt,
				}
			}
		}
	}()

	return nil
}

func (k *Keyboard) ListenForKeyboardEvents(events chan<- *Keyboard) {
	err := keyboard.Open()
	if err != nil {
		fmt.Println("Error opening keyboard events: ", err)
		return
	}
	defer func() {
		keyboard.Close()
		fmt.Println("Keyboard closed.")
	}()

	k.isRunning = true

	for k.isRunning {
		// Read the next key event
		char, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Println("Error reading keyboard: ", err)
			return
		}

		k.Mu.Lock()
		k.Char = string(char)
		k.Key = strconv.Itoa(int(key))
		k.Mu.Unlock()

		// Handle special keys and modifier keys (Shift, Ctrl, Alt)
		switch key {
		//case keyboard.KeyShift:
		//	k.Shift = true
		//	fmt.Println("Shift key pressed")
		//case keyboard.KeyCtrl:
		//	k.Ctrl = true
		//	fmt.Println("Ctrl key pressed")
		//case keyboard.KeyAlt:
		//	k.Alt = true
		//	fmt.Println("Alt key pressed")
		case keyboard.KeyTab:
			fmt.Println("Tab key pressed")
		case keyboard.KeyEsc:
			fmt.Println("Escape key pressed")
		case keyboard.KeyEnter:
			fmt.Println("Enter key pressed")
		case keyboard.KeySpace:
			fmt.Println("Spacebar pressed")
		case keyboard.KeyBackspace:
			fmt.Println("Backspace pressed")
			//case keyboard.KeyUp:
			//	fmt.Println("Up Arrow pressed")
			//case keyboard.KeyDown:
			//	fmt.Println("Down Arrow pressed")
			//case keyboard.KeyLeft:
			//	fmt.Println("Left Arrow pressed")
			//case keyboard.KeyRight:
			//	fmt.Println("Right Arrow pressed")

			// Reset modifier keys when they are released (optional)
			//if key == keyboard.KeyShift {
			//	k.Shift = false
			//}
			//if key == keyboard.KeyCtrl {
			//	k.Ctrl = false
			//}
			//if key == keyboard.KeyAlt {
			//	k.Alt = false
			//}
		}

		events <- &Keyboard{
			Char: k.Char,
			Key:  k.Key,
		}
	}

}
