Design Patterns
===============

The collector leverages established software design patterns to ensure a modular, maintainable, and extensible architecture. The key patterns used for handling how and where notifications are sent are the **Strategy** and **Factory** patterns, applied at the **Destination** level.

Strategy Pattern
----------------

The Strategy pattern is used to define a family of dispatching algorithms, encapsulate each one, and make them interchangeable. In our context, each **Destination** type is a concrete implementation of a dispatching strategy.

- **Strategy Interface (`destination.Destination`)**: A common interface, which will be defined in ``pkg/destination/destination.go``, provides the contract for all destination strategies. It will expose a method like `Send(issue issue.Issue) error`.

- **Concrete Strategies (`SlackDestination`, `MsTeamsDestination`, etc.)**: Each specific destination type implements the `Destination` interface. For example, `SlackDestination` encapsulates the full logic for sending a notification to Slack. Internally, it will utilize a `SlackSender` to handle the specific API communication, but the strategy itself is the `SlackDestination`.

- **Context**: The application's routing engine acts as the context. It is configured with a list of `Destination` strategies. When an issue needs to be sent, the router iterates over the relevant destinations and calls the `Send` method on each one, unaware of the specific implementation details (Slack, PagerDuty, etc.).

This pattern allows the routing logic to remain completely independent of how and where notifications are sent.

Factory Pattern
---------------

To create and configure the various `Destination` objects, the Factory pattern is employed. This abstracts the instantiation process, making the system more flexible and easier to manage.

A `DestinationFactory`, which will be located in ``pkg/destination/factory.go``, is responsible for creating `Destination` instances based on the provided configuration.

- **Factory Role**: The factory takes the destination configurations (e.g., from `destinations_config.go`) as input.

- **Object Creation**: Based on the `type` of the destination (e.g., "slack", "pagerduty"), the factory instantiates the correct concrete `Destination` object (e.g., `SlackDestination`). This process includes creating and configuring the underlying `Sender` required by that destination.

- **Decoupling**: This decouples the main application logic from the concrete implementation details of each destination. The application only needs to know how to use the factory to get a list of configured `Destination` strategies, not how to build each one from scratch.

By combining these patterns, the architecture remains clean, decoupled, and easy to extend with new integrations. 