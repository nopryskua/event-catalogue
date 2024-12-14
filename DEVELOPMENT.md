# Troubleshooting

## SSH Agent

In case you have an SSH agent running to manage your SSH keys and clonned the repository using SSH, make sure that environmental variables `SSH_AGENT_SOCK` and `SSH_AUTH_SOCK` point at the corresponding SSH agent socket. In case they are correctly set, running `ssh-add -L` locally should return an non-empty list containing the key of interest. The `Dev Containers` extension is smart enough to support SSH from inside the container when the variables are set correctly.