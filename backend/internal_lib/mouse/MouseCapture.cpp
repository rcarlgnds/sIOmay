
#include "Mouse.hpp"


bool MouseStateQueue::push(  int dx, int dy, int dScroll,
            bool leftClick, bool rightClick, bool midClick) 
{
    if (count == MOUSE_STATE_BUFFER) {
        return false; // buffer full
    }

    // Write directly into the buffer at tail, avoiding a temporary MouseState
    MouseState::SetState(mouseStateArr[tail], dx, dy, dScroll, leftClick, rightClick, midClick);

    tail = (tail + 1) % MOUSE_STATE_BUFFER;
    count++;
    return true;
}

bool MouseStateQueue::pop(MouseState& outState) {
    if (count == 0) {
        return false; // buffer empty
    }
    outState = mouseStateArr[head];
    head = (head + 1) % MOUSE_STATE_BUFFER;
    count--;
    return true;
}


//MouseCapter Impl
std::unique_ptr<MouseCapture> MouseCapture::instance = nullptr;
MouseCapture * MouseCapture::GetInstance(){
    if (!instance){
        instance = std::make_unique<MouseCapture>(MouseCapture());
    }
    return instance.get();
}


