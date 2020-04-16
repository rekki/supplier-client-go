# REKKI API Client (Go)

This is a Go client so restaurant suppliers can retrieve orders posted through Rekki.
It is simply a wrapper around our [HTTP API][0]

## Usage

Simply initialise the Rekki client with your API token, and start using it
straight away!

### `GetOrders`

Filters orders created at or after the given UNIX timestamp.

```go
c := rekki.NewClient("<API token here>", nil) // use the default http.Client created
orders, err := c.GetOrders(123456)

if err != nil {
	fmt.Printf(err.Error())
}

for i, v := range orders {
	fmt.Printf("Order: %+v\n", v)
}
```

[0]: https://github.com/rekki/supplier-api/blob/master/documentation/order-api.md
