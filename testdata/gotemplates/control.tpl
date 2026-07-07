{{$label := "Central store"}}Store: {{$label}}
Weather: {{if gt .Temperature 30}}hot{{else if gt .Temperature 15}}warm{{else}}cold{{end}}
{{$total := 0.0}}{{range .Prices}}{{$total = addf $total .}}{{end}}Total price: {{$total}}
Access: {{if and (index .User "active") (ge (index .User "age") 18)}}granted{{else}}denied{{end}}
