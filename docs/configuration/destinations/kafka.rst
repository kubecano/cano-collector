.. _kafka-destination:

Kafka
=====

This destination streams structured data to Kafka topics for data pipeline integration.

Configuration
-------------

Basic Configuration
~~~~~~~~~~~~~~~~~~

.. code-block:: yaml

    - name: kafka_destination_name
      type: kafka
      params:
        # Kafka broker URL
        kafka_url: "localhost:9092"
        # Topic to publish messages to
        topic: "robusta-events"

Authenticated Configuration
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: yaml

    - name: kafka_destination_name
      type: kafka
      params:
        # Kafka broker URL
        kafka_url: "localhost:9096"
        # Topic to publish messages to
        topic: "robusta-events"
        # Authentication configuration
        auth:
          sasl_mechanism: SCRAM-SHA-512
          security_protocol: SASL_SSL
          sasl_plain_username: robusta
          sasl_plain_password: password

Parameter Reference
-------------------

``kafka_url``
  *(Required)* The URL of your Kafka broker(s) in the format `host:port`.

``topic``
  *(Required)* The Kafka topic where messages will be published.

``auth``
  *(Optional)* Authentication configuration for Kafka. Supports various mechanisms including SASL and SSL. 