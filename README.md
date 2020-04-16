# REKKI API Client (Go)

This is a Go client so restaurant suppliers can retrieve orders posted through Rekki.
It is simply a wrapper around our [HTTP API][0]

## :nut_and_bolt: Usage

Simply initialise the Rekki client with your API token, and start using it
straight away!

### :inbox_tray: `GetOrders`

Filters orders created at or after the given UNIX timestamp.

```go
c := &rekki.Client{BaseURL: "https://api.rekki.com", ApiToken: "<API token here>"}
orders, err := c.GetOrders(123456)

if err != nil {
	fmt.Printf(err.Error())
}

for i, v := range orders {
	fmt.Printf("Order: %+v\n", v)
}
```

[0]: https://github.com/rekki/supplier-api/blob/master/documentation/order-api.md
