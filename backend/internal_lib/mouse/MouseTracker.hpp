#pragma once
#include "Mouse.hpp"
#include "windows.h"

void PollMouseWindows(MouseCapture& cap);

extern HHOOK hHook;
LRESULT CALLBACK MouseProc(int nCode, WPARAM wParam, LPARAM lParam);
void MessagePump();

void startHook();

void WinApplyMouseState(const MouseState& state, MouseState &prevMouseState);