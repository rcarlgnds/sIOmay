package helpers

import (
	"sIOmay/core"
	"testing"
)

func TestMouseByteDataGeneration(t *testing.T) {
	mouse := NewMouse()
	
	// Test that ByteData is initially nil
	if mouse.HasNewByteData() {
		t.Error("Expected no byte data initially")
	}
	
	// Simulate creating byte data manually
	byteData := core.NewBytedata()
	byteData.MouseMove(100, 200)
	mouse.ByteData = byteData
	
	// Test that we can retrieve byte data
	if !mouse.HasNewByteData() {
		t.Error("Expected byte data to be available")
	}
	
	data := mouse.GetByteData()
	if len(data) != 7 {
		t.Errorf("Expected 7 bytes, got %d", len(data))
	}
	
	// Test that we can decode the position
	x, y := byteData.GetMousePosition()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%d, %d)", x, y)
	}
	
	// Test clearing byte data
	mouse.ClearByteData()
	if mouse.HasNewByteData() {
		t.Error("Expected no byte data after clearing")
	}
}

func TestMouseClickByteData(t *testing.T) {
	byteData := core.NewBytedata()
	
	// Test left click
	byteData.MouseClickLeft()
	if !byteData.HasMouseClickLeft() {
		t.Error("Expected left click to be detected")
	}
	
	// Test right click
	byteData.Clear()
	byteData.MouseClickRight()
	if !byteData.HasMouseClickRight() {
		t.Error("Expected right click to be detected")
	}
	
	// Test scroll
	byteData.Clear()
	byteData.MouseScroll(-3)
	if !byteData.HasMouseScroll() {
		t.Error("Expected scroll to be detected")
	}
	rotation := byteData.GetScrollRotation()
	if rotation != -3 {
		t.Errorf("Expected rotation -3, got %d", rotation)
	}
}
