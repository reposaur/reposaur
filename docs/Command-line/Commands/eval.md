#cli 

Evaluate specified policies against input data. The input defaults to standard input. Output is one more SARIF reports, depending if one object or a collection are given as input.

## Usage

```bash
rsr eval [-p POLICY...] [-o OUTPUT] INPUT
```

## Flags

| Name       | Alias | Description                                                                                      |
| ---------- | ----- | ------------------------------------------------------------------------------------------------ |
| `--output` | `-o`  | Output filename. Defaults to standard output (`-`)                                               |
| `--policy` | `-p`  | Path to local policy files, directories or Git repository. Defaults to current working directory |
| `--trace`  | `-t`  | Enable tracing, outputting to standard error the execution trace. Defaults to `false`            |

## Example

```bash
$ gh api /repos/reposaur/reposaur | rsr eval
# { ... }

$ gh api /orgs/reposaur/repos | rsr eval
# { ... }
# { ... }
# { ... }

$ rsr eval repo.json -o reports.json
# {...}

$ rsr eval -p git@github.com:reposaur/policy.git repo.json
```