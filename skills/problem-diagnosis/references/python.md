# Python Diagnosis Reference

Use this reference for Python exceptions, pytest failures, packaging/import
issues, async problems, or performance symptoms. Diagnose only; do not modify
files.

Collect the traceback, Python version, package manager context, virtualenv
state, test command, relevant environment variables, and the smallest script or
pytest invocation that reproduces the symptom.

Prefer non-mutating commands such as:

```bash
python --version
python -m pytest path/to/test.py -q
python -m pytest path/to/test.py::test_name -q
python -m pip check
python -X dev script.py
```

Record import paths, module shadowing, dependency conflicts, async task
lifecycle, and fixture/setup behavior as hypotheses when relevant.

