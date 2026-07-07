{% set label = "Central store" %}Store: {{ label }}
Weather: {% if temperature > 30 %}hot{% elseif temperature > 15 %}warm{% else %}cold{% endif %}
{% set total = 0 %}{% for p in prices %}{% set total = total + p %}{% endfor %}Total price: {{ total }}
Access: {% if user.active and user.age >= 18 %}granted{% else %}denied{% endif %}
