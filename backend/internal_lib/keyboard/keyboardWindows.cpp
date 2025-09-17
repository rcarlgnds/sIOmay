#include <windows.h>
#include <iostream>
#include "keyboard.hpp"


HHOOK kHHook;
LRESULT CALLBACK keyboardProc(int nCode, WPARAM wParam, LPARAM lParam) {
    if (nCode == HC_ACTION) {

        KeyboardCapture *kCapture = KeyboardCapture::GetInstance();
        KBDLLHOOKSTRUCT* p = (KBDLLHOOKSTRUCT*)lParam;
        
        switch (wParam) {
            // Sengaja case emtpy buat capture WM_KEYDOWN + WM_SYSKEYDOWN
            case WM_KEYDOWN:

            case WM_SYSKEYDOWN:
                // std::cout << "Key Down: " << p->vkCode << std::endl;
                kCapture->push(1, p->vkCode);
                break;

            case WM_KEYUP:

            case WM_SYSKEYUP:
                // std::cout << "Key Up: " << p->vkCode << std::endl;
                kCapture->push(0, p->vkCode);
                break;
        }
    }
    return CallNextHookEx(kHHook, nCode, wParam, lParam);
}

void startKeyboardTrack(){
    HINSTANCE hInstance = GetModuleHandle(NULL);

    kHHook = SetWindowsHookEx(WH_KEYBOARD_LL, keyboardProc, hInstance, 0);
    if (!kHHook) {
        std::cerr << "Failed to install hook!" << std::endl;
        return;
    }

    MSG msg;
    while (GetMessage(&msg, NULL, 0, 0)) {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }

    UnhookWindowsHookEx(kHHook);
    return;
}


void WinApplyKeyInput(bool isKeyDown, int vkCode) {
    INPUT input{};
    input.type = INPUT_KEYBOARD;
    input.ki.wVk = static_cast<WORD>(vkCode);  // virtual key code
    input.ki.wScan = 0;                        // hardware scan code (0 = let system map it)
    input.ki.dwFlags = isKeyDown ? 0 : KEYEVENTF_KEYUP;
    input.ki.time = 0;
    input.ki.dwExtraInfo = GetMessageExtraInfo();

    SendInput(1, &input, sizeof(INPUT));
}

