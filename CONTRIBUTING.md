# Contributing to AITriage

First off, thank you for considering contributing to AITriage! It's people like you that make AITriage such a great tool.

## Where do I go from here?

If you've noticed a bug or have a feature request, make one! It's generally best if you get confirmation of your bug or approval for your feature request this way before starting to code.

## Fork & create a branch

If this is something you think you can fix, then fork AITriage and create a branch with a descriptive name.

A good branch name would be (where issue #325 is the ticket you're working on):

```sh
git checkout -b fix/325/add-aws-rules
```

## Local Development

AITriage requires Go 1.25+.

1. Clone the repository:
   ```bash
   git clone https://github.com/cybertortuga/aitriage.git
   cd aitriage
   ```
2. Build the project:
   ```bash
   make build
   ```
3. Run the tests:
   ```bash
   make test
   ```
4. Run the linter:
   ```bash
   make lint
   ```

## Pull Request Process

1. Ensure any install or build dependencies are removed before the end of the layer when doing a build.
2. Update the README.md with details of changes to the interface, this includes new environment variables, exposed ports, useful file locations and container parameters.
3. You may merge the Pull Request in once you have the sign-off of two other developers, or if you do not have permission to do that, you may request the second reviewer to merge it for you.
