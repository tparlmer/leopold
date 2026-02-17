# Leopold

Go library for OTP-style process supervision.

Leopold provides Erlang/OTP supervision primitives for managing OS processes. It is a library, not a framework — bring your own application, Leopold handles the supervision.

## Status

Early development. Not ready for use.

## Core concepts

- **ChildSpec** — a recipe for starting and restarting a process
- **Supervisor** — watches child processes and applies restart policies
- **Restart intensity** — a circuit breaker that prevents infinite restart loops

## Design

See the [design doc](https://github.com/tparlmer/ai-nexus/blob/main/notes/leopold-design.md) for API surface, type definitions, and implementation phases.

## License

MIT
