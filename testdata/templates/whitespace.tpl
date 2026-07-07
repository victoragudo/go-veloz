a
{%- if true -%}
  b
{%- endif -%}
c
{# a template comment leaves no output #}
d {{- "e" -}} f
