package main

import (
    "encoding/json"
    "log"
    "math/rand"
    "net/http"
    "regexp"
    "strconv"
    "time"
	"math"
	"fmt"
	"strings"
    "github.com/gorilla/mux"
)

// The receipt structure with relevant fields
type Receipt struct {
    Retailer     string  `json:"retailer"`
    PurchaseDate string  `json:"purchaseDate"`
    PurchaseTime string  `json:"purchaseTime"`
    Total        string  `json:"total"`
    Items        []Item  `json:"items"`
}

// Items in the receipt
type Item struct {
    ShortDescription string `json:"shortDescription"`
    Price            string `json:"price"`
}

// Response for points awarded
type PointsResponse struct {
    Points int `json:"points"`
}

// Response containing the receipt ID
type ReceiptIDResponse struct {
    ID string `json:"id"`
}

// Store receipts in-memory (temporary, obviously)
var receiptStorage = make(map[string]Receipt)

func main() {
    r := mux.NewRouter()

    // Routes for processing the receipt and getting points
    r.HandleFunc("/receipts/process", processReceiptHandler).Methods("POST")
    r.HandleFunc("/receipts/{id}/points", getPointsHandler).Methods("GET")

    log.Println("Server's up! Listening on :8080...")
    log.Fatal(http.ListenAndServe(":8080", r))
}

// This function handles receipt submission and returns a receipt ID
func processReceiptHandler(w http.ResponseWriter, r *http.Request) {
    var receipt Receipt
    err := json.NewDecoder(r.Body).Decode(&receipt)
    if err != nil {
        http.Error(w, "Uh oh, invalid request!", http.StatusBadRequest)
        return
    }

    // Generate a random receipt ID
    receiptID := generateReceiptID()
    receiptStorage[receiptID] = receipt

    response := ReceiptIDResponse{ID: receiptID}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// This function handles getting points for a given receipt ID
func getPointsHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    // Check if the receipt exists
    receipt, exists := receiptStorage[id]
    if !exists {
        http.Error(w, "Receipt not found, sorry!", http.StatusNotFound)
        return
    }

    // Calculate the points based on receipt data
    points := calculatePoints(receipt)
    response := PointsResponse{Points: points}

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// This function generates a random receipt ID
func generateReceiptID() string {
    rand.Seed(time.Now().UnixNano())
    letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
    id := make([]rune, 16)
    for i := range id {
        id[i] = letters[rand.Intn(len(letters))]
    }
    return string(id)
}



//////// main logic of point calculartion


func calculatePoints(receipt Receipt) int {
    points := 0

    // 1 point for each alphanumeric character in the retailer name
    re := regexp.MustCompile(`[a-zA-Z0-9]`)
    retailerAlnumChars := re.FindAllString(receipt.Retailer, -1)
    retailerPoints := len(retailerAlnumChars)
    points += retailerPoints
    fmt.Printf("Retailer points: %d (Total: %d)\n", retailerPoints, points)

    // 50 points if the total is a round dollar amount (like 5.00)
    totalFloat, _ := strconv.ParseFloat(receipt.Total, 64)
    if totalFloat == float64(int(totalFloat)) {
        points += 50
        fmt.Printf("Round total points: 50 (Total: %d)\n", points)
    }

    // 25 points if the total is a multiple of 0.25
    if int(totalFloat*100)%25 == 0 {
        points += 25
        fmt.Printf("Multiple of 0.25 points: 25 (Total: %d)\n", points)
    }

    // 5 points for every two items in the receipt
    itemPairs := (len(receipt.Items) / 2) * 5
    points += itemPairs
    fmt.Printf("Item pairs points: %d (Total: %d)\n", itemPairs, points)

    // Extra points if the trimmed item description length is a multiple of 3
    for _, item := range receipt.Items {
        // Properly trim the description using strings.TrimSpace
        trimmedDescription := strings.TrimSpace(item.ShortDescription)
        if len(trimmedDescription)%3 == 0 {
            itemPrice, _ := strconv.ParseFloat(item.Price, 64)
            extraItemPoints := int(math.Ceil(itemPrice * 0.2))
            points += extraItemPoints
            fmt.Printf("Item description '%s' points: %d (Total: %d)\n", trimmedDescription, extraItemPoints, points)
        }
    }

    // 6 points if the purchase date is on an odd day
    purchaseDate, _ := time.Parse("2006-01-02", receipt.PurchaseDate)
    if purchaseDate.Day()%2 != 0 {
        points += 6
        fmt.Printf("Odd day points: 6 (Total: %d)\n", points)
    }

    // 10 points if the purchase time is between 2 PM and 4 PM
    purchaseTime, _ := time.Parse("15:04", receipt.PurchaseTime)
    if purchaseTime.Hour() >= 14 && purchaseTime.Hour() < 16 {
        points += 10
        fmt.Printf("Purchase time points: 10 (Total: %d)\n", points)
    }

    fmt.Printf("Final total points: %d\n", points)
    return points
}