package rekki

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestFetchOrder(t *testing.T) {
	var fetchOrderTestData = []struct {
		in               string
		expectedApiToken string
		order            map[string]string
	}{
		{
			`{"orders":[{"customer_account_no":"","confirmed_at":null,"contact_info":"+31634000000","contact_name":"Rekki Rekki","location_name":"Reki","delivery_address":"Herengracht 514, Amsterdam, 1017 CC","delivery_on":"2020-03-07","inserted_at_ts":1583516677,"notes":"","reference":"rekki-fetch-1","supplier_notes":"","items":[{"id":"52008c36-5b98-4a4d-bc59-9b02c0f9dbb6","name":"Cucumber","price":"0.00","price_cents":0,"product_code":"Cc","quantity":1,"units":"kg","spec":""}]}]}`,
			"XXXXX-XXXXX-XXXXX",
			map[string]string{
				"rekki-fetch-1": `{"customer_account_no":"","confirmed_at":null,"contact_info":"+31634000000","contact_name":"Rekki Rekki","location_name":"Reki","delivery_address":"Herengracht 514, Amsterdam, 1017 CC","delivery_on":"2020-03-07","inserted_at_ts":1583516677,"notes":"","reference":"rekki-fetch-1","supplier_notes":"","items":[{"id":"52008c36-5b98-4a4d-bc59-9b02c0f9dbb6","name":"Cucumber","price":"0.00","price_cents":0,"product_code":"Cc","quantity":1,"units":"kg","spec":""}]}`,
			},
		},
		{
			`{"orders":[{"customer_account_no":"","confirmed_at":null,"contact_info":"+31634000000","contact_name":"Rekki Rekki","location_name":"Reki","delivery_address":"Herengracht 514, Amsterdam, 1017 CC","delivery_on":"2020-03-07","inserted_at_ts":1583516677,"notes":"","supplier_notes":"","items":[{"id":"52008c36-5b98-4a4d-bc59-9b02c0f9dbb6","name":"Cucumber","price":"0.00","price_cents":0,"product_code":"Cc","quantity":1,"units":"kg","spec":""}]}]}`,
			"XXXXX-XXXXX-XXXXX",
			map[string]string{},
		},
	}

	for _, tt := range fetchOrderTestData {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqAuth := r.Header.Get("Authorization")
			bearerToken := strings.TrimPrefix(reqAuth, "Bearer ")
			if tt.expectedApiToken != bearerToken {
				t.Fatal("invalid bearer token")
			}
			fmt.Fprintln(w, tt.in)
		}))
		defer ts.Close()

		api, err := NewAPI(&http.Client{}, ts.URL, tt.expectedApiToken)
		if err != nil {
			t.Fatal(err)
		}

		orders, err := api.ListNotIntegratedOrders(context.TODO(), 0)
		if err != nil {
			t.Fatal(err)
		}

		for k, v := range tt.order {
			var order Order
			json.Unmarshal([]byte(v), &order)
			if reflect.DeepEqual(order, orders[k]) == false {
				t.Fatalf("invalid order by api \nexpected: %#v \n got: %#v", order, orders[k])
			}
		}
	}
}

func TestSetOrderIntegrated(t *testing.T) {
	var orderTestData = []struct {
		expectedApiToken string
		order            []string
		expectedRequest  string
	}{
		{
			"XXXXX-XXXXX-XXXXX",
			[]string{
				"rekki-set-1",
			},
			`{"orders":["rekki-set-1"]}`,
		},
	}

	for _, tt := range orderTestData {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqAuth := r.Header.Get("Authorization")
			bearerToken := strings.TrimPrefix(reqAuth, "Bearer ")
			if tt.expectedApiToken != bearerToken {
				t.Fatal("invalid bearer token")
			}
			b, _ := ioutil.ReadAll(r.Body)

			if string(b) != tt.expectedRequest {
				t.Fatalf("invalid order list expected: %s - got %s", tt.expectedRequest, string(b))
			}
		}))
		defer ts.Close()

		api, err := NewAPI(&http.Client{}, ts.URL, tt.expectedApiToken)
		if err != nil {
			t.Fatal(err)
		}

		if err := api.SetOrderIntegrated(context.TODO(), tt.order); err != nil {
			t.Fatal(err)
		}
	}
}

func TestConfirmOrders(t *testing.T) {
	var orderTestData = []struct {
		expectedApiToken string
		order            []string
		expectedRequest  string
	}{
		{
			"XXXXX-XXXXX-XXXXX",
			[]string{
				"rekki-set-1",
			},
			`{"orders":["rekki-set-1"]}`,
		},
	}

	for _, tt := range orderTestData {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqAuth := r.Header.Get("Authorization")
			bearerToken := strings.TrimPrefix(reqAuth, "Bearer ")
			if tt.expectedApiToken != bearerToken {
				t.Fatal("invalid bearer token")
			}
			b, _ := ioutil.ReadAll(r.Body)

			if string(b) != tt.expectedRequest {
				t.Fatalf("invalid order list expected: %s - got %s", tt.expectedRequest, string(b))
			}
		}))
		defer ts.Close()

		api, err := NewAPI(&http.Client{}, ts.URL, tt.expectedApiToken)
		if err != nil {
			t.Fatal(err)
		}

		if err := api.ConfirmOrder(context.TODO(), tt.order...); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSetFailedIntegration(t *testing.T) {
	var orderTestData = []struct {
		expectedApiToken string
		order            string
		err              string
		attempts         int
		code             int
	}{
		{
			"XXXXX-XXXXX-XXXXX",
			`{"customer_account_no":"","confirmed_at":null,"contact_info":"+31634000000","contact_name":"Rekki Rekki","location_name":"Reki","delivery_address":"Herengracht 514, Amsterdam, 1017 CC","delivery_on":"2020-03-07","inserted_at_ts":1583516677,"notes":"","reference":"rekki-fetch-1","supplier_notes":"","items":[{"id":"52008c36-5b98-4a4d-bc59-9b02c0f9dbb6","name":"Cucumber","price":"0.00","price_cents":0,"product_code":"Cc","quantity":1,"units":"kg","spec":""}]}`,
			`failed to integrate`,
			5,
			http.StatusOK,
		},
	}

	for _, tt := range orderTestData {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqAuth := r.Header.Get("Authorization")
			bearerToken := strings.TrimPrefix(reqAuth, "Bearer ")
			if tt.expectedApiToken != bearerToken {
				t.Fatal("invalid bearer token")
			}
			b, _ := ioutil.ReadAll(r.Body)
			var oe OrderIntegrationError
			json.Unmarshal(b, &oe)
			if oe.Attempts != tt.attempts {
				t.Fatalf("invalid attempt; expected %d, got: %d", tt.attempts, oe.Attempts)
			}

			if oe.Error != tt.err {
				t.Fatalf("invalid error; expected %s, got: %s", tt.err, oe.Error)
			}

			var expectedOrder Order
			json.Unmarshal([]byte(tt.order), &expectedOrder)
			if reflect.DeepEqual(oe.Order, expectedOrder) == false {
				t.Fatalf("invalid order\n expected %v \n got: %v", expectedOrder, oe.Order)
			}

		}))
		defer ts.Close()

		api, err := NewAPI(&http.Client{}, ts.URL, tt.expectedApiToken)
		if err != nil {
			t.Fatal(err)
		}

		var order Order
		json.Unmarshal([]byte(tt.order), &order)
		e := OrderIntegrationError{Order: order, Attempts: tt.attempts, Error: tt.err}
		if err := api.SetOrderError(context.TODO(), e); err != nil {
			t.Fatal(err)
		}
	}
}

func TestUnmarshalCustomDate(t *testing.T) {
	sd := SimpleDate{}
	sd.UnmarshalJSON([]byte("2020-03-07"))

	if sd.Year() != 2020 {
		t.Errorf("Expected year 2020, got %d", sd.Year())
	}

	if sd.Month() != 03 {
		t.Errorf("Expected month 03, got %d", sd.Month())
	}

	if sd.Day() != 07 {
		t.Errorf("Expected day 07, got %d", sd.Day())
	}
}
