# REKKI API Client (Go)

![badge](https://github.com/rekki/supplier-client-go/workflows/Go/badge.svg)


This is a Go client so restaurant suppliers can retrieve orders posted through Rekki.
It is simply a wrapper around our [HTTP API][0]

## Usage

Simply initialise the Rekki client with your API token, and start using it
straight away!

### Getting new orders

Filters orders created at or after the given UNIX timestamp.

```go
c := rekki.NewClient(nil, "api.rekki.com", "<API token here>") // use the default http.Client created
orders, err := c.ListNotIntegratedOrders(ctx.TODO(), 730512000)
if err != nil {
	fmt.Printf(err.Error())
}

for i, v := range orders {
	fmt.Printf("Order: %+v\n", v)
}
```

### Marking an order as succesfully integrated

Marks a set of orders as integrated in the Rekki platform.

```go
c := rekki.NewClient(nil, "api.rekki.com", "<API token here>")
err := c.SetOrderIntegrated(ctx.TODO(), []string{"order-id-1", "order-id-2"})
if err != nil {
	fmt.Printf(err.Error())
}
```

### Mark an order as failed to integrate

Marks an order as failed to integrate in the Rekki platform.

```go
c := rekki.NewClient(nil, "api.rekki.com", "<API token here>")

// For brevity, we'll omit how to construct an order construct.
e := OrderIntegrationError{Order: order, Attempts: 5, Error: "Invalid product code"}
err := c.SetOrderIntegrated(ctx.TODO(), e)
if err != nil {
	fmt.Printf(err.Error())
}
```

[0]: https://github.com/rekki/supplier-api/blob/master/documentation/order-api.md
