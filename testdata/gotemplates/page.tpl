{{define "header"}}Sales report for {{.Store}}{{end}}{{define "body"}}
{{- range .Products}}
- {{.Name}}: {{.Price}} EUR
{{- end}}
{{end}}{{define "footer"}}This report belongs to {{.Store}}{{end}}BEGIN {{coalesce .Title "Report"}}
{{template "header" .}}
{{template "body" .}}
{{template "footer" .}}
END
