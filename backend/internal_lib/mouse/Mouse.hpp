#pragma once

#include <memory>
#define MOUSE_STATE_BUFFER  255


struct MouseState{
    int dx = 0;
    int dy = 0;
    int dScroll = 0;

    bool midClick = false;
    bool rightClick = false;
    bool leftClick = false;


    static void SetState(MouseState & container,int dx, int dy, int dscroll,
                  bool lClick, bool rClick, bool midClick); 

};


struct MouseStateQueue{
    private:
        MouseState mouseStateArr[MOUSE_STATE_BUFFER] = {};
        int head =0;
        int tail =0;
        int count =0;
    public:

  
        bool push(  int dx, int dy, int dScroll,
                    bool leftClick, bool rightClick, bool midClick);

        bool pop(MouseState& outState);

    inline bool isEmpty() const { return count == 0; };
    inline bool isFull() const { return count == MOUSE_STATE_BUFFER; };
};


struct MouseCapture {
    private:
        static std::unique_ptr<MouseCapture> instance;
        MouseStateQueue queue;

    public:
        static MouseCapture * GetInstance();
        inline bool push(int dx, int dy, int dScroll, bool left, bool right, bool mid) {
            return queue.push(dx, dy, dScroll, left, right, mid);
        }

        inline bool poll(MouseState &outState) {
            return queue.pop(outState);
        }

};  