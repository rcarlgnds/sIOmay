//Ini yang buat queue keyboard statenya 100% sama kyk yang mouse nanti di buat template ya VU
// biar gk duplicate codenya. karena yang sekarang hardcode dan duplicate

#pragma once 
#define KEYBOARD_STATE_BUFFER 256


#include "KeyboardWindows.h"
#include<memory>
struct KeyboardState{
    bool press; // 0 key up, 1 key down
    int code ; // keycode, key apa yang ditekan

    static void setKeyboardState(KeyboardState &s, bool press, int code);
};

struct KeyboardStateQueue{
    private:
        KeyboardState keyboardStateArr[KEYBOARD_STATE_BUFFER] = {};
        int head =0;
        int tail =0;
        int count =0;
    public:
  
        bool push(bool press, int code);

        bool pop(KeyboardState& outState);

    inline bool isEmpty() const { return count == 0; };
    inline bool isFull() const { return count == KEYBOARD_STATE_BUFFER; };
};

struct KeyboardCapture {
    private:
        static std::unique_ptr<KeyboardCapture> instance;
        KeyboardStateQueue queue;

    public:
        static KeyboardCapture * GetInstance();
        inline bool push(bool press, int code) {
            return queue.push(press, code);
        }

        inline bool poll(KeyboardState &outState) {
            return queue.pop(outState);
        }

};


