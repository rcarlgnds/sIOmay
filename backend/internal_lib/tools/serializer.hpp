#pragma once
// UNIVERSAL FORMAT FOR THIS PROGRAM
// byte[0]     = Operation Code 
// byte[1]     = Sub Operations Code
// byte[2-3]   = reserved for additional data (unused currently)
// byte[4-7]   = data1 [x]
// byte[8-11]  = data2 [y]
// byte[12-15] = data3 [z]
#include <stdint.h>
#include "../mouse/Mouse.hpp"
#include "../keyboard/Keyboard.hpp"
#include "../system/system.hpp"
//Operation Codes
#define MOUSE_ACTION        0x01 
#define KEYBOARD_ACTION     0x02

//Sub Operation Codes
// Mouse
#define MOUSE_MOVE          0x01
#define MOUSE_LEFT_CLICK    0x02
#define MOUSE_RIGHT_CLICK   0x04
#define MOUSE_MID_CLICK     0x08
#define MOUSE_SCROLL        0x10

// Keyboard
#define KEYBOARD_KEYDOWN    0x01
#define KEYBOARD_KEYUP      0x00   

//General Function
void clearBuff(uint8_t *buf, int len );

bool isMouseData(uint8_t *buf, int len);

bool isKeyboardData(uint8_t *buf, int len);

bool isCommandData(uint8_t *buf, int len);

//ConvertFunction MouseState To Byte
void formatMouseData(MouseState &mState, uint8_t *buf, int len);

void parseMouseData(MouseState &mState, const uint8_t *buf, int len);

//ConvertFunction KeyboardState To Byte
void formatKeyboardData(KeyboardState &kState, uint8_t *buf, int len);

void parseKeyboardData(KeyboardState &kState, const uint8_t *buf, int len);

SystemCommand parseCommandData(int *data, const uint8_t *buf, int len);

void formatStopCommandData(uint8_t *buf, int len, int ip[4]);