package testdata.repo

valid_suffixes := ["-be", "-fe", "-infra"]

# METADATA
# schemas:
#   - input: schema.mock.repo
violation_bad_name_suffix {
    count([v | v := endswith(input.name, valid_suffixes[_]); v]) == 0
}