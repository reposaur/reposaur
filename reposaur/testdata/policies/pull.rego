package testdata.pull

# METADATA
# schemas:
#   - input: schema.mock.pull
violation_title_malformed {
    not regex.match("(?i)^(\\w+)(\\(.*\\))?(!)?:.*", input.title)
}