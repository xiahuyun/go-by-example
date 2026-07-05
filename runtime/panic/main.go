package main

func runPanic() {
	panic("test panic")
}
func main() {
	defer func() {
		if r := recover(); r != nil {
			println("Recovered from panic:", r)
		}
	}()

	runPanic()
}
