package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gorilla/mux"
)

// helpr funciton to set up temp router for the purpose of test
func setupRouter() *mux.Router {
    r := mux.NewRouter()
    r.HandleFunc("/receipts/process", processReceiptHandler).Methods("POST")
    r.HandleFunc("/receipts/{id}/points", getPointsHandler).Methods("GET")
    return r
}

func TestProcessReceipt(t *testing.T) {
    router := setupRouter()

    // provided gihub example
    receiptData := Receipt{
        Retailer:     "Target",
        PurchaseDate: "2022-01-01",
        PurchaseTime: "13:01",
        Total:        "35.35",
        Items: []Item{
            {ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
            {ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
            {ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
            {ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
            {ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
        },
    }

    // Encode the receipt into JSON format
    requestBody, err := json.Marshal(receiptData)
    if err != nil {
        t.Fatalf("Failed to encode receipt data: %v", err)
    }

    // Create a new HTTP request to process the receipt
    req, err := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(requestBody))
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }

    // Create a ResponseRecorder to capture the response
    rr := httptest.NewRecorder()

    // Call the handler
    router.ServeHTTP(rr, req)

    // Check the status code is what we expect
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    // Decode the response
    var receiptIDResponse ReceiptIDResponse
    err = json.NewDecoder(rr.Body).Decode(&receiptIDResponse)
    if err != nil {
        t.Fatalf("Failed to decode response: %v", err)
    }

    // Check if receipt was stored
    storedReceipt, exists := receiptStorage[receiptIDResponse.ID]
    if !exists {
        t.Fatalf("Receipt not stored in memory")
    }

    // Verify that the stored receipt matches the input
    if storedReceipt.Retailer != receiptData.Retailer {
        t.Errorf("Stored receipt retailer mismatch: got %v want %v", storedReceipt.Retailer, receiptData.Retailer)
    }
}




func TestGetPointsForMMCornerMarket(t *testing.T) {
    router := setupRouter()

    // Define the receipt data for M&M Corner Market
    receiptData := Receipt{
        Retailer:     "M&M Corner Market",
        PurchaseDate: "2022-03-20",
        PurchaseTime: "14:33",
        Total:        "9.00",
        Items: []Item{
            {ShortDescription: "Gatorade", Price: "2.25"},
            {ShortDescription: "Gatorade", Price: "2.25"},
            {ShortDescription: "Gatorade", Price: "2.25"},
            {ShortDescription: "Gatorade", Price: "2.25"},
        },
    }

    // Step 1: Process the receipt
    requestBody, err := json.Marshal(receiptData)
    if err != nil {
        t.Fatalf("Failed to encode receipt data: %v", err)
    }

    req, err := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(requestBody))
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }

    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    var receiptIDResponse ReceiptIDResponse
    err = json.NewDecoder(rr.Body).Decode(&receiptIDResponse)
    if err != nil {
        t.Fatalf("Failed to decode response: %v", err)
    }

    // Step 2: Get the points for the processed receipt
    req, err = http.NewRequest("GET", "/receipts/"+receiptIDResponse.ID+"/points", nil)
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }

    rr = httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    var pointsResponse PointsResponse
    err = json.NewDecoder(rr.Body).Decode(&pointsResponse)
    if err != nil {
        t.Fatalf("Failed to decode response: %v", err)
    }

    // Verify the points calculation
    expectedPoints := 109 // Based on the breakdown provided for M&M Corner Market
    if pointsResponse.Points != expectedPoints {
        t.Errorf("Points calculation mismatch: got %v want %v", pointsResponse.Points, expectedPoints)
    }
}








//// unit test to get the points
func TestGetPoints(t *testing.T) {
    router := setupRouter()

    // Define a test receipt ID and store it in memory
    receiptID := "test-receipt-id"
    receiptStorage[receiptID] = Receipt{
        Retailer:     "Target",
        PurchaseDate: "2022-01-01",
        PurchaseTime: "13:01",
        Total:        "35.35",
        Items: []Item{
            {ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
            {ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
            {ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
            {ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
            {ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
        },
    }

    // Create a new HTTP request to get points for the stored receipt
    req, err := http.NewRequest("GET", "/receipts/"+receiptID+"/points", nil)
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }

    // Create a ResponseRecorder to capture the response
    rr := httptest.NewRecorder()

    // Call the handler
    router.ServeHTTP(rr, req)

    // Check the status code is what we expect
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    // Decode the response
    var pointsResponse PointsResponse
    err = json.NewDecoder(rr.Body).Decode(&pointsResponse)
    if err != nil {
        t.Fatalf("Failed to decode response: %v", err)
    }

    // Verify the points calculation
    expectedPoints := 28 // Based on the breakdown provided
    if pointsResponse.Points != expectedPoints {
        t.Errorf("Points calculation mismatch: got %v want %v", pointsResponse.Points, expectedPoints)
    }
}
