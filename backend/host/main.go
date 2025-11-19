package main

/*
#cgo LDFLAGS: -L. -lcgo_compatible -lstdc++ -lws2_32 -luser32 -static
#include "../internal_lib/extern.hpp"
*/
import "C"

func main() {
	C.startClientC()
}
