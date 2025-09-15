package main

import (
	"context"
	"logSearch/aggregate"
	"encoding/json"
	"log"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	// "logSearch/aggregate" // Replace with the correct import path
)

var ctx = context.Background()

func main() {
	app := fiber.New()

	// Enable CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Allow requests from this origin
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Initialize Redis client
	log.Print("Connecting to Redis.")
	// rdb := redis.NewClient(&redis.Options{
	// 	Addr: "redis:6379", // Adjust the address if needed
	// })

	log.Print("Connected to Redis.")
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Route when uuid parameter is not passed
	app.Get("/search", func(c *fiber.Ctx) error {
		newUUID := uuid.New().String()


		// Call MainFunction with the new UUID
		aggregate.MainFunction(c, newUUID, c.Query("startTime"), c.Query("endTime"), c.Query("regex",""), c.Query("text",""))

		// Only return the new UUID
		return c.JSON(fiber.Map{"uuid": newUUID})

		


	})

	// Route when uuid parameter is passed
	app.Get("/search/:uuid/:offset", func(c *fiber.Ctx) error {
		param := c.Params("uuid")
		offsetStr := c.Params("offset")
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid offset")
		}

		log.Printf("Fetching logs for UUID: %s with offset: %d", param, offset)

		// Fetch logs from Redis
		data, err := aggregate.Rdb.Get(ctx, param).Result()
		if err != nil {
			if err == redis.Nil {
				return c.Status(fiber.StatusNotFound).SendString("UUID not found")
			}
			return c.Status(fiber.StatusInternalServerError).SendString("Error fetching data from Redis")
		}

		// var results []map[string]interface{}

		// if err := json.Unmarshal([]byte(data), &results); err != nil {
		// 	return c.Status(fiber.StatusInternalServerError).SendString("Error unmarshalling JSON")
		// }
		// events:=ReadEvents()
		var results []aggregate.Event
		if err := json.Unmarshal([]byte(data), &results); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error unmarshalling JSON")
		}
		// result := make([]map[string]interface{}, len(events))
	// for i, event := range results {
	// 	result[i] = map[string]interface{}{
	// 		"timestamp": event.Timestamp,
	// 		"data":      event.Data,
	// 	}
	// }

		// Paginate the results
		chunkSize := 50
		if offset > len(results) {
			return c.Status(fiber.StatusBadRequest).SendString("Offset out of range")
		}

		end := offset + chunkSize
		if end > len(results) {
			end = len(results)
		}

		// paginatedResults := results[offset:end]

		return c.JSON(fiber.Map{
			"data":   results,
			"offset": end,
		})
	})

	// start()

	log.Fatal(app.Listen(":3000"))
}
