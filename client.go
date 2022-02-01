package rekki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type OrderList struct {
	Orders []Order `json:"orders"`
}

type Order struct {
	CustomerAccountNo string      `json:"customer_account_no"`
	ConfirmedAt       *time.Time  `json:"confirmed_at"`
	ContactInfo       string      `json:"contact_info"`
	ContactName       string      `json:"contact_name"`
	LocationName      string      `json:"location_name"`
	DeliveryAddress   string      `json:"delivery_address"`
	PostCode          string      `json:"-"`
	DeliveryOn        simpleDate  `json:"delivery_on"`
	InsertedAtTs      int64       `json:"inserted_at_ts"`
	Notes             string      `json:"notes"`
	Reference         string      `json:"reference"`
	SupplierNotes     string      `json:"supplier_notes"`
	Items             []OrderItem `json:"items"`
}

type OrderItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       string  `json:"price"`
	PriceCents  int64   `json:"price_cents"`
	ProductCode string  `json:"product_code"`
	Quantity    float64 `json:"quantity"`
	Units       string  `json:"units"`
	Spec        string  `json:"spec"`
}

// OrderMap is an alias for a map<string, Order>
type OrderMap map[string]Order

// OrderIntegrationError is a struct for setting errors for failures
type OrderIntegrationError struct {
	Order    Order  `json:"order"`
	Error    string `json:"error"`
	Attempts int    `json:"attempts"`
}

type API interface {
	ListNotIntegratedOrders(ctx context.Context, sinceTS int64) (OrderMap, error)
	SetOrderIntegrated(ctx context.Context, orderReferences []string) error
	SetOrderError(ctx context.Context, e OrderIntegrationError) error
}

type externalSupplierAPI struct {
	listURL          string
	setIntegratedURL string
	setErrorURL      string
	token            string
	client           *http.Client
}

func NewAPI(client *http.Client, host string, token string) (API, error) {
	if client == nil {
		tr := &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true, // assume input is already compressed
		}
		client = &http.Client{Transport: tr}
	}

	api := externalSupplierAPI{token: token, client: client}
	listURL, err := buildURL(host, "api/integration/v1/orders/list_not_integrated")
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse list url")
	}
	api.listURL = listURL

	setURL, err := buildURL(host, "api/integration/v1/orders/set_integrated")
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse set integrated url")
	}
	api.setIntegratedURL = setURL

	errURL, err := buildURL(host, "api/integration/v1/orders/set_error")
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse set error url")
	}
	api.setErrorURL = errURL

	return &api, nil
}

func buildURL(host string, p string) (string, error) {
	h, err := url.Parse(host)
	if err != nil {
		return "", errors.Wrapf(err, "unable to parse HOST:%s", host)
	}

	h.Path = path.Join(h.Path, p)
	return h.String(), nil
}

type GetOrdersRequestBody struct {
	SinceTS int64 `json:"since"`
}

// ListNotIntegratedOrders fetches orders with `{"since":0}`
func (a *externalSupplierAPI) ListNotIntegratedOrders(ctx context.Context, sinceTS int64) (OrderMap, error) {
	reqBody, err := json.Marshal(&GetOrdersRequestBody{SinceTS: sinceTS})
	if err != nil {
		return nil, errors.Wrap(err, "unable to serialise body")
	}

	req, err := newRekkiRequest(ctx, a.listURL, a.token, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create req for fetching orders")
	}

	res, err := a.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "req is failed for fetching orders")
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request is failed %d - %s", res.StatusCode, string(body))
	}

	var r OrderList
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal order")
	}

	// TODO We're iterating here already to assemble the map. Should this be
	// responsability of the consumer and we just return a list?
	orders := make(OrderMap)
	for _, v := range r.Orders {
		orders[v.Reference] = v
	}

	return orders, nil
}

// SetOrderIntegrated marks the order as succesfully integrated in ther Rekki platform.
func (a *externalSupplierAPI) SetOrderIntegrated(ctx context.Context, orderReferences []string) error {
	var p struct {
		Orders []string `json:"orders"`
	}

	p.Orders = orderReferences
	body, err := json.Marshal(&p)
	if err != nil {
		return errors.Wrap(err, "failed to marshal")
	}

	r, err := newRekkiRequest(ctx, a.setIntegratedURL, a.token, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "failed to create req for integrating the order")
	}

	res, err := a.client.Do(r)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("request is failed %d - %s", res.StatusCode, string(b))
	}

	return errors.Wrap(err, "failed req for setting integrated")
}

// SetOrderError marks an orders as failed to integrate. Reasons can vary from technical error
// to uncomplete data.
func (a *externalSupplierAPI) SetOrderError(ctx context.Context, e OrderIntegrationError) error {
	body, err := json.Marshal(&e)
	if err != nil {
		return errors.Wrap(err, "failed to marshal")
	}

	r, err := newRekkiRequest(ctx, a.setErrorURL, a.token, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "failed to create req for setting the failed order integration")
	}

	res, err := a.client.Do(r)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("request is failed %d - %s", res.StatusCode, string(b))
	}

	return errors.Wrap(err, "failed req for setting integrated")
}

func newRekkiRequest(ctx context.Context, url string, token string, r io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, r)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	bearer := "Bearer " + token
	req.Header.Set("Authorization", bearer)
	req.Header.Set("X-REKKI-Authorization-Type", "supplier_api_token")

	return req, nil
}

// This alias has been created to deserialise the date
// since the response from the API is not RFC-valid.
type simpleDate struct {
	time.Time
}

func (ct *simpleDate) MarshalJSON() ([]byte, error) {
	if ct.Time.UnixNano() == 0 {
		return []byte("null"), nil
	}

	ctLayout := "2006-01-02"
	return []byte(fmt.Sprintf("\"%s\"", ct.Time.Format(ctLayout))), nil
}

func (sd *simpleDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		sd.Time = time.Time{}
		return
	}

	ctLayout := "2006-01-02"
	sd.Time, err = time.Parse(ctLayout, s)
	return
}
