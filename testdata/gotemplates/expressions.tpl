Numbers: {{addi 2 (muli 3 4)}} {{muli (addi 2 3) 4}} {{divf 10.0 4.0}} {{modi 10 3}} {{modf 10.5 3.0}} {{powi 2 8}} {{addi -5 2}}
Strings: {{print "store" "-" "madrid"}}
Compare: {{eq 3 3}} {{ne 3 4}} {{lt 2 5}} {{ge 5 5}} {{lt "abc" "abd"}} {{eqbi true 1}}
Logic: {{and true "yes"}} {{or false "fallback"}} {{not false}}
Ternary: {{if ge .Age 18}}adult{{else}}minor{{end}}
Elvis: {{coalesce .Nickname "guest"}}
Membership: {{hasint 2 1 2 3}} {{not (hasint 9 1 2 3)}} {{instr "a" "cat"}}
Index: {{atidx 1 10 20 30}} {{atidx -1 10 20 30}}
Map literal: {{index (dict "city" "Madrid" "country" "ES") "city"}} {{index (dict "key" "value") "key"}}
Null values: [{{.Missing}}] []
