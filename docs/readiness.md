# Toggler Project Readiness

## Production readiness

The Toggler code base master branch is considered stable
and in production use.

## Architecture

The `toggler` project core design is based on `the clean architecture` principles,
and split across architecture layers.

The folder structure also tries to represent it trough `screaming architecture` elements.

## Testability & Maintainability

The `toggler` project coverage is made with * behaviour driven development* principles,
and as such, the tests aim to justify system behaviour, not implementation.
You probably heard this already many times, but in this case thing about in a way
where you have `Postgres` implementation without a single query or DB table assertion. :)
Purely just behaviour testes.

Why is this good? Through this, I can have as many different implementations,
and share the expectations from separate components such as storage implementation,
and any contributor can jump in and contribute to it, even without deep TDD or BDD practices.
Refactoring the internal implementation of the project components should be easy as making all test green.

[For more about the specifications, read here.](/docs/design/sharedspecs.md)

# Scalability

The Service follows the [12-factor app](https://12factor.net/) principles,
and scale-out via the process model.
The application doesn't use external resource-dependent implementation,
so as long the external resource you use can be scaled out, you will be fine.

If you need to add a new storage implementation,
because you need to use that,
feel free to create an Issue or a PR.

Suppose your company can't use the storage/cache implementations the project currently has. In that case, you should be able to implement your adapter easily with the help of resource specifications in `usecases/specs.Storage`.

For examples, you can check the already existing storage implementations as well.

# [Rollout related Features](/docs/release/README.md)

# [Design](/docs/design/README.md)
