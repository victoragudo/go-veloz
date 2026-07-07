Numbers: {{ 2 + 3 * 4 }} {{ (2 + 3) * 4 }} {{ 10 / 4 }} {{ 10 % 3 }} {{ 10.5 % 3 }} {{ 2 ** 8 }} {{ -5 + 2 }}
Strings: {{ "store" ~ "-" ~ "madrid" }}
Compare: {{ 3 == 3 }} {{ 3 != 4 }} {{ 2 < 5 }} {{ 5 >= 5 }} {{ "abc" < "abd" }} {{ true == 1 }}
Logic: {{ true and "yes" }} {{ false or "fallback" }} {{ not false }}
Ternary: {{ age >= 18 ? "adult" : "minor" }}
Elvis: {{ nickname ?: "guest" }}
Membership: {{ 2 in [1, 2, 3] }} {{ 9 not in [1, 2, 3] }} {{ "a" in "cat" }}
Index: {{ [10, 20, 30][1] }} {{ [10, 20, 30][-1] }}
Map literal: {{ {city: "Madrid", country: "ES"}.city }} {{ {"key": "value"}["key"] }}
Null values: [{{ missing }}] [{{ null }}]
