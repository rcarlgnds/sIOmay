#include "keyboard.hpp"

void KeyboardState::setKeyboardState(KeyboardState &s, bool press, int code){
    s.press =press;
    s.code = code;
}

bool KeyboardStateQueue::push( bool press, int code){
    if (this->count == KEYBOARD_STATE_BUFFER){
        return false;
    }
    KeyboardState::setKeyboardState(this->keyboardStateArr[tail], press, code);
    this->count++;
    this->tail = (this->tail+1)%KEYBOARD_STATE_BUFFER;
    return true;
}

bool KeyboardStateQueue::pop(KeyboardState &outState){
    if (this->count == 0){
        return false;
    }
    outState = this->keyboardStateArr[this->head];
    this->head = (this->head+1)%KEYBOARD_STATE_BUFFER;
    this->count--;
    return true;
}


std::unique_ptr<KeyboardCapture> KeyboardCapture::instance = nullptr;
KeyboardCapture * KeyboardCapture::GetInstance(){
    if (!instance){
        instance = std::make_unique<KeyboardCapture>(KeyboardCapture());
    }
    return instance.get();
}
