#!/bin/bash

jq -r '.default
            | to_entries[]
            | .key + .value.version' \
            Pipfile.lock > requirements.txt
