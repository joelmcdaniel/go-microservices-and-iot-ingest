package main

import (
	"arena"
	"fmt"
	"unsafe"

	"github.com/sugawarayuuta/refloat"
)

// TelemetryEvent is our target struct.
// Notice we use string and float64, just like normal.
type TelemetryEvent struct {
	SensorID  string
	Value     float64
	Timestamp int64
}

// unsafeString converts a byte slice to a string without allocation.
// ⚠️ WARNING: The string is only valid as long as the original 'b' is valid.
// If 'b' is modified, the string changes. If 'b' is freed, the string is garbage.
func unsafeString(b []byte) string {
	// We use unsafe.String to cast the pointer directly
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// ParseTelemetry simulates parsing a simple CSV or JSON format.
// Format: "ID,Value" -> "sensor-abc,99.5"
// We use the Arena to allocate the struct itself.
func ParseTelemetry(mem *arena.Arena, payload []byte) *TelemetryEvent {
	// 1. Allocate the struct inside the Arena.
	// This is much faster than standard 'new(TelemetryEvent)'
	event := arena.New[TelemetryEvent](mem)

	// SIMPLIFIED PARSING LOGIC for demonstration
	// In a real scenario, you'd use a zero-alloc JSON library or efficient loop.
	// Let's assume a comma separates ID and Value.
	// Find the comma
	commaIndex := -1
	for i, b := range payload {
		if b == ',' {
			commaIndex = i
			break
		}
	}

	if commaIndex != -1 {
		// 2. Zero-Copy String Creation
		// We take the bytes for the ID and cast them to string.
		// NO MEMORY COPY HAPPENS HERE.
		idBytes := payload[:commaIndex]
		event.SensorID = unsafeString(idBytes)

		valBytes := payload[commaIndex+1:]
		val, _ := refloat.ParseFloat(unsafeString(valBytes), 64)
		event.Value = val
	}

	return event

}

func main() {
	// Simulation of a high-frequency loop
	inputData := []byte("turbine-x99,45.2")

	// 1. Create the Arena
	// This allocates a large block of memory from the OS.
	mem := arena.NewArena()

	// 2. The Process
	// In a web server, you might create one arena per request.
	event := ParseTelemetry(mem, inputData)

	fmt.Printf("Parsed Event: ID=%s,Val=%.2f\n", event.SensorID, event.Value)

	// 3. The Cleanup
	// We don't rely on GC to clean up 'event'.
	// We free the ENTIRE arena manually.
	// 'event' is now invalid and accessing it will cause a crash (segfault).
	mem.Free()

	// fmt.Println(event.SensorID) <--- THIS WOULD PANIC
	fmt.Println("Arena freed. No GC pressure generated.")
}
