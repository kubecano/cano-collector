.. _kafka-destination:

Kafka Destination Configuration
==============================

The Kafka destination allows cano-collector to publish alerts and findings to Kafka topics for downstream processing.

Configuration
-------------

.. code-block:: yaml

    destinations:
      kafka:
        - name: "alerts-topic"
          brokers:
            - "kafka-1:9092"
            - "kafka-2:9092"
            - "kafka-3:9092"
          topic: "cano-events"
          sasl:
            mechanism: "PLAIN"
            username: "cano"
            password: "your-password"
          tls:
            enabled: true
            caFile: "/etc/cano-collector/certs/kafka-ca.crt"
            certFile: "/etc/cano-collector/certs/kafka-client.crt"
            keyFile: "/etc/cano-collector/certs/kafka-client.key"

Parameters
----------

.. list-table::
   :header-rows: 1

   * - Parameter
     - Type
     - Required
     - Description
   * - name
     - string
     - Yes
     - Unique name for this Kafka destination
   * - brokers
     - list
     - Yes
     - List of Kafka broker addresses
   * - topic
     - string
     - Yes
     - Kafka topic to publish messages to
   * - sasl.mechanism
     - string
     - No
     - SASL authentication mechanism (PLAIN, SCRAM-SHA-256, etc.)
   * - sasl.username
     - string
     - No
     - SASL username for authentication
   * - sasl.password
     - string
     - No
     - SASL password for authentication
   * - tls.enabled
     - boolean
     - No
     - Enable TLS encryption
   * - tls.caFile
     - string
     - No
     - Path to CA certificate file
   * - tls.certFile
     - string
     - No
     - Path to client certificate file
   * - tls.keyFile
     - string
     - No
     - Path to client private key file

Message Format
--------------

Messages published to Kafka are in JSON format:

.. code-block:: json

    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Pod CrashLooping",
      "description": "Pod is in CrashLoopBackOff state",
      "severity": "warning",
      "status": "firing",
      "subject": {
        "name": "my-app-pod",
        "type": "pod",
        "namespace": "default"
      },
      "timestamp": "2024-01-15T10:30:00Z",
      "enrichments": [
        {
          "blocks": [
            {
              "type": "markdown",
              "text": "Pod logs:\n```\nError: connection refused\n```"
            }
          ]
        }
      ]
    } 