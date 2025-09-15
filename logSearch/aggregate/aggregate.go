package aggregate

import (
	// "context"
	"log"
	"regexp"
	"time"

	// "fmt"

	"github.com/gofiber/fiber/v2"

	// "github.com/go-redis/redis/v8"
	// "github.com/minio/minio-go/v7"
	// "github.com/minio/minio-go/v7/pkg/credentials"
)

func MainFunction(c *fiber.Ctx, uuid string, starttime string, endtime string, regex string, contains string) {




	// Initialize Redis client
	// rdb := redis.NewClient(&redis.Options{
	// 	Addr: "redis:6379", // Adjust the address if needed
	// })

	// timestartStr := c.Query("timestart")
	// timeendStr := c.Query("timeend")

	// Parse the timeStart
	_, err := time.Parse(time.RFC3339, starttime)
	if err != nil {
		c.Status(fiber.StatusBadRequest).SendString("Invalid timeStart format")
		return
	}
	_, err1 := time.Parse(time.RFC3339, endtime)
	if err1 != nil {
		c.Status(fiber.StatusBadRequest).SendString("Invalid timeEnd format")
		return
	}
	if regex != "" {
		_, err := regexp.Compile(regex)
		if err != nil {
			log.Printf("Invalid regex: %v", regex)
			c.Status(fiber.StatusBadRequest).SendString("Invalid regex")
		return
		}

	}
	
	AccumulateResults(FilterRequest{StartTime: starttime, EndTime: endtime, Regex: regex, Contains:  contains},uuid)


	// // Generate all possible prefixes between startPrefix and endPrefix
	// var prefixes []string
	// for t := timeStart; !t.After(timeEnd); t = t.Add(time.Hour) {
	// 	prefix := fmt.Sprintf("%d/%02d/%02d/%02d/", t.Year(), t.Month(), t.Day(), t.Hour())
	// 	prefixes = append(prefixes, prefix)
	// }

	// // Load the Shared AWS Configuration (~/.aws/config)
	// cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	// if err != nil {
	// 	log.Fatalf("unable to load SDK config, %v", err)
	// }

	// // Create an Amazon S3 service client
	// svc := s3.NewFromConfig(cfg)

	// // Define the bucket
	// bucket := "your-bucket-name" // Replace with your bucket name

	// var objects []string
	// for _, prefix := range prefixes {
	// 	// List objects under the specified prefix
	// 	input := &s3.ListObjectsV2Input{
	// 		Bucket: &bucket,
	// 		Prefix: &prefix,
	// 	}

	// 	paginator := s3.NewListObjectsV2Paginator(svc, input)

	// 	for paginator.HasMorePages() {
	// 		page, err := paginator.NextPage(context.TODO())
	// 		if err != nil {
	// 			log.Fatalf("failed to get page, %v", err)
	// 		}

	// 		for _, obj := range page.Contents {
	// 			objects = append(objects, *obj.Key)
	// 		}
	// 	}
	// }

	// // Return the list of objects
	// if err := c.JSON(fiber.Map{
	// 	"objects": objects,
	// }); err != nil {
	// 	log.Printf("failed to send JSON response, %v", err)
	// }

}
