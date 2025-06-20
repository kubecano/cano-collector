Design Patterns
===============

The collector leverages established software design patterns to ensure a modular, maintainable, and extensible architecture. The key patterns used for handling destinations and senders are the **Strategy** and **Factory** patterns.

Strategy Pattern
----------------

The Strategy pattern is used to define a family of algorithms, encapsulate each one, and make them interchangeable. In our context, each **Destination** acts as a concrete implementation of a dispatching strategy.

- **Strategy Interface (`Sender`)**: The `sender.Sender` interface in ``pkg/sender/sender.go`` defines the common contract for all dispatching strategies. It typically includes a `Send(issue issue.Issue)` method.

- **Concrete Strategies (`SlackSender`, `MsTeamsSender`, etc.)**: Each specific sender implements the `Sender` interface. For example, `SlackSender` encapsulates the logic for sending a message to Slack via its API, while `OpsGenieSender` contains the logic for creating an alert in OpsGenie.

- **Context (`Destination`)**: The `Destination` acts as the context that is configured with a concrete Sender strategy. When an issue needs to be sent, the `Destination` delegates the task to its configured `Sender` object.

This pattern allows the core logic to remain independent of how notifications are sent. We can easily add new notification methods (strategies) without changing the components that use them.

Factory Pattern
---------------

To create and configure the various `Destination` objects, the Factory pattern is employed. This abstracts the instantiation process, making the system more flexible and easier to manage.

The factory, located in ``pkg/sender/factory.go``, is responsible for creating `Sender` instances based on the provided configuration.

- **Factory Role**: The factory takes the destination configuration (e.g., from `destinations_config.go`) as input.
- **Object Creation**: Based on the `type` of the destination (e.g., "slack", "pagerduty"), the factory instantiates the corresponding `Sender` object (`SlackSender`, `PagerdutySender`, etc.), injecting the necessary configuration parameters like API keys, webhook URLs, and channel names.
- **Decoupling**: This decouples the main application logic from the concrete implementation details of each sender. The application only needs to know how to use the factory to get a configured `Sender` instance, not how to build one from scratch.

By combining these patterns, the architecture remains clean, decoupled, and easy to extend with new integrations. 