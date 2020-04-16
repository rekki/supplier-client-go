package rekki

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestGetCustomer_singleMatching(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := rw.Write([]byte(`
			{
				"orders": [
					{
					"customer_account_no": "R8813", 
					"confirmed_at": "2019-08-12T12:20:10.968294",
					"contact_info": "+447123456789",
					"contact_name": "John Doe",
					"location_name": "Coffee & Cake Cafe",
					"delivery_address": "123 Fake Street, Test",
					"delivery_on": "2019-08-29",
					"inserted_at_ts": 1565458065,
					"notes": "fresh tomato please",
					"reference": "A16915",
					"supplier_notes": "please use back entrance",
					"items": [
						{
						"id": "4fae0a5d3ee045922bae04eb2aee3d52",
						"name": "tomatoes",
						"price": "2.0",
						"product_code": "tm12",
						"quantity": 1,
						"units": "kg",
						"spec": "customer needs vines on tomatoes"
						}
					]
					}
				]
			}
		`))
		if err != nil {
			t.Errorf("got error %+v", err)
		}
	}))

	defer server.Close()

	c := NewClient("api-token", nil)
	c.baseURL = server.URL

	got, err := c.GetOrders(123456)

	want := []Order{
		Order{
			CustomerAccountNo: "R8813",
			ConfirmedAt:       "2019-08-12T12:20:10.968294",
			ContactInfo:       "+447123456789",
			ContactName:       "John Doe",
			LocationName:      "Coffee & Cake Cafe",
			DeliveryAddress:   "123 Fake Street, Test",
			DeliveryOn:        "2019-08-29",
			InsertedAtTs:      1565458065,
			Notes:             "fresh tomato please",
			Reference:         "A16915",
			SupplierNotes:     "please use back entrance",
			Items: []Item{
				Item{
					ID:          "4fae0a5d3ee045922bae04eb2aee3d52",
					Name:        "tomatoes",
					Price:       "2.0",
					ProductCode: "tm12",
					Quantity:    1,
					Units:       "kg",
					Spec:        "customer needs vines on tomatoes",
				},
			},
		},
	}

	if err != nil {
		t.Errorf("Got an error: %s", err.Error())
	}

	if len(got) != len(want) {
		t.Errorf("got length %d, want %d", len(got), len(want))
	}

	if cmp.Equal(got, want, cmpopts.IgnoreFields(Order{}, "Items")) == false {
		t.Errorf("got %+v, want %+v", got, want)
	}

	if got[0].Items[0] != want[0].Items[0] {
		t.Errorf("Item - got item %+v, want %+v", got[0].Items[0], want[0].Items[0])
	}
}
