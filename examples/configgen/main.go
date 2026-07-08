package main

import (
	"os"

	veloz "github.com/victoragudo/go-veloz"
)

const nginxSrc = `{% for site in sites -%}
server {
    listen {{ site.tls ? 443 : 80 }};
    server_name {{ site.domain }};
{%- if site.tls %}
    ssl_certificate     /etc/ssl/{{ site.domain }}.pem;
    ssl_certificate_key /etc/ssl/{{ site.domain }}.key;
{%- endif %}

    location / {
        proxy_pass http://{{ site.upstream }};
{%- for name, value in site.headers %}
        proxy_set_header {{ name }} "{{ value }}";
{%- endfor %}
    }
}
{% endfor -%}`

func main() {
	engine := veloz.New(veloz.WithAutoescape(false))
	tmpl := engine.MustCompile("nginx", nginxSrc)

	data := map[string]any{
		"sites": []map[string]any{
			{
				"domain":   "shop.example.com",
				"upstream": "shop-backend:8080",
				"tls":      true,
				"headers":  map[string]string{"X-Forwarded-Proto": "https", "X-Request-Id": "$request_id"},
			},
			{
				"domain":   "status.example.com",
				"upstream": "status:3000",
				"tls":      false,
				"headers":  map[string]string{},
			},
		},
	}

	if err := tmpl.RenderTo(os.Stdout, data); err != nil {
		panic(err)
	}
}
