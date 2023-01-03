# METADATA
# title: Example Package
# description: Example Package Description
# authors:
#   - John Doe <john.doe@example.com>
# organizations:
#   - Acme
package testdata.metadata

# METADATA
# title: Example Rule
# description: Example Rule Description
# schemas:
#   - input: schema.mock.repo
# authors:
#   - John Doe <john.doe@example.com>
# organizations:
#   - Acme
# custom:
#   tags: [example, guidelines]
#   security-severity: 3.5
violation_example_rule {
    input.id == 123
}