Kafka
=====

Publishes the raw `Issue` object as a JSON message to a Kafka topic.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      kafka:
        - name: "my-kafka-pipeline"
          brokers: "kafka-broker-1:9092,kafka-broker-2:9092"
          topic: "cluster-issues"
          # Optional SASL authentication
          sasl_enabled: true
          username: "my-kafka-user"
          password: "my-kafka-password"
          # scram_sha_256 | scram_sha_512 | plain
          mechanism: "scram_sha_512"

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`brokers`** (string, required)
    A comma-separated list of Kafka broker addresses.

-   **`topic`** (string, required)
    The Kafka topic to publish messages to.

-   **`sasl_enabled`** (boolean, optional)
    Set to `true` to enable SASL authentication. Defaults to `false`.

-   **`username`** (string, optional)
    The SASL username. Required if `sasl_enabled` is `true`.

-   **`password`** (string, optional)
    The SASL password. Required if `sasl_enabled` is `true`.

-   **`mechanism`** (string, optional)
    The SASL mechanism. Supported values are `scram_sha_256`, `scram_sha_512`, and `plain`. Required if `sasl_enabled` is `true`. 