package rekki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Price       string `json:"price"`
	ProductCode string `json:"product_code"`
	Quantity    int    `json:"quantity"`
	Units       string `json:"units"`
	Spec        string `json:"spec"`
}

type Order struct {
	CustomerAccountNo string `json:"customer_account_no"`
	ConfirmedAt       string `json:"confirmed_at"`
	ContactInfo       string `json:"contact_info"`
	ContactName       string `json:"contact_name"`
	LocationName      string `json:"location_name"`
	DeliveryAddress   string `json:"delivery_address"`
	DeliveryOn        string `json:"delivery_on"`
	InsertedAtTs      int    `json:"inserted_at_ts"`
	Notes             string `json:"notes"`
	Reference         string `json:"reference"`
	SupplierNotes     string `json:"supplier_notes"`
	Items             []Item `json:"items"`
}

type GetOrdersResponse struct {
	Orders []Order `json:"orders"`
}

type GetOrdersRequestBody struct {
	SinceTS int64 `json:"since"`
}

type Client struct {
	h        *http.Client
	baseURL  string
	apiToken string
}

func NewClient(apiToken string, h *http.Client) *Client {
	if h == nil {
		tr := &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true, // assume input is already compressed
		}
		h = &http.Client{Transport: tr}
	}

	return &Client{baseURL: "api.rekki.com", apiToken: apiToken, h: h}
}

func (c *Client) GetOrders(sinceTS int64) ([]Order, error) {
	reqBody, err := json.Marshal(&GetOrdersRequestBody{SinceTS: sinceTS})
	if err != nil {
		return nil, errors.Wrap(err, "unable to serialise body")
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/catalog/integration/list_orders_by_supplier", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-REKKI-Authorization-Type", "supplier_api_token")
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	res, err := c.h.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Could not read body from /list_orders_by_supplier callout")
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Upstream error response: %s", string(body))
	}

	defer res.Body.Close()

	var or GetOrdersResponse
	err = json.Unmarshal(body, &or)
	if err != nil {
		return nil, err
	}

	return or.Orders, nil
}
