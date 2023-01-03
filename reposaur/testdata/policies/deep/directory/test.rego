package testdata.deep

valid_suffixes := ["-be", "-fe", "-infra"]

violation_bad_name_suffix {
    print("test", count([v | v := endswith(input.name, valid_suffixes[_]); v]) == 0)
    count([v | v := endswith(input.name, valid_suffixes[_]); v]) == 0
}