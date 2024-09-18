package main

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// receipt structure
type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Total        string `json:"total"`
	Items        []Item `json:"items"`
}

// item structure
type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

// PointsResponse and ReceiptIDResponse structure
type PointsResponse struct {
	Points int `json:"points"`
}

type ReceiptIDResponse struct {
	ID string `json:"id"`
}

// store receipts in-memory
var (
	receiptStorage    = make(map[string]Receipt)
	mu                sync.Mutex
	alphanumericRegex = regexp.MustCompile(`[a-zA-Z0-9]`)
)

func main() {
	r := mux.NewRouter()

	// Route Handlers
	r.HandleFunc("/receipts/process", processReceiptHandler).Methods("POST")
	r.HandleFunc("/receipts/{id}/points", getPointsHandler).Methods("GET")

	log.Println("Server's up! Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// handler to process receipt and store it
func processReceiptHandler(w http.ResponseWriter, r *http.Request) {
	receipt, err := decodeRequest(r)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	receiptID := storeReceipt(receipt)
	respondWithJSON(w, http.StatusOK, ReceiptIDResponse{ID: receiptID})
}

// handler to get points for a stored receipt
func getPointsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	receipt, found := getReceipt(vars["id"])
	if !found {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	points, err := calculatePoints(receipt)
	if err != nil {
		log.Printf("Error calculating points: %v", err)
		http.Error(w, "Error calculating points", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, PointsResponse{Points: points})
}

// decode the incoming request to a Receipt object
func decodeRequest(r *http.Request) (Receipt, error) {
	var receipt Receipt
	err := json.NewDecoder(r.Body).Decode(&receipt)
	return receipt, err
}

// Store the receipt and return the generated id
func storeReceipt(receipt Receipt) string {
	receiptID := generateReceiptID()
	mu.Lock()
	receiptStorage[receiptID] = receipt
	mu.Unlock()
	return receiptID
}

// Retrieve receipt by id
func getReceipt(id string) (Receipt, bool) {
	mu.Lock()
	receipt, exists := receiptStorage[id]
	mu.Unlock()
	return receipt, exists
}

// generte a random receipt id
func generateReceiptID() string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	id := make([]rune, 16)
	for i := range id {
		id[i] = letters[rand.Intn(len(letters))]
	}
	return string(id)
}

// Helper function to respond with json
func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// main logic of point calculation with error handling
func calculatePoints(receipt Receipt) (int, error) {
	points := 0

	// Calculate points from retailer name
	retailerPoints := calculateRetailerPoints(receipt.Retailer)
	points += retailerPoints
	log.Printf("Retailer points: %d (Total: %d)\n", retailerPoints, points)

	// Parse the total amount
	totalFloat, err := strconv.ParseFloat(receipt.Total, 64)
	if err != nil {
		return 0, err
	}

	// Calculate points from the total amount
	totalPoints := calculateTotalPoints(totalFloat)
	points += totalPoints
	log.Printf("Total points: %d (Total: %d)\n", totalPoints, points)

	// Calculate points from items
	itemPoints := calculateItemPoints(receipt.Items)
	points += itemPoints
	log.Printf("Item points: %d (Total: %d)\n", itemPoints, points)

	// Calculate points from purchase date
	oddDatePoints := calculateOddDatePoints(receipt.PurchaseDate)
	points += oddDatePoints
	log.Printf("Odd date points: %d (Total: %d)\n", oddDatePoints, points)

	// Calculate points from purchase time
	timePoints := calculateTimePoints(receipt.PurchaseTime)
	points += timePoints
	log.Printf("Time points: %d (Total: %d)\n", timePoints, points)

	log.Printf("Final total points: %d\n", points)
	return points, nil
}

// Calculate points based on the retailer name
func calculateRetailerPoints(retailer string) int {
	retailerAlnumChars := alphanumericRegex.FindAllString(retailer, -1)
	points := len(retailerAlnumChars)
	log.Printf("Retailer name '%s' has %d alphanumeric characters\n", retailer, points)
	return points
}

// Calculate points based on the total amount
func calculateTotalPoints(total float64) int {
	points := 0
	if total == float64(int(total)) {
		points += 50
		log.Printf("Total amount is a round dollar amount, added 50 points\n")
	}
	if int(total*100)%25 == 0 {
		points += 25
		log.Printf("Total amount is a multiple of 0.25, added 25 points\n")
	}
	return points
}

// Calculate points based on items
func calculateItemPoints(items []Item) int {
	points := (len(items) / 2) * 5 // 5 points for every two items
	log.Printf("%d item pairs, added %d points\n", len(items)/2, points)
	for _, item := range items {
		trimmedDescription := strings.TrimSpace(item.ShortDescription)
		itemPrice, err := strconv.ParseFloat(item.Price, 64)
		if err != nil {
			log.Printf("Invalid price for item: %s", item.ShortDescription)
			continue
		}
		if len(trimmedDescription)%3 == 0 {
			extraPoints := int(math.Ceil(itemPrice * 0.2))
			points += extraPoints
			log.Printf("Item '%s' has description length divisible by 3, added %d points (Item price: %.2f)\n", trimmedDescription, extraPoints, itemPrice)
		} else {
			log.Printf("Item '%s' does not have description length divisible by 3, no extra points\n", trimmedDescription)
		}
	}
	return points
}

// Calculate points based on the purchase date
func calculateOddDatePoints(dateStr string) int {
	purchaseDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Printf("Invalid purchase date: %s", dateStr)
		return 0
	}
	if purchaseDate.Day()%2 != 0 {
		log.Printf("Purchase date '%s' is an odd day, added 6 points\n", dateStr)
		return 6
	}
	log.Printf("Purchase date '%s' is not an odd day, no points\n", dateStr)
	return 0
}

// Calculate points based on the purchase time
func calculateTimePoints(timeStr string) int {
	purchaseTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		log.Printf("Invalid purchase time: %s", timeStr)
		return 0
	}
	if purchaseTime.Hour() >= 14 && purchaseTime.Hour() < 16 {
		log.Printf("Purchase time '%s' is between 2:00 PM and 4:00 PM, added 10 points\n", timeStr)
		return 10
	}
	log.Printf("Purchase time '%s' is not between 2:00 PM and 4:00 PM, no points\n", timeStr)
	return 0
}
