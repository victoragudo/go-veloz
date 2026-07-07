Range up: {{ range(1, 5) | join(",") }}
Range down: {{ range(10, 0, -5) | join(",") }}
Max and min: {{ max(3, 8, 2) }} {{ min([4, 1, 7]) }} {{ max(prices) }}
Length: {{ length(names) }} {{ length("madrid") }}
Filter used as function: {{ upper("veloz") }}
Function used as filter: {{ 5 | max(9) }}
