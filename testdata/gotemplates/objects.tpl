Product: {{.Product.Name}} at {{.Product.Price}} EUR
Tags: {{join .Product.Tags "/"}} and the first one is {{index .Product.Tags 0}}
Label: {{.Product.Label}}
Price with discount: {{.Product.WithDiscount 25.0}}
Nested data: {{.Order.customer.email}} bought {{(index .Order.lines 0).qty}} units
