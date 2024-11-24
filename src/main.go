package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Product struct {
	Id    int     `json: "id"`
	Name  string  `json: "name"`
	Price float32 `json: "price"`
	Stock int     `json: "stock"`
}

type Customer struct {
	Id      int    `json: "id"`
	Email   string `json: "email"`
	Name    string `json: "name"`
	Address string `json: "address"`
}

type Purchase struct {
	Id        int      `json: "id"`
	ProductId int      `json: "product_id"`
	Quantity  int      `json: "quantity"`
	Customer  Customer `json: "customer"`
}

var productsMem []*Product
var purchasesMem []*Purchase
var customersMem []*Customer

var customerIncrement = 0
var purchaseIncrement = 0

func SetupRouter() *gin.Engine {
	return gin.Default()
}

func getProductsRouter(r *gin.Engine) *gin.Engine {
	r.GET("/products", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, productsMem)
	})
	return r
}

func getPurchaseRouter(r *gin.Engine) *gin.Engine {
	r.GET("/purchase", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, purchasesMem)
	})
	return r
}

func createPurchaseRouter(r *gin.Engine) *gin.Engine {
	r.POST("/purchase", func(ctx *gin.Context) {
		var purchase Purchase
		if err := ctx.BindJSON(&purchase); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		} else {
			pur, err := createPurchase(&purchase)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
			} else {
				ctx.JSON(http.StatusOK, pur)
			}
		}
	})
	return r
}

func main() {
	addProductsToMemory()

	r := SetupRouter()

	_ = getProductsRouter(r)
	_ = getPurchaseRouter(r)
	_ = createPurchaseRouter(r)

	r.Run()
}

func getProductFile() []byte {
	file, err := os.Open("../products.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)

	if err != nil {
		panic(err)
	}

	return bytes
}

func parseProductsJsonFile(bytes []byte) []*Product {
	var products []*Product

	if err := json.Unmarshal(bytes, &products); err != nil {
		panic(err)
	}

	return products
}

func addProductsToMemory() {
	productBytes := getProductFile()
	products := parseProductsJsonFile(productBytes)
	productsMem = products
}

func decreaseProductStock(productId int, stockReduce int) (*Product, error) {
	if len(productsMem) == 0 {
		return nil, errors.New("no products in memory")
	}

	p, err := findProductInMemoryById(productId)

	if err != nil {
		return nil, err
	}

	if p.Stock < stockReduce {
		return nil, errors.New("not enough stock")
	}

	p.Stock -= stockReduce

	return p, nil
}

func findProductInMemoryById(id int) (*Product, error) {
	for _, p := range productsMem {
		if p.Id == id {
			return p, nil
		}
	}

	return nil, fmt.Errorf("no product found for id %d", id)
}

func findCustomerInMemoryByEmail(email string) (*Customer, error) {
	for _, c := range customersMem {
		if c.Email == email {
			return c, nil
		}
	}
	return nil, fmt.Errorf("no customer found for email %s", email)
}

func createCustomer(newCustomer *Customer) (*Customer, error) {
	if newCustomer.Email == "" {
		return nil, errors.New("customer email can't be null")
	}
	if newCustomer.Name == "" {
		return nil, errors.New("customer name can't be null")
	}
	if newCustomer.Address == "" {
		return nil, errors.New("customer address can't be null")
	}
	//create if not exist
	if len(customersMem) == 0 {
		customerIncrement++
		newCustomer.Id = customerIncrement
		_ = append(customersMem, newCustomer)
		return newCustomer, nil
	}

	c, err := findCustomerInMemoryByEmail(newCustomer.Email)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func createPurchase(newPurchase *Purchase) (*Purchase, error) {
	_, err := decreaseProductStock(newPurchase.ProductId, newPurchase.Quantity)
	if err != nil {
		return nil, err
	}

	_, err = createCustomer(&newPurchase.Customer)

	if err != nil {
		return nil, err
	}

	purchaseIncrement++
	newPurchase.Id = purchaseIncrement
	_ = append(purchasesMem, newPurchase)

	return newPurchase, nil
}
