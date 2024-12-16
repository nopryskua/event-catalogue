# Launching the Catalogue

The project contains `.devcontainer/devcontainer.json`. Therefore, the easiest way to run the project and contribute is by installing the VS code extension `Dev Containers` or launching the project at `Codespaces`.

An alternative way to run the application is by using `docker compose up` locally which will spin up the local dev environment.

# Troubleshooting

## SSH Agent

In case you have an SSH agent running to manage your SSH keys and clonned the repository using SSH, make sure that environmental variables `SSH_AGENT_SOCK` and `SSH_AUTH_SOCK` point at the corresponding SSH agent socket. In case they are correctly set, running `ssh-add -L` locally should return an non-empty list containing the key of interest. The `Dev Containers` extension is smart enough to support SSH from inside the container when the variables are set correctly.

## Dev Container Config

The easiest way to iterate on the configuration of the dev container itself is by using the Dev Container CLI. It may be installed on MacOS by running `brew install devcontainer`. The command to check and spin the environment is `devcontainer up --workspace-folder=.` at the root of the repository. This will show a clean error.