FROM docker.elastic.co/elasticsearch/elasticsearch:9.0.3

# Install kuromoji plugin for Japanese text analysis
RUN bin/elasticsearch-plugin install analysis-kuromoji

# Set proper permissions
USER elasticsearch
