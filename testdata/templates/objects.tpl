Product: {{ product.Name }} at {{ product.Price }} EUR
Tags: {{ product.Tags | join("/") }} and the first one is {{ product.Tags[0] }}
Label: {{ product.Label() }}
Price with discount: {{ product.WithDiscount(25) }}
Nested data: {{ order.customer.email }} bought {{ order.lines[0].qty }} units
