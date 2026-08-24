package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// Demonstrates Redis Geo commands for location-based queries
func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()

	defer rdb.Close()

	// Clean up
	rdb.Del(ctx, "cities")

	fmt.Println("=== Redis Geo Commands ===")

	// ========== GEOADD - Add Locations ==========

	fmt.Println("\n1. Adding Locations with GEOADD")

	// Syntax: GEOADD key longitude latitude member [longitude latitude member ...]
	// Note: Redis uses [longitude, latitude] order (not latitude, longitude)
	cities := map[string][2]float64{
		"New York":    {-74.006, 40.7128},
		"Los Angeles": {-118.2437, 34.0522},
		"Chicago":     {-87.6298, 41.8781},
		"London":      {-0.1276, 51.5074},
		"Paris":       {2.3522, 48.8566},
		"Tokyo":       {139.6917, 35.6895},
		"Sydney":      {151.2093, -33.8688},
		"Berlin":      {13.4050, 52.5200},
	}

	for city, coords := range cities {
		_, err := rdb.GeoAdd(ctx, "cities",
			&redis.GeoLocation{
				Name:      city,
				Longitude: coords[0],
				Latitude:  coords[1],
			},
		).Result()
		if err != nil {
			log.Fatalf("GEOADD failed: %v", err)
		}
		fmt.Printf("Added %s at [%f, %f]\n", city, coords[0], coords[1])
	}

	fmt.Println()

	// ========== GEODIST - Calculate Distance ==========

	fmt.Println("2. Calculating Distance with GEODIST")

	// Get distance between two cities
	// Options: m, km, ft, mi
	distance, err := rdb.GeoDist(ctx, "cities", "New York", "London", "km").Result()
	if err != nil {
		log.Fatalf("GEODIST failed: %v", err)
	}
	fmt.Printf("Distance: New York -> London = %.2f km\n", distance)

	distance, err = rdb.GeoDist(ctx, "cities", "Tokyo", "Sydney", "mi").Result()
	fmt.Printf("Distance: Tokyo -> Sydney = %.2f miles\n", distance)

	fmt.Println()

	// ========== GEOHASH - Get Geo Hash ==========

	fmt.Println("3. Getting Geo Hash with GEOHASH")

	// Geo hashes are base32 encoded strings representing location
	// Useful for indexing and proximity searches
	hashes, err := rdb.GeoHash(ctx, "cities", "Paris", "Berlin", "London").Result()
	if err != nil {
		log.Fatalf("GEOHASH failed: %v", err)
	}

	for i, city := range []string{"Paris", "Berlin", "London"} {
		fmt.Printf("%s hash: %s\n", city, hashes[i])
	}

	fmt.Println()

	// ========== GEOPOS - Get Coordinates ==========

	fmt.Println("4. Getting Coordinates with GEOPOS")

	positions, err := rdb.GeoPos(ctx, "cities", "New York", "Tokyo").Result()
	if err != nil {
		log.Fatalf("GEOPOS failed: %v", err)
	}

	for i, city := range []string{"New York", "Tokyo"} {
		if positions[i] != nil {
			fmt.Printf("%s position: [lon: %.5f, lat: %.5f]\n", city, positions[i].Longitude, positions[i].Latitude)
		} else {
			fmt.Printf("%s not found\n", city)
		}
	}

	fmt.Println()

	// ========== GEORADIUS - Search by Radius ==========

	fmt.Println("5. Searching by Radius with GEORADIUS")

	// Find cities within 5000 km of London
	radiusResults, err := rdb.GeoRadius(ctx, "cities", -0.1276, 51.5074,
		&redis.GeoRadiusQuery{
			Radius:    5000,
			Unit:      "km",
			WithCoord: true,  // include coordinates
			WithDist:  true,  // include distance
			Count:     10,    // limit results
			Sort:      "ASC", // sort by distance
		}).Result()
	if err != nil {
		log.Fatalf("GEORADIUS failed: %v", err)
	}

	fmt.Printf("Cities within 5000 km of London:\n")
	for _, loc := range radiusResults {
		fmt.Printf("  %s: %.2f km away\n", loc.Name, loc.Dist)
	}

	fmt.Println()

	// ========== GEORADIUSBYMEMBER - Search from Member ==========

	fmt.Println("6. Searching from Member with GEORADIUSBYMEMBER")

	// Find cities within 2000 km of Paris
	nearby, err := rdb.GeoRadiusByMember(ctx, "cities", "Paris",
		&redis.GeoRadiusQuery{
			Radius:    2000,
			Unit:      "km",
			WithCoord: true,
			WithDist:  true,
			Sort:      "ASC",
		}).Result()
	if err != nil {
		log.Fatalf("GEORADIUSBYMEMBER failed: %v", err)
	}

	fmt.Printf("Cities within 2000 km of Paris:\n")
	for _, loc := range nearby {
		fmt.Printf("  %s: %.2f km away\n", loc.Name, loc.Dist)
	}

	fmt.Println()

	// ========== GEOSEARCH - More Flexible Search ==========
	// Note: go-redis v9 uses different field names for GeoSearchQuery

	fmt.Println("7. Flexible Search with GEOSEARCH")

	// Search by radius from a coordinate
	boxSearch, err := rdb.GeoSearch(ctx, "cities",
		&redis.GeoSearchQuery{
			Longitude:  -0.1276,
			Latitude:   51.5074,
			Radius:     3000,
			RadiusUnit: "km",
			Sort:       "ASC",
		}).Result()
	if err != nil {
		log.Fatalf("GEOSEARCH failed: %v", err)
	}

	fmt.Printf("Cities within 3000 km of London (GEOSEARCH):\n")
	for _, loc := range boxSearch {
		fmt.Printf("  %s\n", loc)
	}

	fmt.Println()

	// ========== GEOSEARCHSTORE - Store Results ==========
	// Note: go-redis v9 uses GeoSearchStoreQuery

	fmt.Println("8. Storing Search Results with GEOSEARCHSTORE")

	// Store search results in another key (as a sorted set)
	// This is useful for caching frequently-used searches
	count, err := rdb.GeoSearchStore(ctx, "near_london", "cities",
		&redis.GeoSearchStoreQuery{
			GeoSearchQuery: redis.GeoSearchQuery{
				Longitude:  -0.1276,
				Latitude:   51.5074,
				Radius:     4000,
				RadiusUnit: "km",
			},
		}).Result()
	if err != nil {
		log.Fatalf("GEOSEARCHSTORE failed: %v", err)
	}

	fmt.Printf("Stored %d results in 'near_london'\n", count)

	// Verify stored results
	stored, _ := rdb.ZRange(ctx, "near_london", 0, -1).Result()
	fmt.Printf("Stored locations: %v\n", stored)

	fmt.Println()

	// ========== CLEANUP ==========

	fmt.Println("\n=== Cleanup ===")
	rdb.Del(ctx, "cities", "near_london")
	fmt.Println("Cleanup completed")
}
