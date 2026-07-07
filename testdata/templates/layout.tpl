BEGIN {{ title | default("Report") }}
{% block header %}Default header{% endblock %}
{% block body %}{% endblock %}
{% block footer %}{% include "footer" %}{% endblock %}
END
