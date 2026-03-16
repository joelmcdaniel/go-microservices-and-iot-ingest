package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"
)

// 1. Context Awareness
// Every long-running function must accept a Context as its first argument.
// It must check ctx.Done() frequently to see if it should stop.
func monitorSensor(ctx context.Context, sensorID string) error {
	fmt.Printf("[%s] Sensor starting up...\n", sensorID)

	ticker := time.NewTicker(500 * time.Microsecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// The supervisor pulled the plug. Clean up and exit.
			fmt.Printf("[%s] Context cancelled, shutting down sensor.\n", sensorID)
			return ctx.Err() // Usually returns context.Canceled
		case <-ticker.C:
			// Simulate a reading
			temp := rand.Intn(120)
			fmt.Printf("[%s] Reading: %d°C\n", sensorID, temp)

			// SIMULATED FAILURE:
			// If temperature exceeds 100, the sensor burns out.
			if temp > 100 {
				return fmt.Errorf("sensor%s overheated! temp=%d", sensorID, temp)
			}
		}
	}
}

func databaseHealthCheck(ctx context.Context) error {
	fmt.Println("[DB] Connection monitor started...")
	// Simulate a database connection check loop
	for {
		select {
		case <-ctx.Done():
			fmt.Println("[DB] Context cancelled, stopping monitor.")
			return ctx.Err()
		case <-time.After(2 * time.Second):
			// Database is healthy...
			fmt.Println("[DB] Heartbeat: OK")
		}
	}
}

func main() {
	// 2. Create the Supervisor
	// We start with a background context and wrap it in an errgroup.
	// 'g' is our group supervisor.
	// 'ctx' is the cancelable context tied to this group.
	g, ctx := errgroup.WithContext(context.Background())

	fmt.Println("--- FACTORY SYSTEM ONLINE ---")

	// 3. Assign Tasks
	// We lauch the Sensor Monitor.
	g.Go(func() error {
		// We pass the group's context, NOT the background context
		// This connects the sensor to the "kill switch".
		return monitorSensor(ctx, "TURBINE-O1")
	})

	// We launch the Database Health Check.
	g.Go(func() error {
		return databaseHealthCheck(ctx)
	})

	// 4. Wait for Outcome
	// g.Wait() blocks until ALL goroutines have exited.
	// If any gorouting returned an error, g.Wait() returns that error.
	err := g.Wait()

	fmt.Println("--- SYSTEM SHUTDOWN ---")
	if err != nil {
		fmt.Printf("Fatal Error:%v\n", err)
	} else {
		fmt.Println("System exited cleanly.")
	}
}
