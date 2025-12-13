<!--
SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>

SPDX-License-Identifier: Apache-2.0
-->

# newrepo

Take a new / empty repo and prepare it based on a template.

```
Usage: newrepo <name> <description>
```

## flow

1. Clone empty repo
1. Create initial commit
1. Add git remote tpl
1. Pull from tpl
1. Adjust configs and README
1. Commit
