package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/user/roku-stremio-bridge/internal/debrid"
	"github.com/user/roku-stremio-bridge/internal/stremio"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	r := gin.Default()

	// Example configuration
	addons := []string{
		"https://v3-cinemeta.strem.io",
	}
	rdKey := os.Getenv("RD_API_KEY")
	
	agg := stremio.NewAggregator(addons)
	rd := debrid.NewRealDebridClient(rdKey)

	// Aggregator endpoint for catalogs
	r.GET("/catalog/:type/:id", func(c *gin.Context) {
		contentType := c.Param("type")
		id := c.Param("id")

		items, err := agg.FetchCatalog(contentType, id)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"items": items})
	})

	// Stream resolver endpoint
	r.GET("/streams/:type/:id", func(c *gin.Context) {
		magnet := c.Query("magnet")
		if magnet == "" {
			c.JSON(400, gin.H{"error": "magnet query param required"})
			return
		}

		// Simple parsing of season/episode from query or ID
		// For now, let's assume they are passed as query params
		season := 0
		episode := 0
		fmt.Sscanf(c.Query("s"), "%d", &season)
		fmt.Sscanf(c.Query("e"), "%d", &episode)

		streamURL, err := rd.ResolveMagnet(magnet, season, episode)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"streams": []gin.H{
				{
					"name": "Real-Debrid",
					"url":  streamURL,
				},
			},
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Orchestrator starting on port %s", port)
	r.Run(":" + port)
}
