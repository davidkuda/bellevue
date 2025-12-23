# E2E Test
We are trying to use every programming language in existance, so the end-to-end tests are writen with playwright using python.

## Install
### 1. Install Node js
On ubuntu you can use
```sh
# Download and install nvm (node version manager):
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash

# in lieu of restarting the shell
\. "$HOME/.nvm/nvm.sh"

# Download and install Node.js:
nvm install 24

# Verify the Node.js version:
node -v # Should print "v24.11.1".

# Verify npm version:
npm -v # Should print "11.6.2".
```

See https://playwright.dev/docs/getting-started-vscode#core-features

### 2. Install Python 
```sh
cd e2e_tests
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

`pytest . --browser firefox --headed`