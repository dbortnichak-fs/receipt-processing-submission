package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http"
	"net/http/httptest"
	_ "net/http/httptest"
	"path"
	"strings"
	"testing"
)

func TestPointsForRetailer(t *testing.T) {
	alphaName := "Example Retailer"
	got := pointsForRetailer(alphaName)
	want := int64(15)

	if got != want {
		t.Errorf("alpha only name got %d, wanted %d", got, want)
	}

	nonAlphaName := "*** Example & Retailer ***"
	got = pointsForRetailer(nonAlphaName)
	want = int64(15)

	if got != want {
		t.Errorf("non alpha name got %d, wanted %d", got, want)
	}

	alphaNumericName := "Example Retailer 123"
	got = pointsForRetailer(alphaNumericName)
	want = int64(18)

	if got != want {
		t.Errorf("alphaNumeric name got %d, wanted %d", got, want)
	}

}

func TestPointsForTotal(t *testing.T) {
	totalStr := "35.00"
	got := pointsForTotal(totalStr)
	want := int64(75)

	if got != want {
		t.Errorf("no cents and multiple of .25 got %d, wanted %d", got, want)
	}

	totalStr = "1.25"
	got = pointsForTotal(totalStr)
	want = int64(25)

	if got != want {
		t.Errorf("multiple of .25 got %d, wanted %d", got, want)
	}
}

func TestPointsForItems(t *testing.T) {
	items := []Item{
		{ShortDescription: "Desc 1", Price: "1.0"},
		{ShortDescription: "Desc 2", Price: "2.0"},
		{ShortDescription: "Desc 3", Price: "3.0"},
	}

	got := pointsForItems(items)
	want := int64(5)

	if got != want {
		t.Errorf("5 points for every 2 items got %d, wanted %d", got, want)
	}
}

func TestPointsForItemsDescriptions(t *testing.T) {
	items := []Item{
		{ShortDescription: "Description1", Price: "10.0"},
	}

	got := pointsForItemsDescriptions(items)
	want := int64(2)

	if got != want {
		t.Errorf("item description is a multiple of 3, multiply the price by 0.2 and round up, got %d, wanted %d", got, want)
	}
}

func TestPointsForUsingLLM(t *testing.T) {
	totalStr := "19.90"
	got := pointsForUsingLLM(totalStr, true)
	want := int64(5)

	if got != want {
		t.Errorf("5 points if the total is > 10.00 and you used an LLM, got %d, wanted %d", got, want)
	}
}

func TestPointsForOddDayPurchase(t *testing.T) {
	dateStr := "2022-01-01"
	got := pointsForOddDayPurchase(dateStr)
	want := int64(6)

	if got != want {
		t.Errorf("6 points if the day in the purchase date is odd, got %d, wanted %d", got, want)
	}

	dateStr = "2022-01-02"
	got = pointsForOddDayPurchase(dateStr)
	want = int64(0)

	if got != want {
		t.Errorf("0 points if the day in the purchase date is even, got %d, wanted %d", got, want)
	}
}

func TestPointsForTimeOfPurchase(t *testing.T) {
	timeStr := "15:20"
	got := pointsForTimeOfPurchase(timeStr)
	want := int64(10)

	if got != want {
		t.Errorf("10 points if the time of purchase is after 2:00pm and before 4:00pm, got %d, wanted %d", got, want)
	}

	timeStr = "12:00"
	got = pointsForTimeOfPurchase(timeStr)
	want = int64(0)

	if got != want {
		t.Errorf("0 points if the time of purchase is not in the window, got %d, wanted %d", got, want)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(processPostReceipt)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		fmt.Printf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	return rr
}

func TestReceiptEndpoints(t *testing.T) {

	filePath := path.Join("examples", "simple-receipt.json")
	sampleData, _ := ioutil.ReadFile(filePath)
	reader := strings.NewReader(string(sampleData))
	req, _ := http.NewRequest("POST", "/receipts/process", reader)
	req.Header.Add("Accept", "application/json")
	response := executeRequest(req)
	var m map[string]string
	if response.Code == http.StatusOK {
		json.Unmarshal(response.Body.Bytes(), &m)
		if m["id"] == "" {
			t.Errorf("Expected the 'id' key of the response to be set '%s'", m["id"])
		} else {
			t.Logf("Success the 'id' key of the response is present '%s'", m["id"])
		}
	}
	str := []string{"", "receipts", string(m["id"]), "points", ""}
	uri := strings.Join(str, "/")
	req, _ = http.NewRequest("GET", uri, nil)
	handler := http.HandlerFunc(getReceiptPoints)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code == http.StatusOK {
		json.Unmarshal(rr.Body.Bytes(), &m)
		if m["points"] == "31" {
			fmt.Printf("Success the receipts points endpoint returned '%s'", m["points"])
		}
	} else {
		t.Errorf("Expected the response to contain a receeipt")
	}

}
