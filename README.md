# session-manager
Session Manager server reimplementation for PC Creator 2

# Public Instances
There are currently no publicly available instances to use. If you run an instance and want to be featured here, open a pull request or an issue.

# Compatibility
The server does not behave 1 to 1 with the game one. The base behaviour is basically the same (starting sessions, ending sessions), but listening for session count is improved (I couldn't figure out how it works, I think it is a bit bugged so I implemented it differently)

# Building
You must have the [Go programming language](https://go.dev) installed
1. Clone the repository
2. Run the `build.sh` script
3. If prompted, install any needed dependencies and run the script again

It will output a binary named `main` that is the server
