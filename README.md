# Host Event Handler

## VSCode Configuration

### Debug

To be able to debug the please create a `launch.json` file under `.vscode` folder and add the following configuration:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch cli",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "args": [
                // Your github personal access token (PAT)
                // To know how to generate a PAT follow the link https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
                "-github-token=GITHUB_TOKEN",
                // File path to the event payload
                // To know more about the event payload follow the link https://docs.github.com/en/actions/learn-github-actions/contexts#github-context
                "-event-payload=FILE_PATH_TO_EVENT_PAYLOAD"
            ],
            "program": "${workspaceFolder}/cmd/cli/main.go"
        }
    ]
}
```

### License

Please install the extension [licenser](https://marketplace.visualstudio.com/items?itemName=ymotongpoo.licenser) and add the following settings to your workspace.

```json
{
    "licenser.license": "Custom",
    "licenser.author": "Explore.dev, Unipessoal Lda",
    "licenser.customHeader": "Copyright (C) @YEAR@ @AUTHOR@ - All Rights Reserved\nUse of this source code is governed by a license that can be\nfound in the LICENSE file.",
}
```
