#include <iostream>
#include <thread>
#include <chrono>
#include <windows.h>
#include "Mouse.hpp"
#include "MouseTracker.hpp"
#include <atomic>

#define SLEEP_DURATION 20 // ms

HHOOK hHook = NULL;
// Global Hool Data
std::atomic<int> scrollDelta(0);
//Windows Mouse Proc 
LRESULT CALLBACK MouseProc(int nCode, WPARAM wParam, LPARAM lParam)
{
    if (nCode == HC_ACTION && wParam == WM_MOUSEWHEEL)
    {
        MSLLHOOKSTRUCT* pMouse = (MSLLHOOKSTRUCT*)lParam;

        // Extract wheel delta from mouseData (HIWORD)
        short delta = GET_WHEEL_DELTA_WPARAM(pMouse->mouseData);

        scrollDelta += delta; // accumulate scroll
        // std::cout << "Scroll delta: " << delta << " | Total: " << scrollDelta << std::endl;
    }

    return CallNextHookEx(hHook, nCode, wParam, lParam);
}


void MessagePump(){
    // ini ada bug, HInstance, hHook , etc harus dalam 1 fungsi kalau gk ngelag gk tau kenapa
    MSG msg;
    HINSTANCE hInstance = GetModuleHandle(NULL);
    hHook = SetWindowsHookEx(WH_MOUSE_LL, MouseProc, hInstance, 0);
    while (GetMessage(&msg, NULL, 0, 0))
    {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }
}

void startHook(){
    std::thread eventPump(MessagePump);
    eventPump.detach();
}

// Poll the mouse and push to capture only on change
// ini mausikin cap kedalam rada redundant soalnya cap itu singleton
void PollMouseWindows(MouseCapture& cap) {
    //start hook untuk event mouse yang butuh hook
    startHook();

    POINT prevPos = { -1, -1 };
    bool prevLeft = false, prevRight = false, prevMiddle = false;

    while (true) {
        POINT pos;
        GetCursorPos(&pos);

        bool leftNow   = (GetAsyncKeyState(VK_LBUTTON) & 0x8000) != 0;
        bool rightNow  = (GetAsyncKeyState(VK_RBUTTON) & 0x8000) != 0;
        bool middleNow = (GetAsyncKeyState(VK_MBUTTON) & 0x8000) != 0;

        // Wheel tracking
        short wheelNow = HIWORD(GetAsyncKeyState(VK_MBUTTON)); // or track WM_MOUSEWHEEL in a real window
  

        // Ini namanya diganti
        int dx = pos.x;
        int dy = pos.y;
        int dz = scrollDelta.exchange(0);

        // Only push if there is a change
        if (dx != 0 || dy != 0 || leftNow != prevLeft || rightNow != prevRight
            || middleNow != prevMiddle || dz != 0) {

            cap.push(dx, dy, dz, leftNow, rightNow, middleNow);

            prevPos = pos;
            prevLeft = leftNow;
            prevRight = rightNow;
            prevMiddle = middleNow;
        }
        // std::cout<<"dx: "<<dx<<" dy: "<<dy<<" dz: "<<dz<<std::endl;

        std::this_thread::sleep_for(std::chrono::milliseconds(SLEEP_DURATION));
    }
}

void WinApplyMouseState(const MouseState& state, MouseState& prevState) {
    INPUT input = {0};
    input.type = INPUT_MOUSE;

    // --- Move mouse relative ---
    if (state.dx != 0 || state.dy != 0) {
        input.mi.dx = state.dx;
        input.mi.dy = state.dy;
        input.mi.dwFlags = MOUSEEVENTF_MOVE|MOUSEEVENTF_ABSOLUTE;
        SendInput(1, &input, sizeof(INPUT));
    }

    // --- Scroll wheel ---
    if (state.dScroll != 0) {
        input.mi = {};
        input.type = INPUT_MOUSE;
        input.mi.mouseData = state.dScroll ; // 1 = 120 units
        input.mi.dwFlags = MOUSEEVENTF_WHEEL;
        SendInput(1, &input, sizeof(INPUT));
    }

    // --- Left click ---
    if (state.leftClick != prevState.leftClick) {
        // Down
        if (state.leftClick ==1){
            input.mi = {};
            input.type = INPUT_MOUSE;
            input.mi.dwFlags = MOUSEEVENTF_LEFTDOWN;
            SendInput(1, &input, sizeof(INPUT));
        }
        else{
            input.mi = {};
            input.type = INPUT_MOUSE;
            input.mi.dwFlags = MOUSEEVENTF_LEFTUP;
            SendInput(1, &input, sizeof(INPUT));
        }
        prevState.leftClick = state.leftClick;
    }


    // --- Right click ---
    if (state.rightClick != prevState.rightClick) {
        if(state.rightClick == 1){
            input.mi = {};
            input.type = INPUT_MOUSE;
            input.mi.dwFlags = MOUSEEVENTF_RIGHTDOWN;
            SendInput(1, &input, sizeof(INPUT));
        }else{
            input.mi = {};
            input.type = INPUT_MOUSE;
            input.mi.dwFlags = MOUSEEVENTF_RIGHTUP;
            SendInput(1, &input, sizeof(INPUT));
        }
        prevState.rightClick = state.rightClick;
    }

    // --- Middle click ---
    if (state.midClick != prevState.midClick) {
        if (state.midClick == 1){
            input.mi = {};
            input.type = INPUT_MOUSE;
            input.mi.dwFlags = MOUSEEVENTF_MIDDLEDOWN;
            SendInput(1, &input, sizeof(INPUT));
        }else{
            input.mi = {};
            input.type = INPUT_MOUSE;
            input.mi.dwFlags = MOUSEEVENTF_MIDDLEUP;
            SendInput(1, &input, sizeof(INPUT));
        }
        prevState.midClick = state.midClick;
    }
}
