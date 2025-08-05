// mouse_move.cpp
#include <windows.h>

extern "C" void MoveMouse() {
    DWORD dwFlags = MOUSEEVENTF_MOVE;
    DWORD dx = 1, dy = 1;
    DWORD dwData = 0;
    ULONG_PTR extra = 0;

    mouse_event(
        dwFlags,
        dx,
        dy,
        dwData,
        extra
    );
}
