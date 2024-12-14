package services

import (
	"AsyncProd/config"
	"AsyncProd/models"
	"AsyncProd/pkg/image"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/streadway/amqp"
)

type ImageProcessingMessage struct {
	ProductID     int      `json:"product_id"`
	UserID        int      `json:"user_id"`
	ImageURLs     []string `json:"image_urls"`
}

//sends image processing msgg to RabbitMQ
func PublishImageProcessingMessage(productID, userID int, imageURLs []string) error {
	message := ImageProcessingMessage{
		ProductID: productID,
		UserID:    userID,
		ImageURLs: imageURLs,
	}

	//Convert msg to JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	//Publish to RabbitMQ
	err = config.RabbitMQChannel.Publish(
		"",                            
		config.ImageProcessingQueue,   
		false,                         
		false,                         
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}

// consumes msgs from RabbitMQ and processes images
func ProcessImageFromQueue() {
	msgs, err := config.RabbitMQChannel.Consume(
		config.ImageProcessingQueue, 
		"",                          
		false,                       
		false,                       
		false,                       
		false,                       
		nil,                       
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			var processMsg ImageProcessingMessage
			err := json.Unmarshal(msg.Body, &processMsg)
			if err != nil {
				log.Printf("Error parsing message: %v", err)
				msg.Reject(false)
				continue
			}

			compressedImages, err := processImagesForProduct(processMsg)
			if err != nil {
				log.Printf("Error processing images: %v", err)
				msg.Reject(true) //requeue
				continue
			}
			err = updateProductCompressedImages(processMsg.ProductID, compressedImages)
			if err != nil {
				log.Printf("Error updating product: %v", err)
				msg.Reject(true)
				continue
			}
			msg.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func processImagesForProduct(msg ImageProcessingMessage) ([]string, error) {
	var compressedImageURLs []string

	for _, imgURL := range msg.ImageURLs {
		log.Printf("Processing image: %s", imgURL)
		compressedImg, err := image.CompressImage(imgURL)
		if err != nil {
			log.Printf("ERROR: Failed to compress image %s: %v", imgURL, err)
			continue
		}
		log.Printf("SUCCESS: Compressed image %s", imgURL)

		s3Key := fmt.Sprintf("products/%d/%s", msg.ProductID, generateUniqueFileName(imgURL))
		log.Printf("Generated S3 key for image %s: %s", imgURL, s3Key)
		log.Printf("Uploading image %s to S3 bucket: %s with key: %s", imgURL, config.S3Bucket, s3Key)
		_, err = config.S3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(config.S3Bucket),
			Key:    aws.String(s3Key),
			Body:   aws.ReadSeekCloser(bytes.NewReader(compressedImg)),
		})
		if err != nil {
			log.Printf("ERROR: Failed to upload image %s to S3: %v", imgURL, err)
			continue
		}
		log.Printf("SUCCESS: Uploaded image %s to S3: %s", imgURL, s3Key)
		compressedURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", config.S3Bucket, s3Key)
		log.Printf("Generated public URL for image %s: %s", imgURL, compressedURL)
		compressedImageURLs = append(compressedImageURLs, compressedURL)
	}

	if len(compressedImageURLs) == 0 {
		log.Printf("WARNING: No images were successfully processed for product ID: %d", msg.ProductID)
	}

	return compressedImageURLs, nil
}

// updates product with compressed image URLs
func updateProductCompressedImages(productID int, compressedImages []string) error {
	product, err := models.GetProductByID(productID)
	if err != nil {
		return fmt.Errorf("failed to retrieve product: %v", err)
	}

	product.CompressedImages = compressedImages
	err = models.UpdateProduct(product)
	if err != nil {
		return fmt.Errorf("failed to update product: %v", err)
	}

	return nil
}

// to generate unique filename
func generateUniqueFileName(originalURL string) string {
	return fmt.Sprintf("%s_%d%s", 
		filepath.Base(originalURL), 
		time.Now().UnixNano(), 
		filepath.Ext(originalURL),
	)
}