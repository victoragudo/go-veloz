Case: {{ "hello" | upper }} {{ "HELLO" | lower }} {{ "hello world" | capitalize }} {{ "point of sale" | title }}
Trim: [{{ "  padded  " | trim }}] [{{ "xxwordxx" | trim("x") }}]
Size: {{ names | length }} {{ names | count }} {{ "azores" | length }}
Default: {{ empty | default("no name") }} {{ "Oriol" | default("ignored") }}
Lists: {{ names | join(", ") }} | {{ names | reverse | join("-") }} | {{ names | first }} | {{ names | last }}
Keys: {{ stock | keys | join(",") }}
Numbers: {{ (-7.5) | abs }} {{ (-4) | abs }} {{ 3.14159 | round(2) }} {{ 2.5 | round }}
Text: {{ "cat" | replace("c", "b") }} {{ "a;b;c" | split(";") | join("+") }}
Reverse text: {{ "veloz" | reverse }}
Edges of text: {{ "veloz" | first }}{{ "veloz" | last }}
