package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Equal[V comparable](t *testing.T, got, expected V) {
	t.Helper()

	if expected != got {
		t.Errorf(`assert.Equal(t, got: %v, expected: %v)`, got, expected)
	}
}

func ErrEqual(t *testing.T, got error, expected string) {
	t.Helper()

	if errors.Is(got, errors.New(expected)) {
		t.Errorf(`assert.ErrEqual(got: %v, expected: %v)`, got, expected)
	}
}

func TestParseFile(t *testing.T) {
	productsByte := getProductFile()

	productsBytesLen := len(productsByte)

	Equal(t, productsBytesLen, 5109)

	products := parseProductsJsonFile(productsByte)

	Equal(t, len(products), 50)
}

func TestProductsMemoryManipulation(t *testing.T) {
	addProductsToMemory()
	decresedStockProduct, err := decreaseProductStock(32, 3)

	Equal(t, err, nil)
	Equal(t, decresedStockProduct.Stock, 8)

	//decrease again should return error
	p, err := decreaseProductStock(32, 9)

	ErrEqual(t, err, "not enough stock")
	Equal(t, p, nil)
}

func TestCreateNewPurchase(t *testing.T) {
	addProductsToMemory()

	//initial stock 22
	newPur := Purchase{
		ProductId: 3,
		Quantity:  2,
		Customer: Customer{
			Name:    "John Doe",
			Address: "Sout Californa 12A",
		},
	}

	_, err := createPurchase(&newPur)

	ErrEqual(t, err, "customer email can't be null!")

	newPur.Customer.Email = "testmail@mail.com"

	createdPurchase, err := createPurchase(&newPur)

	Equal(t, err, nil)
	Equal(t, createdPurchase.Id, 1)
}

func TestCreatePurchaseRoute(t *testing.T) {
	addProductsToMemory()
	r := SetupRouter()
	_ = createPurchaseRouter(r)

	w := httptest.NewRecorder()

	//purchase mock with error no customer email
	purchaseMock := Purchase{
		ProductId: 2,
		Quantity:  3,
		Customer: Customer{
			Name:    "John Wick",
			Address: "Wherever I sleep",
		},
	}

	purchaseMockJson, _ := json.Marshal(purchaseMock)
	req, _ := http.NewRequest("POST", "/purchase", strings.NewReader(string(purchaseMockJson)))

	r.ServeHTTP(w, req)
	Equal(t, http.StatusBadRequest, w.Code)

	//correct purchase test
	w2 := httptest.NewRecorder()

	correctPurchaseMock := Purchase{
		ProductId: 2,
		Quantity:  3,
		Customer: Customer{
			Email:   "johnwick@mail.com",
			Name:    "John Wick",
			Address: "Wherever I sleep",
		},
	}

	//correct purchase mock
	correctPurchaseMockJson, _ := json.Marshal(correctPurchaseMock)
	req2, _ := http.NewRequest("POST", "/purchase", strings.NewReader(string(correctPurchaseMockJson)))

	r.ServeHTTP(w2, req2)
	Equal(t, http.StatusOK, w2.Code)
}
