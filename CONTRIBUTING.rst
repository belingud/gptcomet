============
Contributing
============

Contributions are welcome, and they are greatly appreciated! Every little bit
helps, and credit will always be given.

You can contribute in many ways:

Types of Contributions
----------------------

Report Bugs
~~~~~~~~~~~

Report bugs at https://github.com/belingud/gptcomet/issues

If you are reporting a bug, please include:

* Your operating system name and version.
* Any details about your local setup that might be helpful in troubleshooting.
* Detailed steps to reproduce the bug.

Fix Bugs
~~~~~~~~

Look through the GitHub issues for bugs. Anything tagged with "bug"
and "help wanted" is open to whoever wants to implement a fix for it.

Implement Features
~~~~~~~~~~~~~~~~~~

Look through the GitHub issues for features. Anything tagged with "enhancement"
and "help wanted" is open to whoever wants to implement it.

Write Documentation
~~~~~~~~~~~~~~~~~~~

GPTComet could always use more documentation, whether as part of
the official docs, in docstrings, or even on the web in blog posts, articles,
and such.

Submit Feedback
~~~~~~~~~~~~~~~

The best way to send feedback is to file an issue at
https://github.com/belingud/gptcomet/issues.

If you are proposing a new feature:

* Explain in detail how it would work.
* Keep the scope as narrow as possible, to make it easier to implement.
* Remember that this is a volunteer-driven project, and that contributions
  are welcome :)

Get Started!
------------

Ready to contribute? Here's how to set up `gptcomet` for local
development.

1. Install Dependencies
~~~~~~~~~~~~~~~~~~~~~~~

* Install Go (version 1.20 or higher): https://go.dev/doc/install
* Install Python (version 3.9 or higher): https://www.python.org/downloads/
* Install uv: pip install uv

2. Fork and Clone
~~~~~~~~~~~~~~~~~

| 1. Fork the `gptcomet` repo on GitHub.

| 2. Clone your fork locally:

   .. code-block:: bash

        git clone git@github.com:YOUR_NAME/gptcomet.git
        cd gptcomet

3. Setup Development Environment
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

| 1. Install Go dependencies:

   .. code-block:: bash

        go mod download

| 2. Install Python dependencies:

   .. code-block:: bash

        uv sync

4. Development Workflow
~~~~~~~~~~~~~~~~~~~~~~~

| 1. Create a feature branch:

   .. code-block:: bash

        git checkout -b feature/your-feature-name

| 2. Make your changes following these guidelines:
   - Go code should follow standard Go formatting (run `just format`)
   - Python code should follow PEP 8 guidelines
   - Write tests for new functionality
   - Update documentation as needed

| 3. Run tests:

   .. code-block:: bash

        # Run Go tests
        just test

        # Run Python tests
        just test-py

| 4. Check code quality:

   .. code-block:: bash

        just check

| 5. Commit your changes:

   .. code-block:: bash

        git add .
        git commit -m "Your detailed description of your changes."

| 6. Push your branch:

   .. code-block:: bash

        git push origin feature/your-feature-name

| 7. Create a pull request on GitHub.

Pull Request Guidelines
-----------------------

Before you submit a pull request, check that it meets these guidelines:

1. The pull request should include tests for both Go and Python code.
2. Go code should pass all linters and formatters (run `just check`).
3. Python code should pass all linters and formatters.
4. If the pull request adds functionality, the docs should be updated.
5. New Go code should include proper documentation and examples.
6. Follow the project's coding style and conventions.

Code Style
----------

Go:
- Use gofmt and goimports for formatting
- Follow Effective Go guidelines: https://go.dev/doc/effective_go
- Use descriptive variable names
- Keep functions small and focused

Python:
- Follow PEP 8 style guide
- Use type hints where appropriate
- Keep functions small and focused
- Use descriptive variable names

Testing
-------

We use the following testing frameworks:

- Go: standard testing package
- Python: pytest

All new code should include appropriate tests. Test coverage should be maintained
or improved with each contribution.

Documentation
-------------

Documentation is maintained in the following locations:

- Go: godoc comments in source files
- Python: docstrings in source files
- Project documentation: README.md, CONTRIBUTING.rst

Please update relevant documentation when adding new features or changing
existing functionality.
