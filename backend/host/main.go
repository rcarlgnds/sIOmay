package main

/*
#cgo windows LDFLAGS: -L. -lcgo_compatible -lstdc++ -lws2_32 -luser32 -static
#cgo darwin LDFLAGS: -L. -lmaccompatible -lstdc++
#include "../internal_lib/extern.hpp"
*/
import "C"

func main() {
	C.startClientC()
}
