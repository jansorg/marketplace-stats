# marketplace-stats
This is a tool to create reports for paid plugins on the JetBrains marketplace.

It creates a report for one plugin at a time.

# Building
```bash
go get github.com/jansorg/marketplace-stats
```

# Using
```plain
Usage of ./marketplace-stats:
  -cache-file string
        The file where sales data is cached. Use -fetch to update it. (default "sales.json")
  -fetch
        If online data should be fetched. Needs the --token flag.
  -html string
        The file where the HTML sales report is saved. (default "report.html")
  -pluginID string
        The ID of the plugin, e.g. 12345. (default "13841")
  -token string
        The token to access the API of the JetBrains marketplace.
  -tokenFile string
        Path to a file, which contains the token.
```

# Generating Reports

# Contributing
This tool is made for my own plugins and requirements. I'm not planning to spend time to adjust it to the requirements of others. I'll gladly accept pull requests by others if these don't break existing functionality.

# License
This software is licensed under the GPL, version 3.