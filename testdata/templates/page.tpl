{% extends "layout" %}
{% block header %}Sales report for {{ store }}{% endblock %}
{% block body %}
{%- for p in products %}
- {{ p.Name }}: {{ p.Price }} EUR
{%- endfor %}
{% endblock %}
