#include"Mouse.hpp"

void MouseState::SetState(MouseState & container,int dx, int dy, int dscroll,
                  bool lClick, bool rClick, bool midClick)
{
    container.dx = dx;
    container.dy = dy;
    container.dScroll = dscroll;
    container.leftClick = lClick;
    container.rightClick = rClick;
    container.midClick = midClick;
};