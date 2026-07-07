{% for name in names %}{{ loop.index }}({{ loop.index0 }}) rev {{ loop.revindex }}({{ loop.revindex0 }}) {{ name }}{% if loop.first %} <- first{% endif %}{% if loop.last %} <- last{% endif %} of {{ loop.length }}
{% endfor %}Stock:{% for item, units in stock %} {{ item }}={{ units }}{% endfor %}
Grid: {% for row in grid %}{% for cell in row %}{{ cell }}{% endfor %}|{% endfor %}
Empty case: {% for x in empty_list %}{{ x }}{% else %}the list is empty{% endfor %}
