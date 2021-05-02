package main

import "C" // cgo

//export GetVersion
// sample export function
func GetVersion() string {
	return "1.0"
}

// default main function
func main() {

}
