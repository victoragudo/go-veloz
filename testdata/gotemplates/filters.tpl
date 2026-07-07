Case: {{upper "hello"}} {{lower "HELLO"}} {{capitalize "hello world"}} {{titlecase "point of sale"}}
Trim: [{{trim "  padded  "}}] [{{trimc "xxwordxx" "x"}}]
Size: {{len .Names}} {{len .Names}} {{runelen "azores"}}
Default: {{coalesce .Empty "no name"}} {{coalesce "Oriol" "ignored"}}
Lists: {{join .Names ", "}} | {{join (revstrs .Names) "-"}} | {{index .Names 0}} | {{index .Names (subi (len .Names) 1)}}
Keys: {{join (keysint .Stock) ","}}
Numbers: {{absf -7.5}} {{absf -4.0}} {{round 3.14159 2}} {{round 2.5 0}}
Text: {{replace "cat" "c" "b"}} {{join (split "a;b;c" ";") "+"}}
Reverse text: {{revstr "veloz"}}
Edges of text: {{firstrune "veloz"}}{{lastrune "veloz"}}
