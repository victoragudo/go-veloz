{{$n := len .Names}}{{$last := subi $n 1}}{{range $i, $name := .Names}}{{addi $i 1}}({{$i}}) rev {{subi $n $i}}({{subi $last $i}}) {{$name}}{{if eq $i 0}} <- first{{end}}{{if eq $i $last}} <- last{{end}} of {{$n}}
{{end}}Stock:{{range $item, $units := .Stock}} {{$item}}={{$units}}{{end}}
Grid: {{range .Grid}}{{range .}}{{.}}{{end}}|{{end}}
Empty case: {{range .EmptyList}}{{.}}{{else}}the list is empty{{end}}
